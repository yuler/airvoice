package cc.yuler.airvoice.services

import android.util.Log
import cc.yuler.airvoice.models.PairingValidationResult
import cc.yuler.airvoice.models.PairingValidator
import cc.yuler.airvoice.models.ProtocolMessage
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.channels.BufferOverflow
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener

sealed interface ConnectionStatus {
    object Disconnected : ConnectionStatus
    object Connecting : ConnectionStatus
    data class Connected(val host: String) : ConnectionStatus
    data class Error(val message: String) : ConnectionStatus
}

class ConnectionManager(private val client: OkHttpClient) {
    private val lenientJson = Json { ignoreUnknownKeys = true }
    private val _status = MutableStateFlow<ConnectionStatus>(ConnectionStatus.Disconnected)
    val status: StateFlow<ConnectionStatus> = _status

    private val _incomingMessages = MutableSharedFlow<ProtocolMessage>(
        extraBufferCapacity = 64,
        onBufferOverflow = BufferOverflow.DROP_OLDEST
    )
    val incomingMessages: SharedFlow<ProtocolMessage> = _incomingMessages

    private val _authExpiredEvents = MutableSharedFlow<Unit>(
        extraBufferCapacity = 1,
        onBufferOverflow = BufferOverflow.DROP_OLDEST
    )
    val authExpiredEvents: SharedFlow<Unit> = _authExpiredEvents

    @Volatile private var webSocket: WebSocket? = null
    @Volatile private var currentUrl: String? = null
    @Volatile private var currentToken: String? = null
    @Volatile private var currentAllowRemote: Boolean = false
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    @Volatile private var reconnectJob: kotlinx.coroutines.Job? = null
    private var backoffMs = 2000L

    @Synchronized
    fun connect(wsUrl: String, token: String, allowRemote: Boolean = false) {
        val validation = PairingValidator.validate(wsUrl, token)
        if (validation is PairingValidationResult.Invalid) {
            _status.value = ConnectionStatus.Error(PairingValidator.errorMessage(validation.reason))
            return
        }
        val valid = validation as PairingValidationResult.Valid
        if (!valid.isLocal && !allowRemote) {
            _status.value = ConnectionStatus.Error("非局域网地址未经确认")
            return
        }

        val trimmedToken = token.trim()
        val request = try {
            Request.Builder()
                .url(valid.normalizedUrl)
                .header("Authorization", "Bearer $trimmedToken")
                .build()
        } catch (e: IllegalArgumentException) {
            _status.value = ConnectionStatus.Error("无效的服务器地址")
            return
        }

        currentUrl = valid.normalizedUrl
        currentToken = trimmedToken
        currentAllowRemote = allowRemote
        reconnectJob?.cancel()
        webSocket?.close(1000, "Reconnecting")
        webSocket = null

        _status.value = ConnectionStatus.Connecting
        webSocket = client.newWebSocket(request, object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                synchronized(this@ConnectionManager) {
                    if (webSocket !== this@ConnectionManager.webSocket) return
                    backoffMs = 2000L
                }
                val helloMsg = ProtocolMessage(type = "hello", device = "Android Phone", app = "0.1.0")
                webSocket.send(lenientJson.encodeToString(ProtocolMessage.serializer(), helloMsg))
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                if (webSocket !== this@ConnectionManager.webSocket) return
                Log.d("ConnectionManager", "Received WebSocket message: len=${text.length}")
                try {
                    val msg = lenientJson.decodeFromString<ProtocolMessage>(text)
                    if (msg.type == "hello") {
                        _status.value = ConnectionStatus.Connected(msg.host ?: "Computer")
                    }
                    _incomingMessages.tryEmit(msg)
                } catch (e: Exception) {
                    Log.e("ConnectionManager", "Failed to decode message: ${e.javaClass.simpleName}")
                }
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                var authExpired = false
                synchronized(this@ConnectionManager) {
                    if (webSocket !== this@ConnectionManager.webSocket) return
                    if (response?.code == 401) {
                        reconnectJob?.cancel()
                        currentUrl = null
                        currentToken = null
                        currentAllowRemote = false
                        this@ConnectionManager.webSocket = null
                        backoffMs = 2000L
                        _status.value = ConnectionStatus.Disconnected
                        authExpired = true
                    } else {
                        _status.value = ConnectionStatus.Error(t.message ?: "Connection Failure")
                        triggerReconnect()
                    }
                }
                if (authExpired) {
                    _authExpiredEvents.tryEmit(Unit)
                }
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                synchronized(this@ConnectionManager) {
                    if (webSocket !== this@ConnectionManager.webSocket) return
                    _status.value = ConnectionStatus.Disconnected
                    triggerReconnect()
                }
            }
        })
    }

    @Synchronized
    fun disconnect() {
        reconnectJob?.cancel()
        currentUrl = null
        currentToken = null
        currentAllowRemote = false
        webSocket?.close(1000, "User disconnect")
        webSocket = null
        _status.value = ConnectionStatus.Disconnected
    }

    @Synchronized
    fun send(message: ProtocolMessage): Boolean {
        val ws = webSocket ?: return false
        val jsonStr = lenientJson.encodeToString(ProtocolMessage.serializer(), message)
        return ws.send(jsonStr)
    }

    @Synchronized
    private fun triggerReconnect() {
        reconnectJob?.cancel()
        reconnectJob = scope.launch {
            val currentBackoff = synchronized(this@ConnectionManager) {
                val b = backoffMs
                backoffMs = (backoffMs * 2).coerceAtMost(30000L)
                b
            }
            delay(currentBackoff)
            synchronized(this@ConnectionManager) {
                if (!isActive) return@launch
                val url = currentUrl
                val token = currentToken
                val allowRemote = currentAllowRemote
                if (url != null && token != null) {
                    connect(url, token, allowRemote = allowRemote)
                }
            }
        }
    }
}
