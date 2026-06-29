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
    private var sendingText = ""
    private var pendingMessageId: String? = null
    private var timeoutJob: kotlinx.coroutines.Job? = null
    private val sentMessages = mutableMapOf<String, String>()

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
                Log.d("AutoSendController", "Collector received msg: type=${msg.type}, id=${msg.id}, pendingMessageId=$pendingMessageId, success=${msg.ok}")
                if (msg.type == "ack") {
                    val msgId = msg.id ?: return@collect
                    val sentText = sentMessages.remove(msgId)

                    if (msgId == pendingMessageId) {
                        timeoutJob?.cancel()
                        pendingMessageId = null
                        _inFlight.value = false
                    }

                    if (sentText != null) {
                        val success = msg.ok == true
                        if (success) {
                            lastAckedText = sentText
                        }
                        onSentAck(success, sentText, if (sentText == textFlow.value) SendTrigger.MANUAL else SendTrigger.AUTO)
                        sendPendingText()
                    }
                }
            }
        }

        // Reset isSending when connection drops
        connectionManager.status
            .onEach { status ->
                if (status is ConnectionStatus.Disconnected || status is ConnectionStatus.Error) {
                    if (_inFlight.value) {
                        timeoutJob?.cancel()
                        pendingMessageId = null
                        _inFlight.value = false
                        sentMessages.clear()
                        onSentAck(false, sendingText, SendTrigger.AUTO)
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
            attemptSend(text, SendTrigger.MANUAL)
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
        timeoutJob?.cancel()
        pendingMessageId = null
        sentMessages.clear()
    }

    fun attemptSend(text: String, trigger: SendTrigger): Boolean {
        if (_inFlight.value) return false
        val trimmed = text.trim()
        if (trimmed.isEmpty()) return false
        if (trimmed == lastAckedText.trim()) return false

        beginSend()
        sendingText = text
        
        val msgId = UUID.randomUUID().toString()
        pendingMessageId = msgId
        
        val textMessage = ProtocolMessage(
            type = "text",
            id = msgId,
            content = text,
            ts = System.currentTimeMillis() / 1000
        )
        
        sentMessages[msgId] = text
        val sent = connectionManager.send(textMessage)
        if (sent) {
            // Start 5 second fallback timeout, keep in sentMessages for 30s to handle late ACKs
            timeoutJob = scope.launch {
                delay(5000L)
                if (_inFlight.value && pendingMessageId == msgId) {
                    pendingMessageId = null
                    _inFlight.value = false
                    onSentAck(false, sendingText, trigger)
                }
                delay(25000L)
                sentMessages.remove(msgId)
            }
            return true
        } else {
            sentMessages.remove(msgId)
            _inFlight.value = false
            onSentAck(false, text, trigger)
            return false
        }
    }

    private fun sendPendingText() {
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
