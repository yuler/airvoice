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

## 4. Verification & Network Debugging Address

We will provide a manual test URL for the user to test intranet network reachability between the phone and the PC:
`http://192.168.20.189:7383/health`

Visiting this URL on the phone's browser should return `ok` if the phone and PC are on the same Wi-Fi and there is no firewall blocking port `7383`.
