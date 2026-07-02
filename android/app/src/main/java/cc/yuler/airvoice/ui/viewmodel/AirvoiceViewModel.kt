package cc.yuler.airvoice.ui.viewmodel

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import cc.yuler.airvoice.models.PairingPayload
import android.util.Log
import cc.yuler.airvoice.services.AutoSendController
import cc.yuler.airvoice.services.ConnectionManager
import cc.yuler.airvoice.services.ConnectionStatus
import cc.yuler.airvoice.services.SendTrigger
import cc.yuler.airvoice.services.StorageManager
import cc.yuler.airvoice.utils.VibratorHelper
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

enum class Screen {
    ONBOARDING,
    SCANNER,
    HOME
}

class AirvoiceViewModel(application: Application) : AndroidViewModel(application) {
    private val storage = StorageManager(application)
    private val client = (application as cc.yuler.airvoice.AirvoiceApplication).okHttpClient
    val connectionManager = ConnectionManager(client)
    private val vibratorHelper = VibratorHelper(application)

    private val _hasSeenOnboarding = MutableStateFlow(false)
    val hasSeenOnboarding: StateFlow<Boolean> = _hasSeenOnboarding.asStateFlow()

    private val _appTheme = MutableStateFlow("light")
    val appTheme: StateFlow<String> = _appTheme.asStateFlow()

    private val _inputText = MutableStateFlow("")
    val inputText: StateFlow<String> = _inputText.asStateFlow()

    private val _toastMessage = MutableStateFlow<String?>(null)
    val toastMessage: StateFlow<String?> = _toastMessage.asStateFlow()

    private val _isToastError = MutableStateFlow(false)
    val isToastError: StateFlow<Boolean> = _isToastError.asStateFlow()

    private val _sendTimedOut = MutableStateFlow(false)
    val sendTimedOut: StateFlow<Boolean> = _sendTimedOut.asStateFlow()

    val autoSendController: AutoSendController

    private var isRetry = false
    private var toastJob: kotlinx.coroutines.Job? = null

    init {
        viewModelScope.launch {
            // Read settings on startup
            _hasSeenOnboarding.value = storage.hasSeenOnboardingFlow.first()
            _appTheme.value = storage.themeFlow.first()
            
            val conn = storage.connectionInfoFlow.first()
            if (!conn.wsUrl.isNullOrEmpty() && !conn.token.isNullOrEmpty()) {
                connectionManager.connect(conn.wsUrl, conn.token)
            }
        }

        autoSendController = AutoSendController(
            textFlow = _inputText,
            connectionManager = connectionManager,
            onSentAck = { success, sentText, trigger ->
                handleSentAck(success, sentText, trigger)
            }
        )
    }

    fun updateInputText(text: String) {
        _inputText.value = text
        autoSendController.textDidChange(text)
    }

    fun triggerImmediateSend() {
        autoSendController.triggerImmediateSend()
    }

    fun completeOnboarding() {
        viewModelScope.launch {
            storage.saveHasSeenOnboarding(true)
            _hasSeenOnboarding.value = true
        }
    }

    fun toggleTheme() {
        viewModelScope.launch {
            val nextTheme = if (_appTheme.value == "light") "dark" else "light"
            _appTheme.value = nextTheme
            storage.saveTheme(nextTheme)
        }
    }

    fun showToast(message: String, isError: Boolean = false) {
        _isToastError.value = isError
        _toastMessage.value = message
        toastJob?.cancel()
        toastJob = viewModelScope.launch {
            delay(2000)
            _toastMessage.value = null
        }
    }

    fun manualSend() {
        val currentText = _inputText.value
        if (currentText.trim().isEmpty()) {
            showToast("请输入文字", isError = true)
            return
        }
        if (connectionManager.status.value !is ConnectionStatus.Connected) {
            showToast("请先连接电脑", isError = true)
            return
        }
        if (autoSendController.inFlight.value) {
            showToast("上一条仍在发送中", isError = true)
            return
        }
        isRetry = false
        _sendTimedOut.value = false
        viewModelScope.launch {
            autoSendController.attemptSend(currentText, SendTrigger.MANUAL)
        }
    }

    fun cancelSend() {
        if (autoSendController.inFlight.value) {
            autoSendController.clearInFlight()
            showToast("已取消发送", isError = false)
        }
    }

    private fun handleSentAck(success: Boolean, sentText: String, trigger: SendTrigger) {
        Log.d("AirvoiceViewModel", "handleSentAck called: success=$success, sentText=\"$sentText\", currentText=\"${_inputText.value}\"")
        if (success) {
            val currentText = _inputText.value
            val remaining = if (currentText.startsWith(sentText)) {
                currentText.removePrefix(sentText)
            } else {
                currentText
            }
            _inputText.value = remaining
            if (remaining.isNotEmpty()) {
                autoSendController.textDidChange(remaining)
            }
            _sendTimedOut.value = false
            isRetry = false
            vibratorHelper.triggerHapticClick()
            showToast("已发送到电脑", isError = false)
        } else {
            val shouldRetry = (trigger == SendTrigger.AUTO) && !isRetry
            Log.d("AirvoiceViewModel", "handleSentAck failure: shouldRetry=$shouldRetry")
            if (shouldRetry && !sentText.trim().isEmpty()) {
                isRetry = true
                viewModelScope.launch {
                    delay(400)
                    if (connectionManager.status.value is ConnectionStatus.Connected && !autoSendController.inFlight.value) {
                        autoSendController.attemptSend(sentText, SendTrigger.AUTO)
                    } else {
                        isRetry = false
                    }
                }
            } else {
                _sendTimedOut.value = true
                isRetry = false
                showToast("发送失败，请检查连接", isError = true)
            }
        }
    }

    fun pairAndConnect(payload: PairingPayload) {
        viewModelScope.launch {
            storage.saveConnection(payload.ws, payload.token)
            autoSendController.resetLastAcked()
            connectionManager.connect(payload.ws, payload.token)
        }
    }

    fun disconnectAndClear() {
        viewModelScope.launch {
            connectionManager.disconnect()
            storage.clearConnection()
            _inputText.value = ""
        }
    }

    override fun onCleared() {
        super.onCleared()
        connectionManager.disconnect()
        autoSendController.cleanup()
    }
}
