package com.yule.airvoice.services

import com.yule.airvoice.models.ProtocolMessage
import java.util.UUID
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.flow.debounce
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.launch

class AutoSendController(
    private val textFlow: StateFlow<String>,
    private val connectionManager: ConnectionManager,
    private val onSentAck: (Boolean, String) -> Unit
) {
    private val scope = CoroutineScope(Dispatchers.Main + SupervisorJob())
    private var debounceJob: kotlinx.coroutines.Job? = null
    private var lastAckedText = ""
    private var isSending = false
    private var sendingText = ""
    private var pendingMessageId: String? = null
    private var timeoutJob: kotlinx.coroutines.Job? = null

    init {
        startListening()
    }

    private fun startListening() {
        // Listen to debounce flow
        debounceJob = scope.launch {
            @Suppress("OPT_IN_USAGE")
            textFlow
                .debounce(1500L)
                .collectLatest { text ->
                    if (text.isEmpty()) {
                        lastAckedText = ""
                    } else if (text != lastAckedText) {
                        sendText(text)
                    }
                }
        }

        // Listen to incoming acks
        scope.launch {
            connectionManager.incomingMessages.collect { msg ->
                if (msg.type == "ack" && msg.id == pendingMessageId) {
                    timeoutJob?.cancel()
                    pendingMessageId = null
                    isSending = false
                    val success = msg.ok == true
                    if (success) {
                        lastAckedText = sendingText
                    }
                    onSentAck(success, sendingText)
                    sendPendingText()
                }
            }
        }

        // Reset isSending when connection drops
        connectionManager.status
            .onEach { status ->
                if (status is ConnectionStatus.Disconnected || status is ConnectionStatus.Error) {
                    if (isSending) {
                        timeoutJob?.cancel()
                        pendingMessageId = null
                        isSending = false
                        onSentAck(false, sendingText)
                    }
                }
            }
            .launchIn(scope)
    }

    fun triggerImmediateSend() {
        val text = textFlow.value
        if (text.isNotEmpty() && text != lastAckedText) {
            sendText(text)
        }
    }

    private fun sendText(text: String) {
        if (isSending) return
        isSending = true
        sendingText = text
        
        val msgId = UUID.randomUUID().toString()
        pendingMessageId = msgId
        
        val textMessage = ProtocolMessage(
            type = "text",
            id = msgId,
            content = text,
            ts = System.currentTimeMillis() / 1000
        )
        
        val sent = connectionManager.send(textMessage)
        if (sent) {
            // Start 5 second fallback timeout
            timeoutJob = scope.launch {
                delay(5000L)
                if (isSending && pendingMessageId == msgId) {
                    pendingMessageId = null
                    isSending = false
                    onSentAck(false, sendingText)
                }
            }
        } else {
            isSending = false
            onSentAck(false, text)
        }
    }

    private fun sendPendingText() {
        val current = textFlow.value
        if (current.isNotEmpty() && current != lastAckedText) {
            sendText(current)
        }
    }

    fun resetLastAcked() {
        lastAckedText = ""
    }
    
    fun cleanup() {
        scope.cancel()
    }
}
