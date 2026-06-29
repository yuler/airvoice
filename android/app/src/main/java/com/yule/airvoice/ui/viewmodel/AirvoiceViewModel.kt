package com.yule.airvoice.ui.viewmodel

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.yule.airvoice.models.PairingPayload
import com.yule.airvoice.services.AutoSendController
import com.yule.airvoice.services.ConnectionManager
import com.yule.airvoice.services.ConnectionStatus
import com.yule.airvoice.services.StorageManager
import com.yule.airvoice.utils.VibratorHelper
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import okhttp3.OkHttpClient

enum class Screen {
    ONBOARDING,
    SCANNER,
    HOME
}

class AirvoiceViewModel(application: Application) : AndroidViewModel(application) {
    private val storage = StorageManager(application)
    private val client = (application as com.yule.airvoice.AirvoiceApplication).okHttpClient
    val connectionManager = ConnectionManager(client)
    private val vibratorHelper = VibratorHelper(application)

    private val _currentScreen = MutableStateFlow(Screen.ONBOARDING)
    val currentScreen: StateFlow<Screen> = _currentScreen.asStateFlow()

    private val _inputText = MutableStateFlow("")
    val inputText: StateFlow<String> = _inputText.asStateFlow()

    private val _toastEvents = MutableSharedFlow<String>()
    val toastEvents: SharedFlow<String> = _toastEvents.asSharedFlow()

    private var autoSendController: AutoSendController? = null

    init {
        viewModelScope.launch {
            // Read saved connection details on startup
            val wsUrl = storage.wsUrlFlow.first()
            val token = storage.tokenFlow.first()
            if (!wsUrl.isNullOrEmpty() && !token.isNullOrEmpty() && _currentScreen.value == Screen.ONBOARDING) {
                _currentScreen.value = Screen.HOME
                connectionManager.connect(wsUrl, token)
            }
        }

        autoSendController = AutoSendController(
            textFlow = _inputText,
            connectionManager = connectionManager,
            onSentAck = { success, sentText ->
                if (success) {
                    val current = _inputText.value
                    if (current.startsWith(sentText)) {
                        _inputText.value = current.removePrefix(sentText)
                    }
                    viewModelScope.launch {
                        vibratorHelper.triggerHapticClick()
                        _toastEvents.emit("已发送到电脑")
                    }
                } else {
                    viewModelScope.launch {
                        _toastEvents.emit("发送失败，请检查连接")
                    }
                }
            }
        )
    }

    fun updateInputText(text: String) {
        _inputText.value = text
    }

    fun triggerImmediateSend() {
        autoSendController?.triggerImmediateSend()
    }

    fun navigateTo(screen: Screen) {
        _currentScreen.value = screen
    }

    fun pairAndConnect(payload: PairingPayload) {
        viewModelScope.launch {
            storage.saveConnection(payload.ws, payload.token)
            autoSendController?.resetLastAcked()
            connectionManager.connect(payload.ws, payload.token)
            _currentScreen.value = Screen.HOME
        }
    }

    fun disconnectAndClear() {
        viewModelScope.launch {
            connectionManager.disconnect()
            storage.clearConnection()
            _inputText.value = ""
            _currentScreen.value = Screen.ONBOARDING
        }
    }

    override fun onCleared() {
        super.onCleared()
        connectionManager.disconnect()
        autoSendController?.cleanup()
    }
}
