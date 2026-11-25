package cc.yuler.airvoice.ui.viewmodel

import android.app.Application
import android.util.Log
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import cc.yuler.airvoice.AirvoiceApplication
import cc.yuler.airvoice.models.AppLanguage
import cc.yuler.airvoice.models.AppSettings
import cc.yuler.airvoice.models.PairingPayload
import cc.yuler.airvoice.models.PairingValidationResult
import cc.yuler.airvoice.models.PairingValidator
import cc.yuler.airvoice.models.ThemeMode
import cc.yuler.airvoice.services.AutoSendController
import cc.yuler.airvoice.services.ConnectionManager
import cc.yuler.airvoice.services.ConnectionStatus
import cc.yuler.airvoice.services.SendResult
import cc.yuler.airvoice.services.SendTrigger
import cc.yuler.airvoice.services.StorageManager
import cc.yuler.airvoice.ui.i18n.AppStrings
import cc.yuler.airvoice.ui.i18n.stringsFor
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

    private val _settings = MutableStateFlow(AppSettings())
    val settings: StateFlow<AppSettings> = _settings.asStateFlow()

    private val _pairedWsUrl = MutableStateFlow<String?>(null)
    val pairedWsUrl: StateFlow<String?> = _pairedWsUrl.asStateFlow()

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
            _settings.value = storage.settingsFlow.first()

            val conn = storage.readConnectionAndMigrateLegacy()
            _pairedWsUrl.value = conn.wsUrl
            if (!conn.wsUrl.isNullOrEmpty() && !conn.token.isNullOrEmpty()) {
                connectionManager.connect(conn.wsUrl, conn.token, allowRemote = conn.allowRemote)
            }
        }

        viewModelScope.launch {
            connectionManager.authExpiredEvents.collect {
                cancelPendingRetry()
                storage.clearConnection()
                _pairedWsUrl.value = null
                showToast(currentStrings().toastAuthExpired, isError = true)
            }
        }

        autoSendController = AutoSendController(
            textFlow = _inputText,
            connectionManager = connectionManager,
            isAutoSendEnabled = { _settings.value.autoSendEnabled },
            autoSendDelayMs = { _settings.value.autoSendDelayMs },
            sendTimeoutMs = { _settings.value.sendTimeoutMs },
            onSentAck = { result, sentText, trigger ->
                handleSentAck(result, sentText, trigger)
            }
        )
    }

    fun currentStrings(): AppStrings = stringsFor(_settings.value.language)

    fun updateInputText(text: String) {
        _inputText.value = text
        autoSendController.textDidChange(text)
    }

    fun triggerImmediateSend() {
        autoSendController.triggerImmediateSend()
    }

    fun setThemeMode(mode: ThemeMode) {
        _settings.value = _settings.value.copy(themeMode = mode)
        viewModelScope.launch { storage.saveThemeMode(mode) }
    }

    fun setLanguage(language: AppLanguage) {
        _settings.value = _settings.value.copy(language = language)
        viewModelScope.launch { storage.saveLanguage(language) }
    }

    fun setAutoSendEnabled(enabled: Boolean) {
        _settings.value = _settings.value.copy(autoSendEnabled = enabled)
        if (!enabled) {
            autoSendController.cancelCountdown()
        }
        viewModelScope.launch { storage.saveAutoSendEnabled(enabled) }
    }

    fun setAutoSendDelayMs(delayMs: Long) {
        val sanitized = AppSettings.sanitizeDelayMs(delayMs)
        _settings.value = _settings.value.copy(autoSendDelayMs = sanitized)
        viewModelScope.launch { storage.saveAutoSendDelayMs(sanitized) }
    }

    fun setHapticFeedbackEnabled(enabled: Boolean) {
        _settings.value = _settings.value.copy(hapticFeedbackEnabled = enabled)
        viewModelScope.launch { storage.saveHapticFeedbackEnabled(enabled) }
    }

    fun setKeepScreenAwake(enabled: Boolean) {
        _settings.value = _settings.value.copy(keepScreenAwake = enabled)
        viewModelScope.launch { storage.saveKeepScreenAwake(enabled) }
    }

    fun setShowInputTips(enabled: Boolean) {
        _settings.value = _settings.value.copy(showInputTips = enabled)
        viewModelScope.launch { storage.saveShowInputTips(enabled) }
    }

    fun setSendTimeoutMs(timeoutMs: Long) {
        val sanitized = AppSettings.sanitizeTimeoutMs(timeoutMs)
        _settings.value = _settings.value.copy(sendTimeoutMs = sanitized)
        viewModelScope.launch { storage.saveSendTimeoutMs(sanitized) }
    }

    fun resetSettings() {
        _settings.value = AppSettings()
        viewModelScope.launch {
            storage.resetSettings()
        }
        showToast(currentStrings().toastSettingsReset, isError = false)
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
        val strings = currentStrings()
        val currentText = _inputText.value
        if (currentText.trim().isEmpty()) {
            showToast(strings.toastEnterText, isError = true)
            return
        }
        if (connectionManager.status.value !is ConnectionStatus.Connected) {
            showToast(strings.toastConnectFirst, isError = true)
            return
        }
        if (autoSendController.inFlight.value) {
            showToast(strings.toastSendInFlight, isError = true)
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
            showToast(currentStrings().toastSendCancelled, isError = false)
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
        val strings = currentStrings()
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
                if (_settings.value.hapticFeedbackEnabled) {
                    vibratorHelper.triggerHapticClick()
                }
                showToast(strings.toastSentToPc, isError = false)
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
                    showToast(strings.toastSendFailed, isError = true)
                }
            }
        }
    }

    fun pairAndConnect(payload: PairingPayload, allowRemote: Boolean = false) {
        val strings = currentStrings()
        val validation = PairingValidator.validate(payload)
        if (validation is PairingValidationResult.Invalid) {
            showToast(PairingValidator.errorMessage(validation.reason, strings), isError = true)
            return
        }
        val valid = validation as PairingValidationResult.Valid
        if (!valid.isLocal && !allowRemote) {
            showToast(strings.errUnconfirmedRemote, isError = true)
            return
        }

        _pairedWsUrl.value = valid.normalizedUrl
        viewModelScope.launch {
            cancelPendingRetry()
            storage.saveConnection(valid.normalizedUrl, payload.token.trim(), allowRemote = allowRemote)
            autoSendController.resetLastAcked()
            connectionManager.connect(valid.normalizedUrl, payload.token.trim(), allowRemote = allowRemote)
        }
    }

    fun disconnectAndClear() {
        _pairedWsUrl.value = null
        _inputText.value = ""
        _sendTimedOut.value = false
        viewModelScope.launch {
            cancelPendingRetry()
            autoSendController.cancelCountdown()
            autoSendController.clearInFlight()
            connectionManager.disconnect()
            storage.clearConnection()
        }
        showToast(currentStrings().toastPairingCleared, isError = false)
    }

    override fun onCleared() {
        super.onCleared()
        cancelPendingRetry()
        connectionManager.disconnect()
        autoSendController.cleanup()
    }
}
