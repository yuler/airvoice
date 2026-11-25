package cc.yuler.airvoice.services

import android.util.Log
import cc.yuler.airvoice.models.ProtocolMessage
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.TimeoutCancellationException
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeout
import java.util.UUID
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.atomic.AtomicReference

enum class SendTrigger {
    AUTO,
    MANUAL
}

enum class SendResult {
    SUCCESS,
    FAILED,
    CANCELLED
}

class AutoSendController(
    private val textFlow: StateFlow<String>,
    private val connectionManager: ConnectionManager,
    private val isAutoSendEnabled: () -> Boolean = { true },
    private val autoSendDelayMs: () -> Long = { 1500L },
    private val sendTimeoutMs: () -> Long = { 5000L },
    private val onSentAck: (SendResult, String, SendTrigger) -> Unit
) {
    private val scope = CoroutineScope(Dispatchers.Main + SupervisorJob())
    private var debounceJob: kotlinx.coroutines.Job? = null
    private var lastAckedText = ""
    private val pendingAcks = ConcurrentHashMap<String, CompletableDeferred<SendResult>>()
    private val activeMsgId = AtomicReference<String?>(null)

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
                Log.d("AutoSendController", "Collector received msg: type=${msg.type}, id=${msg.id}, ok=${msg.ok}")
                if (msg.type == "ack") {
                    val msgId = msg.id ?: return@collect
                    val result = if (msg.ok == true) SendResult.SUCCESS else SendResult.FAILED
                    pendingAcks.remove(msgId)?.complete(result)
                }
            }
        }

        // Fail pending sends when connection drops
        connectionManager.status
            .onEach { status ->
                if (status is ConnectionStatus.Disconnected || status is ConnectionStatus.Error) {
                    val acksToFail = pendingAcks.keys.toList()
                    if (acksToFail.isNotEmpty()) {
                        activeMsgId.set(null)
                        _inFlight.value = false
                        for (msgId in acksToFail) {
                            pendingAcks.remove(msgId)?.complete(SendResult.FAILED)
                        }
                    }
                }
            }
            .launchIn(scope)
    }

    fun textDidChange(text: String) {
        debounceJob?.cancel()
        val trimmed = text.trim()
        if (trimmed.isEmpty()) {
            resetLastAcked()
            stopCountdown()
            return
        }
        if (!isAutoSendEnabled() || _inFlight.value) {
            stopCountdown()
            return
        }

        val delayDuration = autoSendDelayMs()
        startCountdown()
        debounceJob = scope.launch {
            delay(delayDuration)
            if (!isActive || !isAutoSendEnabled()) return@launch
            _countdownActive.value = false
            scope.launch {
                attemptSend(textFlow.value, SendTrigger.AUTO)
            }
        }
    }

    fun cancelCountdown() {
        stopCountdown()
    }

    fun triggerImmediateSend() {
        if (!isAutoSendEnabled()) return
        val text = textFlow.value
        if (text.isNotEmpty() && text != lastAckedText) {
            scope.launch {
                attemptSend(text, SendTrigger.MANUAL)
            }
        }
    }

    fun clearInFlight() {
        activeMsgId.set(null)
        _inFlight.value = false
        val keys = pendingAcks.keys.toList()
        for (key in keys) {
            pendingAcks.remove(key)?.complete(SendResult.CANCELLED)
        }
    }

    suspend fun attemptSend(text: String, trigger: SendTrigger): Boolean {
        Log.d("AutoSendController", "attemptSend called: textLen=${text.length}, trigger=$trigger, inFlight=${_inFlight.value}")
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
        activeMsgId.set(msgId)
        val deferred = CompletableDeferred<SendResult>()
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
            if (activeMsgId.compareAndSet(msgId, null)) {
                _inFlight.value = false
            }
            onSentAck(SendResult.FAILED, text, trigger)
            return false
        }

        val result = try {
            Log.d("AutoSendController", "Awaiting deferred for msgId=$msgId")
            withTimeout(sendTimeoutMs()) {
                deferred.await()
            }
        } catch (e: TimeoutCancellationException) {
            Log.d("AutoSendController", "Timeout awaiting ack for msgId=$msgId")
            SendResult.FAILED
        } finally {
            pendingAcks.remove(msgId)
            if (activeMsgId.compareAndSet(msgId, null)) {
                _inFlight.value = false
            }
            Log.d("AutoSendController", "Cleaned up msgId=$msgId")
        }

        Log.d("AutoSendController", "attemptSend finished: msgId=$msgId, result=$result")
        when (result) {
            SendResult.SUCCESS -> {
                lastAckedText = text
                onSentAck(SendResult.SUCCESS, text, trigger)
                return true
            }
            SendResult.FAILED -> {
                onSentAck(SendResult.FAILED, text, trigger)
                return false
            }
            SendResult.CANCELLED -> {
                onSentAck(SendResult.CANCELLED, text, trigger)
                return false
            }
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
