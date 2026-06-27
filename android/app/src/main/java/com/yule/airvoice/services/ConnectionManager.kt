package com.yule.airvoice.services

import com.yule.airvoice.models.ProtocolMessage
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
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
    private val _status = MutableStateFlow<ConnectionStatus>(ConnectionStatus.Disconnected)
    val status: StateFlow<ConnectionStatus> = _status

    private val _incomingMessages = MutableSharedFlow<ProtocolMessage>()
    val incomingMessages: SharedFlow<ProtocolMessage> = _incomingMessages

    private var webSocket: WebSocket? = null
    private var currentUrl: String? = null
    private var currentToken: String? = null
    private val scope = CoroutineScope(Dispatchers.IO)
    private var reconnectJob: kotlinx.coroutines.Job? = null
    private var backoffMs = 2000L

    fun connect(wsUrl: String, token: String) {
        currentUrl = wsUrl
        currentToken = token
        reconnectJob?.cancel()
        
        val requestUrl = "$wsUrl?token=$token"
        val request = Request.Builder().url(requestUrl).build()
        
        _status.value = ConnectionStatus.Connecting
        webSocket = client.newWebSocket(request, object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                backoffMs = 2000L
                // Send hello handshake
                val helloMsg = ProtocolMessage(type = "hello", device = "Android Phone", app = "0.1.0")
                webSocket.send(Json.encodeToString(ProtocolMessage.serializer(), helloMsg))
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                try {
                    val msg = Json.decodeFromString<ProtocolMessage>(text)
                    scope.launch {
                        if (msg.type == "hello") {
                            _status.value = ConnectionStatus.Connected(msg.host ?: "Computer")
                        }
                        _incomingMessages.emit(msg)
                    }
                } catch (e: Exception) {
                    e.printStackTrace()
                }
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                _status.value = ConnectionStatus.Error(t.message ?: "Connection Failure")
                triggerReconnect()
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                _status.value = ConnectionStatus.Disconnected
            }
        })
    }

    fun disconnect() {
        reconnectJob?.cancel()
        webSocket?.close(1000, "User disconnect")
        webSocket = null
        _status.value = ConnectionStatus.Disconnected
    }

    fun send(message: ProtocolMessage): Boolean {
        val ws = webSocket ?: return false
        val jsonStr = Json.encodeToString(ProtocolMessage.serializer(), message)
        return ws.send(jsonStr)
    }

    private fun triggerReconnect() {
        reconnectJob?.cancel()
        reconnectJob = scope.launch {
            delay(backoffMs)
            backoffMs = (backoffMs * 2).coerceAtMost(30000L)
            val url = currentUrl
            val token = currentToken
            if (url != null && token != null) {
                connect(url, token)
            }
        }
    }
}
