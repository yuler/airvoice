# Design Spec: Android UI/UE Alignment & PR Review Fixes

To achieve parity with the iOS client and resolve the PR review feedback on PR #2, this design spec outlines the implementation of:
1. Shared `OkHttpClient` instance.
2. Combined DataStore connections read flow.
3. Persistent manual theme (Light/Dark) toggle and onboarding status.
4. Compose-based `AutoSendCountdownBar` and custom Toast UI.
5. `InputMethodTipsView` for third-party input method guides.
6. Linux compatibility fixes for IP detection in `mise` tasks and JDK environment detection.

---

## 1. Storage & Dependency Injection Changes

### 1.1 AirvoiceApplication
Expose a single lazily-initialized `OkHttpClient` in `AirvoiceApplication`:
```kotlin
package com.yule.airvoice

import android.app.Application
import okhttp3.OkHttpClient

class AirvoiceApplication : Application() {
    val okHttpClient: OkHttpClient by lazy {
        OkHttpClient()
    }
}
```

### 1.2 StorageManager
Define connection, theme, and onboarding preferences:
- `KEY_WS` (String)
- `KEY_TOKEN` (String)
- `KEY_THEME` (String) — `"light"` or `"dark"`. Defaults to `"light"`.
- `KEY_HAS_SEEN_ONBOARDING` (Boolean). Defaults to `false`.

Expose these flows:
- `val connectionInfoFlow: Flow<ConnectionInfo>` where `ConnectionInfo` is `data class ConnectionInfo(val wsUrl: String?, val token: String?)`.
- `val themeFlow: Flow<String>`
- `val hasSeenOnboardingFlow: Flow<Boolean>`

Provide methods to update theme and onboarding:
- `suspend fun saveTheme(theme: String)`
- `suspend fun saveHasSeenOnboarding(completed: Boolean)`

---

## 2. Service & Business Logic Changes

### 2.1 ConnectionManager
Address PR review regarding JSON serialization:
- Replace direct `Json.encodeToString` calls with the class-level configured `lenientJson.encodeToString`.

### 2.2 AutoSendController
Manage and expose countdown states:
- `val countdownActive: StateFlow<Boolean>`
- `val countdownToken: StateFlow<Int>`
- `val inFlight: StateFlow<Boolean>`

On input text changes:
- If trimmed text is not empty and not already sending, increment `countdownToken` and set `countdownActive = true`.
- Start a debounce job (1.5s).
- Cancel any ongoing timer.
- On send trigger, set `inFlight = true`, `countdownActive = false`.

### 2.3 AirvoiceViewModel
Address coordinate states and retry logic:
- Retrieve `okHttpClient` from `AirvoiceApplication`.
- Read and expose `theme` state, `hasSeenOnboarding` state, and `inputText` state.
- Implement manual theme toggle (`toggleTheme()`) and persist in DataStore.
- Implement onboarding completion (`completeOnboarding()`) and persist in DataStore.
- On successful ACK:
  - Clear `inputText` completely (set to `""`), matching iOS behavior.
- On failed ACK:
  - If trigger was `.auto` and not yet retried, wait 400ms and retry sending once in the background.

---

## 3. UI and UX Alignment

### 3.1 Theme Override
Introduce `LocalIsDarkTheme` composition local in `AppTheme.kt`:
```kotlin
val LocalIsDarkTheme = staticCompositionLocalOf { false }

@Composable
fun isDarkTheme(): Boolean = LocalIsDarkTheme.current
```
Define colors for:
- `toastBackground`: Light (`#1C1C1E` at 92%), Dark (`#1F2030` at 95%)
- `countdownBar`: `#ffae00` at 40% (Light) / 45% (Dark)
- `tipsBackground`: Light (`#EFEFF4`), Dark (`#FFFFFF` at 6%)

### 3.2 MainScreen
Coordinate the screen flow and custom Toast overlay:
- Render `OnboardingScreen` if `hasSeenOnboarding` is false.
- Render `HomeScreen` if `hasSeenOnboarding` is true.
- Allow navigating to `QRScannerScreen` and cancelling back.
- Overlay a custom Compose-based Toast at the bottom when a toast event is emitted:
  - Slide-in from bottom, fade-in/out.
  - Capsule shape, dark background, checkmark/warning icons.

### 3.3 HomeScreen
- **Countdown Bar**: Sits directly under the status bar. Animates width from 100% to 0% when `countdownActive` is true, resetting when `countdownToken` increments.
- **Header**:
  - Connected status dot with a Compose-based breathing scale/alpha animation.
  - Text indicating connection status.
  - Manual theme toggle button (moon/sun icon in Circle shape with `chipBackground`).
  - QR scanner button (circle chip).
  - Remove "Re-pair" / Disconnect button from the header.
- **InputMethodTipsView**: Custom styled card at the bottom.
  - Links: 豆包输入法 (`https://www.doubao.com/`), 微信输入法 (`https://z.weixin.qq.com/`).
  - Text: "安装后前往 系统设置 → 语言与输入法 启用并切换键盘".

---

## 4. Developer Tools & Environment

### 4.1 mise.toml
Update IP detection for Linux:
```toml
IP=$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || hostname -I 2>/dev/null | awk '{print $1}' || ip route get 1 2>/dev/null | awk '{print $(NF-2);exit}' || echo localhost);
```

### 4.2 java-env.sh
Add a Linux fallback search for Android Studio's embedded JBR:
```bash
if [ -z "$JAVA_HOME" ] || [ ! -x "$JAVA_HOME/bin/java" ]; then
  if [ -d "/opt/android-studio/jbr" ] && [ -x "/opt/android-studio/jbr/bin/java" ]; then
    JAVA_HOME="/opt/android-studio/jbr"
  elif [ -d "$HOME/android-studio/jbr" ] && [ -x "$HOME/android-studio/jbr/bin/java" ]; then
    JAVA_HOME="$HOME/android-studio/jbr"
  fi
fi
```
