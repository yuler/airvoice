package com.yule.airvoice.services

import com.yule.airvoice.models.ProtocolMessage
import java.util.UUID
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.flow.debounce
import kotlinx.coroutines.launch

class AutoSendController(
    private val textFlow: StateFlow<String>,
    private val connectionManager: ConnectionManager,
    private val onSentAck: (Boolean) -> Unit
) {
    private val scope = CoroutineScope(Dispatchers.Main)
    private var debounceJob: Job? = null
    private var lastAckedText = ""
    private var isSending = false
    private var pendingMessageId: String? = null
    private var timeoutJob: Job? = null

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
                    if (text.isNotEmpty() && text != lastAckedText) {
                        sendText(text)
                    }
                }
        }

        // Listen to incoming acks
        scope.launch {
            connectionManager.incomingMessages.collect { msg ->
                if (msg.type == "ack" && msg.id == pendingMessageId) {
                    timeoutJob?.cancel()
                    isSending = false
                    val success = msg.ok == true
                    if (success) {
                        lastAckedText = textFlow.value
                    }
                    onSentAck(success)
                }
            }
        }
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
                    isSending = false
                    onSentAck(false)
                }
            }
        } else {
            isSending = false
            onSentAck(false)
        }
    }

    fun resetLastAcked() {
        lastAckedText = ""
    }
    
    fun cleanup() {
        debounceJob?.cancel()
        timeoutJob?.cancel()
    }
}
