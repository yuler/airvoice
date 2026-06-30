# Android UI/UE Alignment & PR Review Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align the Android native client's UI/UE interactions and screen navigation flow with the iOS version, and resolve the outstanding PR review comments on PR #2.

**Architecture:** We will share a single `OkHttpClient` via the Application class, optimize DataStore operations by combining connection state flows, utilize a central `lenientJson` instance, override the system theme using a custom Compose `CompositionLocal`, and build parity components (custom Slide-In Toast, Countdown Progress Bar, and Input Method Tips Card).

**Tech Stack:** Kotlin, Jetpack Compose, OkHttp, Kotlinx Serialization, Android DataStore Preferences.

---

### Task 1: Developer Tools & Environment Configuration

**Files:**
- Modify: `scripts/java-env.sh`
- Modify: `mise.toml`

- [ ] **Step 1: Update java-env.sh to support Android Studio JBR fallback on Linux**
  Add detection logic for `/opt/android-studio/jbr` and `~/android-studio/jbr` to [java-env.sh](file:///home/yule/Sides/airvoice/scripts/java-env.sh).
  
  Replace the Linux section (lines 22-31):
  ```bash
  # Linux: resolve symlink chain
  if [ -z "$JAVA_HOME" ] || [ ! -x "$JAVA_HOME/bin/java" ]; then
    if [ -x /usr/bin/java ]; then
      java_bin="$(readlink -f /usr/bin/java 2>/dev/null || echo /usr/bin/java)"
      candidate="${java_bin%/bin/java}"
      if [ -x "$candidate/bin/java" ]; then
        JAVA_HOME="$candidate"
      fi
    fi
  fi
  ```
  with:
  ```bash
  # Linux: resolve symlink chain
  if [ -z "$JAVA_HOME" ] || [ ! -x "$JAVA_HOME/bin/java" ]; then
    if [ -x /usr/bin/java ]; then
      java_bin="$(readlink -f /usr/bin/java 2>/dev/null || echo /usr/bin/java)"
      candidate="${java_bin%/bin/java}"
      if [ -x "$candidate/bin/java" ]; then
        JAVA_HOME="$candidate"
      fi
    fi
  fi

  # Android Studio JBR fallback
  if [ -z "$JAVA_HOME" ] || [ ! -x "$JAVA_HOME/bin/java" ]; then
    if [ -d "/opt/android-studio/jbr" ] && [ -x "/opt/android-studio/jbr/bin/java" ]; then
      JAVA_HOME="/opt/android-studio/jbr"
    elif [ -d "$HOME/android-studio/jbr" ] && [ -x "$HOME/android-studio/jbr/bin/java" ]; then
      JAVA_HOME="$HOME/android-studio/jbr"
    fi
  fi
  ```

- [ ] **Step 2: Update mise.toml with cross-platform local IP detection**
  Update the `android:serve` task in [mise.toml](file:///home/yule/Sides/airvoice/mise.toml) line 96:
  
  Target Content:
  ```toml
  run = "bash -c 'APK=android/app/build/outputs/apk/debug/app-debug.apk; if [ ! -f \"$APK\" ]; then echo \"APK not found. Run: mise android:build first\"; exit 1; fi; IP=$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || echo localhost); echo \"\"; echo \"  Download on your phone:  http://${IP}:8888/app-debug.apk\"; echo \"\"; cd android/app/build/outputs/apk/debug && python3 -m http.server 8888 --bind 0.0.0.0'"
  ```
  Replacement Content:
  ```toml
  run = "bash -c 'APK=android/app/build/outputs/apk/debug/app-debug.apk; if [ ! -f \"$APK\" ]; then echo \"APK not found. Run: mise android:build first\"; exit 1; fi; IP=$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || hostname -I 2>/dev/null | awk \"{print \\$1}\" || ip route get 1 2>/dev/null | awk \"{print \\$(NF-2);exit}\" || echo localhost); echo \"\"; echo \"  Download on your phone:  http://${IP}:8888/app-debug.apk\"; echo \"\"; cd android/app/build/outputs/apk/debug && python3 -m http.server 8888 --bind 0.0.0.0'"
  ```

- [ ] **Step 3: Run Android tests to verify compilation and environment setup**
  Run: `mise run android:test`
  Expected: All existing tests PASS.

- [ ] **Step 4: Commit environment updates**
  ```bash
  git add scripts/java-env.sh mise.toml
  git commit -m "chore: support Linux JBR fallback and cross-platform IP detection"
  ```

---

### Task 2: Share OkHttpClient via Application Class

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/AirvoiceApplication.kt`
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/viewmodel/AirvoiceViewModel.kt`

- [ ] **Step 1: Expose OkHttpClient lazily in AirvoiceApplication**
  Modify [AirvoiceApplication.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/AirvoiceApplication.kt):
  
  Target Content:
  ```kotlin
  class AirvoiceApplication : Application()
  ```
  Replacement Content:
  ```kotlin
  import okhttp3.OkHttpClient

  class AirvoiceApplication : Application() {
      val okHttpClient: OkHttpClient by lazy {
          OkHttpClient()
      }
  }
  ```

- [ ] **Step 2: Inject shared client in AirvoiceViewModel**
  Modify [AirvoiceViewModel.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/viewmodel/AirvoiceViewModel.kt) to retrieve this client:
  
  Target Content:
  ```kotlin
      private val storage = StorageManager(application)
      private val client = OkHttpClient()
      val connectionManager = ConnectionManager(client)
  ```
  Replacement Content:
  ```kotlin
      private val storage = StorageManager(application)
      private val client = (application as com.yule.airvoice.AirvoiceApplication).okHttpClient
      val connectionManager = ConnectionManager(client)
  ```

- [ ] **Step 3: Run Android tests to verify client injection**
  Run: `mise run android:test`
  Expected: PASS.

- [ ] **Step 4: Commit Application changes**
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/AirvoiceApplication.kt android/app/src/main/java/com/yule/airvoice/ui/viewmodel/AirvoiceViewModel.kt
  git commit -m "perf: share a single OkHttpClient instance via AirvoiceApplication"
  ```

---

### Task 3: Combine DataStore Connection Flow & Persist Onboarding/Theme

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/services/StorageManager.kt`
- Modify: `android/app/src/test/java/com/yule/airvoice/services/StorageManagerTest.kt`

- [ ] **Step 1: Update StorageManager with consolidated flows, theme and onboarding states**
  Modify [StorageManager.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/services/StorageManager.kt):
  
  Replace the entire class with:
  ```kotlin
  package com.yule.airvoice.services

  import android.content.Context
  import androidx.datastore.preferences.core.booleanPreferencesKey
  import androidx.datastore.preferences.core.edit
  import androidx.datastore.preferences.core.emptyPreferences
  import androidx.datastore.preferences.core.stringPreferencesKey
  import androidx.datastore.preferences.preferencesDataStore
  import java.io.IOException
  import kotlinx.coroutines.flow.Flow
  import kotlinx.coroutines.flow.catch
  import kotlinx.coroutines.flow.map

  private val Context.dataStore by preferencesDataStore(name = "airvoice_prefs")

  data class ConnectionInfo(val wsUrl: String?, val token: String?)

  class StorageManager(private val context: Context) {
      companion object {
          private val KEY_WS = stringPreferencesKey("ws_url")
          private val KEY_TOKEN = stringPreferencesKey("token")
          private val KEY_THEME = stringPreferencesKey("app_theme")
          private val KEY_HAS_SEEN_ONBOARDING = booleanPreferencesKey("has_seen_onboarding")
      }

      val connectionInfoFlow: Flow<ConnectionInfo> = context.dataStore.data
          .catch { exception ->
              if (exception is IOException) {
                  emit(emptyPreferences())
              } else {
                  throw exception
              }
          }
          .map { prefs ->
              ConnectionInfo(prefs[KEY_WS], prefs[KEY_TOKEN])
          }

      val themeFlow: Flow<String> = context.dataStore.data
          .catch { exception ->
              if (exception is IOException) {
                  emit(emptyPreferences())
              } else {
                  throw exception
              }
          }
          .map { prefs -> prefs[KEY_THEME] ?: "light" }

      val hasSeenOnboardingFlow: Flow<Boolean> = context.dataStore.data
          .catch { exception ->
              if (exception is IOException) {
                  emit(emptyPreferences())
              } else {
                  throw exception
              }
          }
          .map { prefs -> prefs[KEY_HAS_SEEN_ONBOARDING] ?: false }

      suspend fun saveConnection(wsUrl: String, token: String) {
          context.dataStore.edit { prefs ->
              prefs[KEY_WS] = wsUrl
              prefs[KEY_TOKEN] = token
          }
      }

      suspend fun clearConnection() {
          context.dataStore.edit { prefs ->
              prefs.remove(KEY_WS)
              prefs.remove(KEY_TOKEN)
          }
      }

      suspend fun saveTheme(theme: String) {
          context.dataStore.edit { prefs ->
              prefs[KEY_THEME] = theme
          }
      }

      suspend fun saveHasSeenOnboarding(completed: Boolean) {
          context.dataStore.edit { prefs ->
              prefs[KEY_HAS_SEEN_ONBOARDING] = completed
          }
      }
  }
  ```

- [ ] **Step 2: Update StorageManagerTest**
  Modify [StorageManagerTest.kt](file:///home/yule/Sides/airvoice/android/app/src/test/java/com/yule/airvoice/services/StorageManagerTest.kt) to test the combined `connectionInfoFlow`, `themeFlow`, and `hasSeenOnboardingFlow`.
  
  Replace the entire class with:
  ```kotlin
  package com.yule.airvoice.services

  import android.content.Context
  import androidx.test.core.app.ApplicationProvider
  import kotlinx.coroutines.flow.first
  import kotlinx.coroutines.runBlocking
  import org.junit.Assert.assertEquals
  import org.junit.Assert.assertFalse
  import org.junit.Assert.assertNull
  import org.junit.Assert.assertTrue
  import org.junit.Before
  import org.junit.Test
  import org.junit.runner.RunWith
  import org.robolectric.RobolectricTestRunner

  @RunWith(RobolectricTestRunner::class)
  class StorageManagerTest {
      private lateinit var context: Context
      private lateinit var storageManager: StorageManager

      @Before
      fun setUp() {
          context = ApplicationProvider.getApplicationContext()
          storageManager = StorageManager(context)
      }

      @Test
      fun testSaveAndClearConnection() = runBlocking {
          var info = storageManager.connectionInfoFlow.first()
          assertNull(info.wsUrl)
          assertNull(info.token)

          storageManager.saveConnection("ws://test", "tok")
          info = storageManager.connectionInfoFlow.first()
          assertEquals("ws://test", info.wsUrl)
          assertEquals("tok", info.token)

          storageManager.clearConnection()
          info = storageManager.connectionInfoFlow.first()
          assertNull(info.wsUrl)
          assertNull(info.token)
      }

      @Test
      fun testSaveTheme() = runBlocking {
          assertEquals("light", storageManager.themeFlow.first())
          storageManager.saveTheme("dark")
          assertEquals("dark", storageManager.themeFlow.first())
      }

      @Test
      fun testSaveOnboarding() = runBlocking {
          assertFalse(storageManager.hasSeenOnboardingFlow.first())
          storageManager.saveHasSeenOnboarding(true)
          assertTrue(storageManager.hasSeenOnboardingFlow.first())
      }
  }
  ```

- [ ] **Step 3: Run StorageManager unit tests**
  Run: `mise run android:test --tests "com.yule.airvoice.services.StorageManagerTest"`
  Expected: PASS.

- [ ] **Step 4: Commit StorageManager updates**
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/services/StorageManager.kt android/app/src/test/java/com/yule/airvoice/services/StorageManagerTest.kt
  git commit -m "feat: combine DataStore flows and persist onboarding/theme configurations"
  ```

---

### Task 4: Utilize lenientJson in ConnectionManager

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/services/ConnectionManager.kt`

- [ ] **Step 1: Modify ConnectionManager to use lenientJson in all serializations**
  Modify [ConnectionManager.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/services/ConnectionManager.kt):
  
  In `onOpen` listener (line 80):
  Target Content:
  ```kotlin
  webSocket.send(Json.encodeToString(ProtocolMessage.serializer(), helloMsg))
  ```
  Replacement Content:
  ```kotlin
  webSocket.send(lenientJson.encodeToString(ProtocolMessage.serializer(), helloMsg))
  ```
  
  In `send` method (line 124):
  Target Content:
  ```kotlin
  val jsonStr = Json.encodeToString(ProtocolMessage.serializer(), message)
  ```
  Replacement Content:
  ```kotlin
  val jsonStr = lenientJson.encodeToString(ProtocolMessage.serializer(), message)
  ```

- [ ] **Step 2: Run ConnectionManager tests**
  Run: `mise run android:test --tests "com.yule.airvoice.services.ConnectionManagerTest"`
  Expected: PASS.

- [ ] **Step 3: Commit ConnectionManager changes**
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/services/ConnectionManager.kt
  git commit -m "refactor: utilize lenientJson for idiomatic serialization in ConnectionManager"
  ```

---

### Task 5: Refactor AutoSendController for iOS Alignment

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/services/AutoSendController.kt`
- Modify: `android/app/src/test/java/com/yule/airvoice/services/AutoSendControllerTest.kt`

- [ ] **Step 1: Rewrite AutoSendController to align with iOS states and triggers**
  Modify [AutoSendController.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/services/AutoSendController.kt):
  
  Replace the entire class with:
  ```kotlin
  package com.yule.airvoice.services

  import com.yule.airvoice.models.ProtocolMessage
  import java.util.UUID
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
                  if (msg.type == "ack" && msg.id == pendingMessageId) {
                      timeoutJob?.cancel()
                      pendingMessageId = null
                      _inFlight.value = false
                      val success = msg.ok == true
                      if (success) {
                          lastAckedText = sendingText
                      }
                      onSentAck(success, sendingText, if (sendingText == textFlow.value) SendTrigger.MANUAL else SendTrigger.AUTO)
                      sendPendingText()
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
          
          val sent = connectionManager.send(textMessage)
          if (sent) {
              // Start 5 second fallback timeout
              timeoutJob = scope.launch {
                  delay(5000L)
                  if (_inFlight.value && pendingMessageId == msgId) {
                      pendingMessageId = null
                      _inFlight.value = false
                      onSentAck(false, sendingText, trigger)
                  }
              }
              return true
          } else {
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
  ```

- [ ] **Step 2: Update AutoSendControllerTest**
  Modify [AutoSendControllerTest.kt](file:///home/yule/Sides/airvoice/android/app/src/test/java/com/yule/airvoice/services/AutoSendControllerTest.kt) to match the new constructor / interfaces:
  
  Replace the entire class with:
  ```kotlin
  package com.yule.airvoice.services

  import com.yule.airvoice.models.ProtocolMessage
  import kotlinx.coroutines.flow.MutableStateFlow
  import kotlinx.coroutines.flow.MutableSharedFlow
  import kotlinx.coroutines.runBlocking
  import org.junit.Assert.assertEquals
  import org.junit.Assert.assertFalse
  import org.junit.Assert.assertTrue
  import org.junit.Before
  import org.junit.Test
  import org.mockito.Mockito.mock
  import org.mockito.Mockito.`when`
  import org.mockito.Mockito.any

  class AutoSendControllerTest {
      private lateinit var connectionManager: ConnectionManager
      private lateinit var textFlow: MutableStateFlow<String>
      private lateinit var incomingMessages: MutableSharedFlow<ProtocolMessage>

      @Before
      fun setUp() {
          connectionManager = mock(ConnectionManager::class.java)
          incomingMessages = MutableSharedFlow()
          `when`(connectionManager.incomingMessages).thenReturn(incomingMessages)
          `when`(connectionManager.status).thenReturn(MutableStateFlow(ConnectionStatus.Disconnected))
          textFlow = MutableStateFlow("")
      }

      @Test
      fun testCountdownStateTransitions() = runBlocking {
          var ackSuccess = false
          var ackTrigger: SendTrigger? = null
          val controller = AutoSendController(textFlow, connectionManager) { success, _, trigger ->
              ackSuccess = success
              ackTrigger = trigger
          }

          assertFalse(controller.countdownActive.value)
          assertEquals(0, controller.countdownToken.value)

          controller.textDidChange("hello")
          assertTrue(controller.countdownActive.value)
          assertEquals(1, controller.countdownToken.value)

          controller.textDidChange("hello world")
          assertTrue(controller.countdownActive.value)
          assertEquals(2, controller.countdownToken.value)
      }
  }
  ```

- [ ] **Step 3: Run AutoSendController tests**
  Run: `mise run android:test --tests "com.yule.airvoice.services.AutoSendControllerTest"`
  Expected: PASS.

- [ ] **Step 4: Commit AutoSendController changes**
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/services/AutoSendController.kt android/app/src/test/java/com/yule/airvoice/services/AutoSendControllerTest.kt
  git commit -m "feat: rewrite AutoSendController to track countdown active, token, and inFlight states"
  ```

---

### Task 6: Refactor AirvoiceViewModel with Parities and Auto-Retry

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/viewmodel/AirvoiceViewModel.kt`

- [ ] **Step 1: Implement AppTheme and Onboarding persistence and manual triggers in AirvoiceViewModel**
  Modify [AirvoiceViewModel.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/viewmodel/AirvoiceViewModel.kt):
  
  Replace the entire class with:
  ```kotlin
  package com.yule.airvoice.ui.viewmodel

  import android.app.Application
  import androidx.lifecycle.AndroidViewModel
  import androidx.lifecycle.viewModelScope
  import com.yule.airvoice.models.PairingPayload
  import com.yule.airvoice.services.AutoSendController
  import com.yule.airvoice.services.ConnectionManager
  import com.yule.airvoice.services.ConnectionStatus
  import com.yule.airvoice.services.SendTrigger
  import com.yule.airvoice.services.StorageManager
  import com.yule.airvoice.utils.VibratorHelper
  import kotlinx.coroutines.delay
  import kotlinx.coroutines.flow.MutableStateFlow
  import kotlinx.coroutines.flow.StateFlow
  import kotlinx.coroutines.flow.asStateFlow
  import kotlinx.coroutines.flow.first
  import kotlinx.coroutines.launch

  class AirvoiceViewModel(application: Application) : AndroidViewModel(application) {
      private val storage = StorageManager(application)
      private val client = (application as com.yule.airvoice.AirvoiceApplication).okHttpClient
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

      var autoSendController: AutoSendController? = null
          private set

      private var lastSentContent: String? = null
      private var lastSentTrigger: SendTrigger? = null
      private var isRetry = false

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
          autoSendController?.textDidChange(text)
      }

      fun triggerImmediateSend() {
          autoSendController?.triggerImmediateSend()
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
          viewModelScope.launch {
              delay(2000)
              if (_toastMessage.value == message) {
                  _toastMessage.value = null
              }
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
          if (autoSendController?.inFlight?.value == true) {
              showToast("上一条仍在发送中", isError = true)
              return
          }
          lastSentContent = currentText
          lastSentTrigger = SendTrigger.MANUAL
          isRetry = false
          _sendTimedOut.value = false
          autoSendController?.attemptSend(currentText, SendTrigger.MANUAL)
      }

      fun cancelSend() {
          if (autoSendController?.inFlight?.value == true) {
              autoSendController?.clearInFlight()
              showToast("已取消发送", isError = false)
          }
      }

      private fun handleSentAck(success: Boolean, sentText: String, trigger: SendTrigger) {
          if (success) {
              _inputText.value = "" // Align with iOS: clear editor completely on success
              _sendTimedOut.value = false
              vibratorHelper.triggerHapticClick()
              showToast("已发送到电脑", isError = false)
          } else {
              val shouldRetry = (trigger == SendTrigger.AUTO) && !isRetry
              if (shouldRetry && !sentText.trim().isEmpty()) {
                  isRetry = true
                  viewModelScope.launch {
                      delay(400)
                      if (connectionManager.status.value is ConnectionStatus.Connected && autoSendController?.inFlight?.value == false) {
                          autoSendController?.attemptSend(sentText, SendTrigger.AUTO)
                      }
                  }
              } else {
                  _sendTimedOut.value = true
                  showToast("发送失败，请检查连接", isError = true)
              }
          }
      }

      fun pairAndConnect(payload: PairingPayload) {
          viewModelScope.launch {
              storage.saveConnection(payload.ws, payload.token)
              autoSendController?.resetLastAcked()
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
          autoSendController?.cleanup()
      }
  }
  ```

- [ ] **Step 2: Verify Android code compiles**
  Run: `mise run android:test`
  Expected: PASS.

- [ ] **Step 3: Commit AirvoiceViewModel changes**
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/ui/viewmodel/AirvoiceViewModel.kt
  git commit -m "feat: implement manual theme toggle, onboarding persistence, and ACK retry/clear parities in ViewModel"
  ```

---

### Task 7: Theme LocalOverride & Custom Slide-In Toast

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/theme/AppTheme.kt`
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/screens/MainScreen.kt`

- [ ] **Step 1: Declare LocalIsDarkTheme and color tokens in AppTheme.kt**
  Modify [AppTheme.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/theme/AppTheme.kt):
  
  Replace the entire file with:
  ```kotlin
  package com.yule.airvoice.ui.theme

  import androidx.compose.runtime.Composable
  import androidx.compose.runtime.staticCompositionLocalOf
  import androidx.compose.ui.graphics.Color

  val LocalIsDarkTheme = staticCompositionLocalOf { false }

  object AppColors {
      val lightBackground = Color(0xFFF5F5F7)
      val darkBackground = Color(0xFF000000)
      val lightSecondaryBackground = Color(0xFFFFFFFF)
      val darkSecondaryBackground = Color(0xFF0D0E15)
      val lightBorder = Color(0xFFE5E5EA)
      val darkBorder = Color(0xFF2E2E2E)
      val lightPrimaryText = Color(0xFF1C1C1E)
      val darkPrimaryText = Color(0xFFFFFFFF)
      val lightSecondaryText = Color(0xFF6C6C70)
      val darkSecondaryText = Color(0xFF8E8E93)
      val accent = Color(0xFF006EFE)
      val lightChipBackground = Color(0x0F000000) // Match iOS 6%
      val darkChipBackground = Color(0x1AFFFFFF)  // Match iOS 10%
      val lightSendButtonBackground = Color(0x0D000000) // Match iOS 5%
      val darkSendButtonBackground = Color(0x1FFFFFFF)  // Match iOS 12%
      val lightPlaceholderText = Color(0xFFAEAEB2)
      val darkPlaceholderText = Color(0xFF8E8E93)
      val statusBarConnected = Color(0xFF00AC3A)
      val statusBarConnecting = Color(0xFFFFAE00)
      val statusBarError = Color(0xFFE2162A)
      val statusBarDisconnected = Color(0xFF8F8F8F)
      
      // Toast colors
      val lightToastBackground = Color(0xEB1C1C1E) // Match iOS 92%
      val darkToastBackground = Color(0xF21F2030)  // Match iOS 95%
      
      // Tips background
      val lightTipsBackground = Color(0xFFEFEFF4)
      val darkTipsBackground = Color(0x0FFFFFFF) // Match iOS 6%
      
      // Countdown bar background
      val lightCountdownBar = Color(0x66FFAE00) // Match iOS 40%
      val darkCountdownBar = Color(0x73FFAE00)  // Match iOS 45%
  }

  @Composable
  fun isDarkTheme(): Boolean = LocalIsDarkTheme.current

  @Composable
  fun backgroundColor(): Color =
      if (isDarkTheme()) AppColors.darkBackground else AppColors.lightBackground

  @Composable
  fun secondaryBackgroundColor(): Color =
      if (isDarkTheme()) AppColors.darkSecondaryBackground else AppColors.lightSecondaryBackground

  @Composable
  fun borderColor(): Color =
      if (isDarkTheme()) AppColors.darkBorder else AppColors.lightBorder

  @Composable
  fun primaryTextColor(): Color =
      if (isDarkTheme()) AppColors.darkPrimaryText else AppColors.lightPrimaryText

  @Composable
  fun secondaryTextColor(): Color =
      if (isDarkTheme()) AppColors.darkSecondaryText else AppColors.lightSecondaryText

  @Composable
  fun chipBackgroundColor(): Color =
      if (isDarkTheme()) AppColors.darkChipBackground else AppColors.lightChipBackground

  @Composable
  fun sendButtonBackgroundColor(): Color =
      if (isDarkTheme()) AppColors.darkSendButtonBackground else AppColors.lightSendButtonBackground

  @Composable
  fun placeholderTextColor(): Color =
      if (isDarkTheme()) AppColors.darkPlaceholderText else AppColors.lightPlaceholderText

  @Composable
  fun toastBackgroundColor(): Color =
      if (isDarkTheme()) AppColors.darkToastBackground else AppColors.lightToastBackground

  @Composable
  fun tipsBackgroundColor(): Color =
      if (isDarkTheme()) AppColors.darkTipsBackground else AppColors.lightTipsBackground

  @Composable
  fun countdownBarColor(): Color =
      if (isDarkTheme()) AppColors.darkCountdownBar else AppColors.lightCountdownBar
  ```

- [ ] **Step 2: Re-architect MainScreen with local theme injection and slide-in custom Toast**
  Modify [MainScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/screens/MainScreen.kt):
  
  Replace the entire file with:
  ```kotlin
  package com.yule.airvoice.ui.screens

  import androidx.compose.animation.AnimatedVisibility
  import androidx.compose.animation.fadeIn
  import androidx.compose.animation.fadeOut
  import androidx.compose.animation.slideInVertically
  import androidx.compose.animation.slideOutVertically
  import androidx.compose.foundation.background
  import androidx.compose.foundation.border
  import androidx.compose.foundation.layout.*
  import androidx.compose.foundation.shape.CircleShape
  import androidx.compose.foundation.shape.RoundedCornerShape
  import androidx.compose.material.icons.Icons
  import androidx.compose.material.icons.filled.CheckCircle
  import androidx.compose.material.icons.filled.Warning
  import androidx.compose.material3.*
  import androidx.compose.runtime.*
  import androidx.compose.ui.Alignment
  import androidx.compose.ui.Modifier
  import androidx.compose.ui.draw.clip
  import androidx.compose.ui.graphics.Color
  import androidx.compose.ui.text.style.TextOverflow
  import androidx.compose.ui.unit.dp
  import androidx.compose.ui.unit.sp
  import com.yule.airvoice.ui.theme.LocalIsDarkTheme
  import com.yule.airvoice.ui.theme.toastBackgroundColor
  import com.yule.airvoice.ui.viewmodel.AirvoiceViewModel

  enum class Screen {
      ONBOARDING,
      SCANNER,
      HOME
  }

  @Composable
  fun MainScreen(viewModel: AirvoiceViewModel) {
      val hasSeenOnboarding by viewModel.hasSeenOnboarding.collectAsState()
      val appTheme by viewModel.appTheme.collectAsState()
      val toastMessage by viewModel.toastMessage.collectAsState()
      val isToastError by viewModel.isToastError.collectAsState()

      var currentScreen by remember { mutableStateOf(Screen.HOME) }

      LaunchedEffect(hasSeenOnboarding) {
          currentScreen = if (hasSeenOnboarding) Screen.HOME else Screen.ONBOARDING
      }

      val isDark = appTheme == "dark"

      CompositionLocalProvider(LocalIsDarkTheme provides isDark) {
          Box(modifier = Modifier.fillMaxSize()) {
              when (currentScreen) {
                  Screen.ONBOARDING -> OnboardingScreen(
                      onStartScanning = {
                          viewModel.completeOnboarding()
                      },
                      onToggleTheme = {
                          viewModel.toggleTheme()
                      }
                  )
                  Screen.SCANNER -> QRScannerScreen(
                      onQrCodeScanned = { payload ->
                          viewModel.pairAndConnect(payload)
                          currentScreen = Screen.HOME
                      },
                      onCancel = {
                          currentScreen = Screen.HOME
                      }
                  )
                  Screen.HOME -> HomeScreen(
                      viewModel = viewModel,
                      onScanQr = {
                          currentScreen = Screen.SCANNER
                      }
                  )
              }

              // Custom Toast Overlay (Parity with iOS Utilities/Toast.swift)
              Box(
                  modifier = Modifier
                      .fillMaxSize()
                      .padding(bottom = 50.dp),
                  contentAlignment = Alignment.BottomCenter
              ) {
                  AnimatedVisibility(
                      visible = toastMessage != null,
                      enter = slideInVertically(initialOffsetY = { it / 2 }) + fadeIn(),
                      exit = slideOutVertically(targetOffsetY = { it / 2 }) + fadeOut()
                  ) {
                      toastMessage?.let { msg ->
                          ToastView(message = msg, isError = isToastError)
                      }
                  }
              }
          }
      }
  }

  @Composable
  fun ToastView(message: String, isError: Boolean) {
      Row(
          modifier = Modifier
              .wrapContentSize()
              .clip(RoundedCornerShape(22.dp))
              .background(toastBackgroundColor())
              .border(1.dp, Color.White.copy(alpha = 0.1f), RoundedCornerShape(22.dp))
              .padding(horizontal = 16.dp, vertical = 12.dp),
          verticalAlignment = Alignment.CenterVertically,
          horizontalArrangement = Arrangement.spacedBy(12.dp)
      ) {
          Icon(
              imageVector = if (isError) Icons.Default.Warning else Icons.Default.CheckCircle,
              contentDescription = null,
              tint = if (isError) Color.Red else Color.Green,
              modifier = Modifier.size(16.dp)
          )

          Text(
              text = message,
              fontSize = 14.sp,
              color = Color.White,
              maxLines = 2,
              overflow = TextOverflow.Ellipsis
          )
      }
  }
  ```

- [ ] **Step 3: Run compiler checks**
  Run: `mise run android:test`
  Expected: PASS.

- [ ] **Step 4: Commit UI foundation updates**
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/ui/theme/AppTheme.kt android/app/src/main/java/com/yule/airvoice/ui/screens/MainScreen.kt
  git commit -m "feat: implement custom CompositionLocal override theme and custom Slide-In Toast UI"
  ```

---

### Task 8: HomeScreen Refactoring & iOS Visual Parities

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt`
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/screens/QRScannerScreen.kt`
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/screens/OnboardingScreen.kt`

- [ ] **Step 1: Redesign HomeScreen with header parities, AutoSendCountdownBar, and InputMethodTipsView**
  Modify [HomeScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt):
  
  Replace the entire file with:
  ```kotlin
  package com.yule.airvoice.ui.screens

  import android.content.Intent
  import android.net.Uri
  import androidx.compose.animation.core.Animatable
  import androidx.compose.animation.core.LinearEasing
  import androidx.compose.animation.core.RepeatMode
  import androidx.compose.animation.core.animateFloat
  import androidx.compose.animation.core.infiniteRepeatable
  import androidx.compose.animation.core.rememberInfiniteTransition
  import androidx.compose.animation.core.tween
  import androidx.compose.foundation.background
  import androidx.compose.foundation.border
  import androidx.compose.foundation.layout.*
  import androidx.compose.foundation.shape.CircleShape
  import androidx.compose.foundation.shape.RoundedCornerShape
  import androidx.compose.material.icons.Icons
  import androidx.compose.material.icons.automirrored.filled.Send
  import androidx.compose.material.icons.filled.Info
  import androidx.compose.material3.*
  import androidx.compose.runtime.*
  import androidx.compose.ui.Alignment
  import androidx.compose.ui.Modifier
  import androidx.compose.ui.draw.clip
  import androidx.compose.ui.draw.scale
  import androidx.compose.ui.focus.FocusRequester
  import androidx.compose.ui.focus.focusRequester
  import androidx.compose.ui.graphics.Color
  import androidx.compose.ui.graphics.vector.ImageVector
  import androidx.compose.ui.platform.LocalContext
  import androidx.compose.ui.platform.LocalDensity
  import androidx.compose.ui.text.font.FontWeight
  import androidx.compose.ui.unit.dp
  import androidx.compose.ui.unit.sp
  import com.yule.airvoice.services.ConnectionStatus
  import com.yule.airvoice.ui.theme.*
  import com.yule.airvoice.ui.viewmodel.AirvoiceViewModel

  @Composable
  fun HomeScreen(
      viewModel: AirvoiceViewModel,
      onScanQr: () -> Unit = {}
  ) {
      val text by viewModel.inputText.collectAsState()
      val status by viewModel.connectionManager.status.collectAsState()
      val countdownActive by viewModel.autoSendController?.countdownActive?.collectAsState() ?: remember { mutableStateOf(false) }
      val countdownToken by viewModel.autoSendController?.countdownToken?.collectAsState() ?: remember { mutableStateOf(0) }
      val inFlight by viewModel.autoSendController?.inFlight?.collectAsState() ?: remember { mutableStateOf(false) }
      val sendTimedOut by viewModel.sendTimedOut.collectAsState()
      val appTheme by viewModel.appTheme.collectAsState()

      val focusRequester = remember { FocusRequester() }

      val bgColor = backgroundColor()
      val textColor = primaryTextColor()
      val subTextColor = secondaryTextColor()
      val editorBg = secondaryBackgroundColor()
      val borderClr = borderColor()
      val placeholderClr = placeholderTextColor()
      val sendBtnBg = sendButtonBackgroundColor()

      val density = LocalDensity.current
      val imeInsets = WindowInsets.ime
      LaunchedEffect(Unit) {
          kotlinx.coroutines.delay(350)
          focusRequester.requestFocus()
      }
      LaunchedEffect(Unit) {
          var prevImeVisible = false
          snapshotFlow { imeInsets.getBottom(density) > 0 }
              .collect { isImeVisible ->
                  if (prevImeVisible && !isImeVisible) {
                      viewModel.triggerImmediateSend()
                  }
                  prevImeVisible = isImeVisible
              }
      }

      val isConnected = status is ConnectionStatus.Connected
      val shouldBlinkStatusDot = isConnected && !sendTimedOut

      val infiniteTransition = rememberInfiniteTransition(label = "breathing")
      val breathAlpha by infiniteTransition.animateFloat(
          initialValue = 1f,
          targetValue = 0.4f,
          animationSpec = infiniteRepeatable(
              animation = tween(1000),
              repeatMode = RepeatMode.Reverse
          ),
          label = "breathAlpha"
      )
      val breathScale by infiniteTransition.animateFloat(
          initialValue = 1f,
          targetValue = 1.25f,
          animationSpec = infiniteRepeatable(
              animation = tween(1000),
              repeatMode = RepeatMode.Reverse
          ),
          label = "breathScale"
      )

      Box(modifier = Modifier.fillMaxSize().background(bgColor)) {
          Column(modifier = Modifier.fillMaxSize()) {
              // Status bar
              Row(
                  modifier = Modifier
                      .fillMaxWidth()
                      .padding(horizontal = 16.dp, vertical = 8.dp),
                  verticalAlignment = Alignment.CenterVertically
              ) {
                  val dotColor = when (status) {
                      is ConnectionStatus.Connected -> if (!sendTimedOut) AppColors.statusBarConnected else AppColors.statusBarConnecting
                      is ConnectionStatus.Connecting -> AppColors.statusBarConnecting
                      is ConnectionStatus.Error -> AppColors.statusBarError
                      ConnectionStatus.Disconnected -> AppColors.statusBarDisconnected
                  }
                  val label = when (status) {
                      is ConnectionStatus.Connected -> "已连接: ${(status as ConnectionStatus.Connected).host}"
                      is ConnectionStatus.Connecting -> "连接中..."
                      is ConnectionStatus.Error -> "连接失败，正在重试"
                      ConnectionStatus.Disconnected -> "未连接"
                  }

                  if (shouldBlinkStatusDot) {
                      Box(
                          modifier = Modifier
                              .size(8.dp)
                              .scale(breathScale)
                              .clip(CircleShape)
                              .background(dotColor.copy(alpha = breathAlpha))
                      )
                  } else {
                      Box(
                          modifier = Modifier
                              .size(8.dp)
                              .clip(CircleShape)
                              .background(dotColor)
                      )
                  }
                  Spacer(modifier = Modifier.width(8.dp))
                  Text(
                      text = label,
                      fontSize = 13.sp,
                      fontWeight = FontWeight.Medium,
                      color = textColor.copy(alpha = 0.8f)
                  )

                  Spacer(modifier = Modifier.weight(1f))

                  Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                      // Theme toggle button
                      IconButton(
                          onClick = { viewModel.toggleTheme() },
                          modifier = Modifier
                              .size(28.dp)
                              .background(chipBackgroundColor(), CircleShape)
                      ) {
                          Text(
                              text = if (appTheme == "light") "🌙" else "☀️",
                              fontSize = 13.sp
                          )
                      }

                      // QR scanner button
                      IconButton(
                          onClick = onScanQr,
                          modifier = Modifier
                              .size(28.dp)
                              .background(chipBackgroundColor(), CircleShape)
                      ) {
                          Text(
                              text = "📷",
                              fontSize = 13.sp
                          )
                      }
                  }
              }

              // AutoSendCountdownBar (drains from 1 to 0 over 1.5s)
              AutoSendCountdownBar(
                  active = countdownActive,
                  token = countdownToken,
                  duration = 1500L
              )

              // Main content
              Column(
                  modifier = Modifier
                      .fillMaxWidth()
                      .weight(1f)
                      .padding(top = 12.dp),
                  verticalArrangement = Arrangement.spacedBy(20.dp)
              ) {
                  // Editor
                  Box(
                      modifier = Modifier
                          .fillMaxWidth()
                          .weight(1f)
                          .padding(horizontal = 20.dp)
                          .border(1.dp, borderClr, RoundedCornerShape(16.dp))
                          .background(editorBg, RoundedCornerShape(16.dp))
                  ) {
                      if (text.isEmpty()) {
                          Text(
                              text = "在此输入，或使用键盘麦克风语音输入...",
                              color = placeholderClr,
                              modifier = Modifier
                                  .padding(horizontal = 16.dp, vertical = 16.dp)
                          )
                      }
                      TextField(
                          value = text,
                          onValueChange = { viewModel.updateInputText(it) },
                          modifier = Modifier
                              .fillMaxSize()
                              .focusRequester(focusRequester),
                          colors = TextFieldDefaults.colors(
                              focusedContainerColor = Color.Transparent,
                              unfocusedContainerColor = Color.Transparent,
                              disabledContainerColor = Color.Transparent,
                              focusedIndicatorColor = Color.Transparent,
                              unfocusedIndicatorColor = Color.Transparent
                          ),
                          textStyle = LocalTextStyle.current.copy(color = textColor)
                      )
                  }

                  // Bottom controls
                  Column(
                      modifier = Modifier
                          .fillMaxWidth()
                          .padding(horizontal = 20.dp)
                          .padding(bottom = 20.dp),
                      horizontalAlignment = Alignment.CenterHorizontally,
                      verticalArrangement = Arrangement.spacedBy(12.dp)
                  ) {
                      if (inFlight) {
                          Row(
                              verticalAlignment = Alignment.CenterVertically,
                              horizontalArrangement = Arrangement.spacedBy(6.dp),
                              modifier = Modifier
                                  .background(chipBackgroundColor(), RoundedCornerShape(14.dp))
                                  .padding(horizontal = 12.dp, vertical = 6.dp)
                          ) {
                              CircularProgressIndicator(
                                  modifier = Modifier.size(12.dp),
                                  strokeWidth = 2.dp,
                                  color = subTextColor
                              )
                              Text("发送中", fontSize = 12.sp, color = subTextColor)
                              TextButton(
                                  onClick = { viewModel.cancelSend() },
                                  contentPadding = PaddingValues(0.dp),
                                  modifier = Modifier.height(20.dp)
                              ) {
                                  Text("取消", fontSize = 12.sp, fontWeight = FontWeight.SemiBold, color = AppColors.accent)
                              }
                          }
                      } else if (!isConnected) {
                          Text(
                              text = "请先扫码连接电脑",
                              fontSize = 12.sp,
                              color = subTextColor
                          )
                      }

                      Button(
                          onClick = {
                              viewModel.manualSend()
                          },
                          modifier = Modifier
                              .fillMaxWidth()
                              .height(44.dp),
                          shape = RoundedCornerShape(22.dp),
                          colors = ButtonDefaults.buttonColors(
                              containerColor = sendBtnBg,
                              disabledContainerColor = sendBtnBg.copy(alpha = 0.5f)
                          ),
                          enabled = isConnected && !inFlight
                      ) {
                          Icon(
                              imageVector = Icons.AutoMirrored.Filled.Send,
                              contentDescription = null,
                              modifier = Modifier.size(16.dp)
                          )
                          Spacer(modifier = Modifier.width(8.dp))
                          Text(
                              "发送到电脑",
                              fontSize = 15.sp,
                              fontWeight = FontWeight.SemiBold
                          )
                      }

                      InputMethodTipsView()
                  }
              }
          }
      }
  }

  @Composable
  fun AutoSendCountdownBar(active: Boolean, token: Int, duration: Long) {
      val progress = remember { Animatable(0f) }

      LaunchedEffect(token) {
          if (active) {
              progress.snapTo(1f)
              progress.animateTo(
                  targetValue = 0f,
                  animationSpec = tween(durationMillis = duration.toInt(), easing = LinearEasing)
              )
          } else {
              progress.snapTo(0f)
          }
      }

      Box(
          modifier = Modifier
              .fillMaxWidth()
              .height(3.dp)
              .background(Color.Transparent)
      ) {
          if (active) {
              Box(
                  modifier = Modifier
                      .fillMaxHeight()
                      .fillMaxWidth(progress.value)
                      .background(countdownBarColor())
              )
          }
      }
  }

  @Composable
  fun InputMethodTipsView() {
      val context = LocalContext.current
      val tipsBg = tipsBackgroundColor()
      val textColor = primaryTextColor()
      val subTextColor = secondaryTextColor()

      Column(
          modifier = Modifier
              .fillMaxWidth()
              .clip(RoundedCornerShape(12.dp))
              .background(tipsBg)
              .padding(14.dp),
          verticalArrangement = Arrangement.spacedBy(10.dp)
      ) {
          Row(
              verticalAlignment = Alignment.CenterVertically,
              horizontalArrangement = Arrangement.spacedBy(8.dp)
          ) {
              Icon(
                  imageVector = Icons.Default.Info,
                  contentDescription = null,
                  modifier = Modifier.size(16.dp),
                  tint = textColor
              )
              Text(
                  "语音输入需要第三方输入法",
                  fontSize = 12.sp,
                  fontWeight = FontWeight.SemiBold,
                  color = textColor
              )
          }

          Text(
              "优先使用豆包或微信输入法，在键盘中点击麦克风说话：",
              fontSize = 12.sp,
              color = subTextColor
          )

          Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
              AppMarketLinkButton("豆包输入法", "https://www.doubao.com/")
              AppMarketLinkButton("微信输入法", "https://z.weixin.qq.com/")
          }

          Text(
              "安装后前往 系统设置 → 语言与输入法 启用并切换键盘",
              fontSize = 10.sp,
              color = subTextColor
          )
      }
  }

  @Composable
  fun AppMarketLinkButton(title: String, url: String) {
      val context = LocalContext.current
      val accent = AppColors.accent
      val isDark = LocalIsDarkTheme.current
      val btnBg = accent.copy(alpha = if (isDark) 0.15f else 0.10f)

      TextButton(
          onClick = {
              val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url))
              context.startActivity(intent)
          },
          contentPadding = PaddingValues(horizontal = 10.dp, vertical = 6.dp),
          colors = ButtonDefaults.textButtonColors(
              containerColor = btnBg
          ),
          modifier = Modifier.height(28.dp),
          shape = RoundedCornerShape(8.dp)
      ) {
          Text(
              text = "↗ $title",
              fontSize = 12.sp,
              fontWeight = FontWeight.Medium,
              color = accent
          )
      }
  }
  ```

- [ ] **Step 2: Update QRScannerScreen to use lenientJson and cancel option**
  Modify [QRScannerScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/screens/QRScannerScreen.kt):
  
  In `imageAnalysis.setAnalyzer` (line 121):
  Target Content:
  ```kotlin
  val payload = Json.decodeFromString<PairingPayload>(rawValue)
  ```
  Replacement Content:
  ```kotlin
  val lenientJson = kotlinx.serialization.json.Json { ignoreUnknownKeys = true }
  val payload = lenientJson.decodeFromString<PairingPayload>(rawValue)
  ```

- [ ] **Step 3: Update OnboardingScreen with Theme Toggle**
  Modify [OnboardingScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/screens/OnboardingScreen.kt):
  
  Replace the entire file with:
  ```kotlin
  package com.yule.airvoice.ui.screens

  import androidx.compose.foundation.background
  import androidx.compose.foundation.border
  import androidx.compose.foundation.layout.*
  import androidx.compose.foundation.shape.CircleShape
  import androidx.compose.foundation.shape.RoundedCornerShape
  import androidx.compose.material3.*
  import androidx.compose.runtime.Composable
  import androidx.compose.runtime.collectAsState
  import androidx.compose.runtime.getValue
  import androidx.compose.ui.Alignment
  import androidx.compose.ui.Modifier
  import androidx.compose.ui.graphics.Color
  import androidx.compose.ui.text.font.FontWeight
  import androidx.compose.ui.unit.dp
  import androidx.compose.ui.unit.sp
  import com.yule.airvoice.ui.theme.*

  @Composable
  fun OnboardingScreen(
      onStartScanning: () -> Unit,
      onToggleTheme: () -> Unit
  ) {
      val bgColor = backgroundColor()
      val textColor = primaryTextColor()
      val subTextColor = secondaryTextColor()
      val cardBg = secondaryBackgroundColor()
      val borderClr = borderColor()

      Surface(
          modifier = Modifier.fillMaxSize(),
          color = bgColor
      ) {
          Column(
              modifier = Modifier
                  .fillMaxSize()
                  .padding(24.dp),
              horizontalAlignment = Alignment.CenterHorizontally
          ) {
              // Header Theme Toggle
              Row(
                  modifier = Modifier
                      .fillMaxWidth()
                      .padding(top = 8.dp),
                  horizontalArrangement = Arrangement.End
              ) {
                  IconButton(
                      onClick = onToggleTheme,
                      modifier = Modifier
                          .size(36.dp)
                          .background(chipBackgroundColor(), CircleShape)
                  ) {
                      Text(
                          text = if (LocalIsDarkTheme.current) "☀️" else "🌙",
                          fontSize = 16.sp
                      )
                  }
              }

              Spacer(modifier = Modifier.weight(1f))

              // Icon
              Box(
                  modifier = Modifier
                      .size(80.dp)
                      .background(AppColors.accent.copy(alpha = 0.15f), CircleShape),
                  contentAlignment = Alignment.Center
              ) {
                  Text(
                      text = "🎙️",
                      fontSize = 36.sp
                  )
              }

              Spacer(modifier = Modifier.height(16.dp))

              // Title
              Text(
                  text = "Airvoice",
                  fontSize = 32.sp,
                  fontWeight = FontWeight.Bold,
                  color = textColor
              )
              Spacer(modifier = Modifier.height(8.dp))
              Text(
                  text = "让手机语音输入无缝连接电脑",
                  fontSize = 16.sp,
                  color = subTextColor
              )

              Spacer(modifier = Modifier.weight(1f))

              // Guide card
              Card(
                  modifier = Modifier.fillMaxWidth(),
                  shape = RoundedCornerShape(20.dp),
                  colors = CardDefaults.cardColors(containerColor = cardBg),
                  border = androidx.compose.foundation.BorderStroke(1.dp, borderClr)
              ) {
                  Column(modifier = Modifier.padding(20.dp)) {
                      Text(
                          text = "输入法安装与配置指南",
                          fontWeight = FontWeight.Bold,
                          fontSize = 16.sp,
                          color = textColor,
                          modifier = Modifier.padding(bottom = 16.dp)
                      )
                      GuideStep("1", "安装推荐输入法", "推荐使用「豆包输入法」或「微信输入法」", textColor, subTextColor)
                      Spacer(modifier = Modifier.height(16.dp))
                      GuideStep("2", "启用键盘", "前往「系统设置」→「语言与输入法」→「管理键盘」", textColor, subTextColor)
                      Spacer(modifier = Modifier.height(16.dp))
                      GuideStep("3", "隐私安全", "Airvoice 仅读取本 App 内的输入框，无需额外权限", textColor, subTextColor)
                  }
              }

              Spacer(modifier = Modifier.weight(1f))

              // Start button
              Button(
                  onClick = onStartScanning,
                  modifier = Modifier
                      .fillMaxWidth()
                      .height(56.dp),
                  shape = RoundedCornerShape(28.dp),
                  colors = ButtonDefaults.buttonColors(containerColor = AppColors.accent)
              ) {
                  Text(text = "开始使用", fontSize = 16.sp, fontWeight = FontWeight.Bold, color = Color.White)
              }
          }
      }
  }

  @Composable
  private fun GuideStep(
      number: String,
      title: String,
      desc: String,
      titleColor: Color,
      descColor: Color
  ) {
      Row(
          horizontalArrangement = Arrangement.spacedBy(12.dp),
          modifier = Modifier.fillMaxWidth()
      ) {
          Box(
              modifier = Modifier
                  .size(24.dp)
                  .background(AppColors.accent.copy(alpha = 0.2f), CircleShape)
                  .border(1.dp, AppColors.accent.copy(alpha = 0.5f), CircleShape),
              contentAlignment = Alignment.Center
          ) {
              Text(
                  text = number,
                  fontSize = 12.sp,
                  fontWeight = FontWeight.Bold,
                  color = titleColor
              )
          }
          Column(modifier = Modifier.weight(1f)) {
              Text(
                  text = title,
                  fontSize = 14.sp,
                  fontWeight = FontWeight.SemiBold,
                  color = titleColor
              )
              Text(
                  text = desc,
                  fontSize = 12.sp,
                  color = descColor
              )
          }
      }
  }
  ```

- [ ] **Step 4: Run all tests to verify full compilation & correctness**
  Run: `mise run android:test`
  Expected: PASS.

- [ ] **Step 5: Commit UI Alignment changes**
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt android/app/src/main/java/com/yule/airvoice/ui/screens/QRScannerScreen.kt android/app/src/main/java/com/yule/airvoice/ui/screens/OnboardingScreen.kt
  git commit -m "feat: align HomeScreen, OnboardingScreen, and QRScannerScreen UI/UX with iOS"
  ```
