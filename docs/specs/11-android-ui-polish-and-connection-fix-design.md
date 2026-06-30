# Design Spec: Android UI Polish & Connection Fix

To align the Android native client's layout with iOS, handle system window insets, enable autofocus keyboard popup, and fix the local WebSocket connection failure.

---

## 1. UI Layout Padding & Safe Areas

Currently, the Android app has `enableEdgeToEdge()` enabled in `MainActivity`, but `HomeScreen.kt` does not apply status bar or navigation bar padding. This causes the header to overlap with the system status bar and the bottom content to overlap with the system navigation bar.

### 1.1 Status Bar & Navigation Bar Padding
We will apply `statusBarsPadding()` and `navigationBarsPadding()` on the main `Column` container in `HomeScreen.kt` so that:
- The top header starts below the system status bar.
- The bottom action buttons and tips are completely within the safe area.

### 1.2 Editor Box Height
To match the iOS client, we will limit the text editor Box height to a neat fixed height of `180.dp` instead of letting it expand via `.weight(1f)`. This ensures it looks clean, consistent, and provides a polished layout whether the keyboard is up or down.

---

## 2. Keyboard Autofocus on Startup

We will retrieve `LocalSoftwareKeyboardController.current` in `HomeScreen.kt`. In the `LaunchedEffect(Unit)` where focus is requested, we will explicitly call `keyboardController?.show()` to pull up the soft keyboard on app launch, replicating the iOS client's autofocus behavior.

---

## 3. Cleartext Network Traffic Configuration

On Android 9 (API 28) and higher, cleartext traffic (unencrypted HTTP and `ws://` WebSocket connections) is disabled by default. Since the local pairing WebSocket server uses unencrypted `ws://` protocols on local subnets, OkHttp attempts fail immediately.

We will add `android:usesCleartextTraffic="true"` to the `<application>` tag in `AndroidManifest.xml` to allow cleartext WebSocket connections during local testing and use.

---

## 4. UI Icons & Button Polish

### 4.1 SVG Vector Icons & Header Spacing
We will replace the raw emojis in the top right corner with Compose Vector Icons:
- **Theme Toggle**: Use `Icons.Default.Brightness4` and `Icons.Default.Brightness7`.
- **QR Scanner**: Use `Icons.Default.PhotoCamera`.
These icons will be styled using standard Compose `Icon` components and tinted with `primaryTextColor()`.
To ensure they are visually distinct, the horizontal spacing between the two icons will be increased to `12.dp` using `Arrangement.spacedBy(12.dp)`.

### 4.2 Custom Send Button (iOS Parity)
Instead of using the default Material3 `Button` (which introduces Android-specific elevation, paddings, and styles), we will implement a custom `Box` with `clickable` layout:
- **Shape**: Rounded corner clip of `22.dp` (pill-shape).
- **Background**: `sendButtonBackgroundColor()` (active) or `sendButtonBackgroundColor().copy(alpha = 0.5f)` (disabled).
- **Content**: Row containing the Send icon and text, colored with `primaryTextColor()` (active) or `secondaryTextColor().copy(alpha = 0.5f)` (disabled).
- **Elevation**: Flat layout with no shadows, matching the iOS design.

---

## 5. WebSocket Message & Collector Debug Logging

To help debug connection issues and ensure ACKs are being received:
- Add a debug log statement `Log.d("ConnectionManager", "Received: $text")` inside the `onMessage` callback of `ConnectionManager.kt`.
- Add a debug log statement `Log.d("AutoSendController", "Collector received msg: type=${msg.type}, id=${msg.id}, pendingMessageId=$pendingMessageId")` inside the `incomingMessages.collect` block of `AutoSendController.kt` to inspect if the ACK message reaches the controller and if the ID matches.

---

## 6. Verification & Network Debugging Address

We will provide a manual test URL for the user to test intranet network reachability between the phone and the PC:
`http://192.168.20.189:7383/health`

Visiting this URL on the phone's browser should return `ok` if the phone and PC are on the same Wi-Fi and there is no firewall blocking port `7383`.

