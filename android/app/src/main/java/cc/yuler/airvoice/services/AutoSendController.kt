package cc.yuler.airvoice.services

import cc.yuler.airvoice.models.ProtocolMessage
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
            kotlinx.coroutines.withContext(kotlinx.coroutines.NonCancellable) {
                attemptSend(text, SendTrigger.AUTO)
            }
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

    fun clearInFlight() {
        _inFlight.value = false
        val keys = pendingAcks.keys.toList()
        for (key in keys) {
            pendingAcks.remove(key)?.complete(false)
        }
    }

    suspend fun attemptSend(text: String, trigger: SendTrigger): Boolean {
        Log.d("AutoSendController", "attemptSend called: text=\"$text\", trigger=$trigger, inFlight=${_inFlight.value}")
        val trimmed = text.trim()
        if (trimmed.isEmpty()) {
            Log.d("AutoSendController", "attemptSend returned early: text is empty")
            return false
        }
        if (trimmed == lastAckedText.trim()) {
            Log.d("AutoSendController", "attemptSend returned early: text matches lastAckedText")
            return false
        }
        if (!_inFlight.compareAndSet(expect = false, update = true)) {
            Log.d("AutoSendController", "attemptSend returned early: inFlight is true")
            return false
        }

        stopCountdown()
        
        val msgId = UUID.randomUUID().toString()
        val deferred = CompletableDeferred<Boolean>()
        pendingAcks[msgId] = deferred
        Log.d("AutoSendController", "Created pending ack for msgId=$msgId")
        
        val textMessage = ProtocolMessage(
            type = "text",
            id = msgId,
            content = text,
            ts = System.currentTimeMillis() / 1000
        )
        
        val sent = connectionManager.send(textMessage)
        if (!sent) {
            Log.d("AutoSendController", "Failed to send message over ConnectionManager: msgId=$msgId")
            pendingAcks.remove(msgId)
            _inFlight.value = false
            onSentAck(false, text, trigger)
            return false
        }

        val success = try {
            Log.d("AutoSendController", "Awaiting deferred for msgId=$msgId")
            withTimeout(5000L) {
                deferred.await()
            }
        } catch (e: TimeoutCancellationException) {
            Log.d("AutoSendController", "Timeout awaiting ack for msgId=$msgId")
            false
        } finally {
            pendingAcks.remove(msgId)
            _inFlight.value = false
            Log.d("AutoSendController", "Cleaned up msgId=$msgId, set inFlight to false")
        }

        Log.d("AutoSendController", "attemptSend finished: success=$success")
        if (success) {
            lastAckedText = text
            onSentAck(true, text, trigger)
            sendPendingText()
        } else {
            onSentAck(false, text, trigger)
        }
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
