package cc.yuler.airvoice.ui.viewmodel

import android.app.Application
import android.util.Log
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import cc.yuler.airvoice.AirvoiceApplication
import cc.yuler.airvoice.models.PairingPayload
import cc.yuler.airvoice.models.PairingValidationResult
import cc.yuler.airvoice.models.PairingValidator
import cc.yuler.airvoice.services.AutoSendController
import cc.yuler.airvoice.services.ConnectionManager
import cc.yuler.airvoice.services.ConnectionStatus
import cc.yuler.airvoice.services.SendResult
import cc.yuler.airvoice.services.SendTrigger
import cc.yuler.airvoice.services.StorageManager
import cc.yuler.airvoice.utils.VibratorHelper
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

enum class Screen {
    SCANNER,
    HOME
}

class AirvoiceViewModel internal constructor(
    application: Application,
    private val storage: StorageManager,
    val connectionManager: ConnectionManager
) : AndroidViewModel(application) {
    constructor(application: Application) : this(
        application = application,
        storage = StorageManager(application),
        connectionManager = ConnectionManager((application as AirvoiceApplication).okHttpClient)
    )

    private val vibratorHelper = VibratorHelper(application)

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
    private var retryJob: Job? = null
    private var toastJob: Job? = null

    init {
        viewModelScope.launch {
            _appTheme.value = storage.themeFlow.first()

            val conn = storage.readConnectionAndMigrateLegacy()
            if (!conn.wsUrl.isNullOrEmpty() && !conn.token.isNullOrEmpty()) {
                connectionManager.connect(conn.wsUrl, conn.token, allowRemote = conn.allowRemote)
            }
        }

        viewModelScope.launch {
            connectionManager.authExpiredEvents.collect {
                cancelPendingRetry()
                storage.clearConnection()
                showToast("配对已失效，请重新扫码", isError = true)
            }
        }

        autoSendController = AutoSendController(
            textFlow = _inputText,
            connectionManager = connectionManager,
            onSentAck = { result, sentText, trigger ->
                handleSentAck(result, sentText, trigger)
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
        cancelPendingRetry()
        _sendTimedOut.value = false
        viewModelScope.launch {
            autoSendController.attemptSend(currentText, SendTrigger.MANUAL)
        }
    }

    fun cancelSend() {
        val hadPendingRetry = retryJob?.isActive == true
        cancelPendingRetry()
        if (autoSendController.inFlight.value || hadPendingRetry) {
            autoSendController.clearInFlight()
            showToast("已取消发送", isError = false)
        }
    }

    private fun cancelPendingRetry() {
        retryJob?.cancel()
        retryJob = null
        isRetry = false
    }

    internal fun handleSentAck(result: SendResult, sentText: String, trigger: SendTrigger) {
        Log.d(
            "AirvoiceViewModel",
            "handleSentAck: result=$result, sentLen=${sentText.length}, currentLen=${_inputText.value.length}, trigger=$trigger"
        )
        when (result) {
            SendResult.SUCCESS -> {
                val currentText = _inputText.value
                val remaining = if (currentText.startsWith(sentText)) {
                    currentText.removePrefix(sentText)
                } else {
                    currentText
                }
                _inputText.value = remaining
                autoSendController.textDidChange(remaining)
                _sendTimedOut.value = false
                cancelPendingRetry()
                vibratorHelper.triggerHapticClick()
                showToast("已发送到电脑", isError = false)
            }
            SendResult.CANCELLED -> {
                cancelPendingRetry()
            }
            SendResult.FAILED -> {
                val shouldRetry = (trigger == SendTrigger.AUTO) && !isRetry
                Log.d("AirvoiceViewModel", "handleSentAck failure: shouldRetry=$shouldRetry")
                if (shouldRetry && sentText.trim().isNotEmpty()) {
                    isRetry = true
                    retryJob?.cancel()
                    retryJob = viewModelScope.launch {
                        delay(400)
                        retryJob = null
                        if (connectionManager.status.value is ConnectionStatus.Connected && !autoSendController.inFlight.value) {
                            autoSendController.attemptSend(sentText, SendTrigger.AUTO)
                        } else {
                            isRetry = false
                        }
                    }
                } else {
                    _sendTimedOut.value = true
                    cancelPendingRetry()
                    showToast("发送失败，请检查连接", isError = true)
                }
            }
        }
    }

    fun pairAndConnect(payload: PairingPayload, allowRemote: Boolean = false) {
        val validation = PairingValidator.validate(payload)
        if (validation is PairingValidationResult.Invalid) {
            showToast(PairingValidator.errorMessage(validation.reason), isError = true)
            return
        }
        val valid = validation as PairingValidationResult.Valid
        if (!valid.isLocal && !allowRemote) {
            showToast("非局域网地址未经确认", isError = true)
            return
        }

        viewModelScope.launch {
            cancelPendingRetry()
            storage.saveConnection(valid.normalizedUrl, payload.token.trim(), allowRemote = allowRemote)
            autoSendController.resetLastAcked()
            connectionManager.connect(valid.normalizedUrl, payload.token.trim(), allowRemote = allowRemote)
        }
    }

    fun disconnectAndClear() {
        viewModelScope.launch {
            cancelPendingRetry()
            connectionManager.disconnect()
            storage.clearConnection()
            _inputText.value = ""
        }
    }

    override fun onCleared() {
        super.onCleared()
        cancelPendingRetry()
        connectionManager.disconnect()
        autoSendController.cleanup()
    }
}
