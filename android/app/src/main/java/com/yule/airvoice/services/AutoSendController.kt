package com.yule.airvoice.services

import com.yule.airvoice.models.ProtocolMessage
import java.util.UUID
import android.util.Log
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.launch
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.withTimeout
import kotlinx.coroutines.TimeoutCancellationException
import java.util.concurrent.ConcurrentHashMap

enum class SendTrigger {
    AUTO,
    MANUAL
}

class AutoSendController(
    private val textFlow: StateFlow<String>,
    private val connectionManager: ConnectionManager,
    private val onSentAck: (Boolean, String, SendTrigger) -> Unit
) {
    private val scope = CoroutineScope(Dispatchers.Main + SupervisorJob())
    private var debounceJob: kotlinx.coroutines.Job? = null
    private var lastAckedText = ""
    private val pendingAcks = ConcurrentHashMap<String, CompletableDeferred<Boolean>>()

    private val _inFlight = MutableStateFlow(false)
    val inFlight: StateFlow<Boolean> = _inFlight.asStateFlow()

    private val _countdownActive = MutableStateFlow(false)
    val countdownActive: StateFlow<Boolean> = _countdownActive.asStateFlow()

    private val _countdownToken = MutableStateFlow(0)
    val countdownToken: StateFlow<Int> = _countdownToken.asStateFlow()

    // Secondary constructors for backwards compatibility during transition
    constructor(
        textFlow: StateFlow<String>,
        connectionManager: ConnectionManager,
        onSentAck: (Boolean) -> Unit
    ) : this(
        textFlow = textFlow,
        connectionManager = connectionManager,
        onSentAck = { success, _, _ -> onSentAck(success) }
    )

    constructor(
        textFlow: StateFlow<String>,
        connectionManager: ConnectionManager,
        onSentAck: (Boolean, String) -> Unit
    ) : this(
        textFlow = textFlow,
        connectionManager = connectionManager,
        onSentAck = { success, sentText, _ -> onSentAck(success, sentText) }
    )

    init {
        startListening()
    }

    private fun startListening() {
        // Listen to incoming acks
        scope.launch {
            connectionManager.incomingMessages.collect { msg ->
                Log.d("AutoSendController", "Collector received msg: type=${msg.type}, id=${msg.id}, success=${msg.ok}")
                if (msg.type == "ack") {
                    val msgId = msg.id ?: return@collect
                    pendingAcks.remove(msgId)?.complete(msg.ok == true)
                }
            }
        }

        // Reset isSending when connection drops
        connectionManager.status
            .onEach { status ->
                if (status is ConnectionStatus.Disconnected || status is ConnectionStatus.Error) {
                    val acksToFail = pendingAcks.keys.toList()
                    if (acksToFail.isNotEmpty()) {
                        for (msgId in acksToFail) {
                            pendingAcks.remove(msgId)?.complete(false)
                        }
                        _inFlight.value = false
                    }
                }
            }
            .launchIn(scope)
    }

    fun textDidChange(text: String) {
        debounceJob?.cancel()
        val trimmed = text.trim()
        if (trimmed.isEmpty() || _inFlight.value) {
            stopCountdown()
            return
        }

        startCountdown()
        debounceJob = scope.launch {
            delay(1500L)
            _countdownActive.value = false
            attemptSend(text, SendTrigger.AUTO)
        }
    }

    fun triggerImmediateSend() {
        val text = textFlow.value
        if (text.isNotEmpty() && text != lastAckedText) {
            scope.launch {
                attemptSend(text, SendTrigger.MANUAL)
            }
        }
    }

    fun beginSend() {
        stopCountdown()
        _inFlight.value = true
    }

    fun markAcked(content: String) {
        lastAckedText = content
        _inFlight.value = false
    }

    fun clearInFlight() {
        _inFlight.value = false
        val keys = pendingAcks.keys.toList()
        for (key in keys) {
            pendingAcks.remove(key)?.complete(false)
        }
    }

    suspend fun attemptSend(text: String, trigger: SendTrigger): Boolean {
        if (_inFlight.value) return false
        val trimmed = text.trim()
        if (trimmed.isEmpty()) return false
        if (trimmed == lastAckedText.trim()) return false

        beginSend()
        
        val msgId = UUID.randomUUID().toString()
        val deferred = CompletableDeferred<Boolean>()
        pendingAcks[msgId] = deferred
        
        val textMessage = ProtocolMessage(
            type = "text",
            id = msgId,
            content = text,
            ts = System.currentTimeMillis() / 1000
        )
        
        val sent = connectionManager.send(textMessage)
        if (!sent) {
            pendingAcks.remove(msgId)
            _inFlight.value = false
            onSentAck(false, text, trigger)
            return false
        }

        val success = try {
            withTimeout(5000L) {
                deferred.await()
            }
        } catch (e: TimeoutCancellationException) {
            false
        } finally {
            pendingAcks.remove(msgId)
            _inFlight.value = false
        }

        if (success) {
            lastAckedText = text
        }
        onSentAck(success, text, trigger)
        sendPendingText()
        return success
    }

    private suspend fun sendPendingText() {
        val current = textFlow.value
        if (current.isNotEmpty() && current != lastAckedText) {
            attemptSend(current, SendTrigger.AUTO)
        }
    }

    private fun startCountdown() {
        _countdownToken.value += 1
        _countdownActive.value = true
    }

    private fun stopCountdown() {
        debounceJob?.cancel()
        _countdownActive.value = false
    }

    fun resetLastAcked() {
        lastAckedText = ""
    }
    
    fun cleanup() {
        scope.cancel()
    }
}
