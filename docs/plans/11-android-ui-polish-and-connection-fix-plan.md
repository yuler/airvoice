# Android UI Polish & Connection Fix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Polish the Android client's layout to respect system window insets, size the text input box to match iOS, auto-show the keyboard on startup, and allow local cleartext WebSocket connections.

**Architecture:** We will configure the Android Manifest to allow cleartext network traffic, add Compose safe area/window inset padding to the home screen container, constrain the editor to a fixed height of `180.dp`, and request the software keyboard on app launch using `LocalSoftwareKeyboardController`.

**Tech Stack:** Kotlin, Jetpack Compose, Android Manifest configurations.

---

### Task 1: Allow Cleartext Network Traffic in Android Manifest

**Files:**
- Modify: `android/app/src/main/AndroidManifest.xml`

- [ ] **Step 1: Enable cleartext traffic in AndroidManifest.xml**
  Modify [AndroidManifest.xml](file:///home/yule/Sides/airvoice/android/app/src/main/AndroidManifest.xml) to add `android:usesCleartextTraffic="true"` to the `<application>` tag.
  
  Target Content (around line 8-15):
  ```xml
      <application
          android:name=".AirvoiceApplication"
          android:allowBackup="true"
          android:icon="@android:drawable/sym_def_app_icon"
          android:label="Airvoice"
          android:roundIcon="@android:drawable/sym_def_app_icon"
          android:supportsRtl="true"
          android:theme="@android:style/Theme.Material.NoActionBar">
  ```
  
  Replacement Content:
  ```xml
      <application
          android:name=".AirvoiceApplication"
          android:allowBackup="true"
          android:icon="@android:drawable/sym_def_app_icon"
          android:label="Airvoice"
          android:roundIcon="@android:drawable/sym_def_app_icon"
          android:supportsRtl="true"
          android:theme="@android:style/Theme.Material.NoActionBar"
          android:usesCleartextTraffic="true">
  ```

- [ ] **Step 2: Commit manifest change**
  Run:
  ```bash
  git add android/app/src/main/AndroidManifest.xml
  git commit -m "feat: allow cleartext HTTP and WebSocket traffic in AndroidManifest"
  ```

---

### Task 2: Apply System Window Insets Padding on HomeScreen

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt`

- [ ] **Step 1: Add status bar and navigation bar padding to the outer Column**
  Modify the outer `Column` inside [HomeScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt) (around line 102) to apply `.statusBarsPadding()` and `.navigationBarsPadding()`.
  
  Target Content (around line 101-103):
  ```kotlin
      Box(modifier = Modifier.fillMaxSize().background(bgColor)) {
          Column(modifier = Modifier.fillMaxSize()) {
              // Status bar
  ```
  
  Replacement Content:
  ```kotlin
      Box(modifier = Modifier.fillMaxSize().background(bgColor)) {
          Column(
              modifier = Modifier
                  .fillMaxSize()
                  .statusBarsPadding()
                  .navigationBarsPadding()
          ) {
              // Status bar
  ```

- [ ] **Step 2: Verify Android code builds**
  Run: `mise run android:test`
  Expected: BUILD SUCCESSFUL and tests pass.

- [ ] **Step 3: Commit window insets change**
  Run:
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt
  git commit -m "style: apply status bar and navigation bar padding to HomeScreen"
  ```

---

### Task 3: Adjust Editor Box to Fixed Height

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt`

- [ ] **Step 1: Change editor Box modifier from weight(1f) to height(180.dp)**
  Modify the editor `Box` inside [HomeScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt) (around line 194) to replace `.weight(1f)` with `.height(180.dp)`.
  
  Target Content (around line 194-201):
  ```kotlin
                  // Editor
                  Box(
                      modifier = Modifier
                          .fillMaxWidth()
                          .weight(1f)
                          .padding(horizontal = 20.dp)
                          .border(1.dp, borderClr, RoundedCornerShape(16.dp))
                          .background(editorBg, RoundedCornerShape(16.dp))
                  ) {
  ```
  
  Replacement Content:
  ```kotlin
                  // Editor
                  Box(
                      modifier = Modifier
                          .fillMaxWidth()
                          .height(180.dp)
                          .padding(horizontal = 20.dp)
                          .border(1.dp, borderClr, RoundedCornerShape(16.dp))
                          .background(editorBg, RoundedCornerShape(16.dp))
                  ) {
  ```

- [ ] **Step 2: Verify Android code builds**
  Run: `mise run android:test`
  Expected: BUILD SUCCESSFUL.

- [ ] **Step 3: Commit editor height change**
  Run:
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt
  git commit -m "style: adjust editor box to fixed height of 180.dp to match iOS"
  ```

---

### Task 4: Implement Keyboard Autofocus on Startup

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt`

- [ ] **Step 1: Import LocalSoftwareKeyboardController and request keyboard popup**
  Modify [HomeScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt) to import `androidx.compose.ui.platform.LocalSoftwareKeyboardController`. Get the keyboard controller in `HomeScreen` and call `.show()` in the startup `LaunchedEffect`.
  
  Target Content (around line 29-37):
  ```kotlin
  import androidx.compose.ui.draw.scale
  import androidx.compose.ui.focus.FocusRequester
  import androidx.compose.ui.focus.focusRequester
  import androidx.compose.ui.graphics.Color
  import androidx.compose.ui.platform.LocalContext
  import androidx.compose.ui.platform.LocalDensity
  import androidx.compose.ui.text.font.FontWeight
  ```
  
  Replacement Content:
  ```kotlin
  import androidx.compose.ui.draw.scale
  import androidx.compose.ui.focus.FocusRequester
  import androidx.compose.ui.focus.focusRequester
  import androidx.compose.ui.graphics.Color
  import androidx.compose.ui.platform.LocalContext
  import androidx.compose.ui.platform.LocalDensity
  import androidx.compose.ui.platform.LocalSoftwareKeyboardController
  import androidx.compose.ui.text.font.FontWeight
  ```
  
  Target Content (around line 63-66):
  ```kotlin
      LaunchedEffect(Unit) {
          kotlinx.coroutines.delay(350)
          focusRequester.requestFocus()
      }
  ```
  
  Replacement Content:
  ```kotlin
      val keyboardController = LocalSoftwareKeyboardController.current
      LaunchedEffect(Unit) {
          kotlinx.coroutines.delay(350)
          focusRequester.requestFocus()
          keyboardController?.show()
      }
  ```

- [ ] **Step 2: Verify Android code builds**
  Run: `mise run android:test`
  Expected: BUILD SUCCESSFUL and tests pass.

- [ ] **Step 3: Commit autofocus keyboard changes**
  Run:
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt
  git commit -m "feat: show keyboard automatically on startup using LocalSoftwareKeyboardController"
  ```

---

### Task 5: Replace Emojis with Vector Icons in HomeScreen

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt`

- [ ] **Step 1: Import vector icons**
  Modify [HomeScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt) to import the required icons.
  
  Target Content:
  ```kotlin
  import androidx.compose.material.icons.filled.Info
  ```
  
  Replacement Content:
  ```kotlin
  import androidx.compose.material.icons.filled.Info
  import androidx.compose.material.icons.filled.Brightness4
  import androidx.compose.material.icons.filled.Brightness7
  import androidx.compose.material.icons.filled.PhotoCamera
  ```

- [ ] **Step 2: Replace emojis with Icon components**
  Modify the theme toggle and QR scanner icon buttons to render standard Material icons.
  
  Target Content (around line 150-176):
  ```kotlin
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
  ```
  
  Replacement Content:
  ```kotlin
                     // Theme toggle button
                     IconButton(
                         onClick = { viewModel.toggleTheme() },
                         modifier = Modifier
                             .size(28.dp)
                             .background(chipBackgroundColor(), CircleShape)
                     ) {
                         Icon(
                             imageVector = if (appTheme == "light") Icons.Default.Brightness4 else Icons.Default.Brightness7,
                             contentDescription = "Toggle Theme",
                             tint = textColor,
                             modifier = Modifier.size(16.dp)
                         )
                     }

                     // QR scanner button
                     IconButton(
                         onClick = onScanQr,
                         modifier = Modifier
                             .size(28.dp)
                             .background(chipBackgroundColor(), CircleShape)
                     ) {
                         Icon(
                             imageVector = Icons.Default.PhotoCamera,
                             contentDescription = "Scan QR",
                             tint = textColor,
                             modifier = Modifier.size(16.dp)
                         )
                     }
  ```

- [ ] **Step 3: Verify Android code builds**
  Run: `mise run android:test`
  Expected: BUILD SUCCESSFUL.

- [ ] **Step 4: Commit vector icon changes**
  Run:
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt
  git commit -m "style: replace home screen header emojis with Compose vector icons"
  ```

---

### Task 6: Improve Send Button Enabled/Disabled Colors

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt`

- [ ] **Step 1: Set contentColor and disabledContentColor on Send Button**
  Modify [HomeScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt) to style button content and disabled state colors.
  
  Target Content (around line 266-279):
  ```kotlin
                        colors = ButtonDefaults.buttonColors(
                            containerColor = sendBtnBg,
                            disabledContainerColor = sendBtnBg.copy(alpha = 0.5f)
                        ),
  ```
  
  Replacement Content:
  ```kotlin
                        colors = ButtonDefaults.buttonColors(
                            containerColor = sendBtnBg,
                            contentColor = primaryTextColor(),
                            disabledContainerColor = sendBtnBg.copy(alpha = 0.5f),
                            disabledContentColor = secondaryTextColor().copy(alpha = 0.5f)
                        ),
  ```

- [ ] **Step 2: Verify Android code builds**
  Run: `mise run android:test`
  Expected: BUILD SUCCESSFUL.

- [ ] **Step 3: Commit button color changes**
  Run:
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt
  git commit -m "style: distinguish enabled and disabled button colors on Send Button"
  ```

---

### Task 7: Add WebSocket Message Debug Logging on Android

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/services/ConnectionManager.kt`

- [ ] **Step 1: Add log statement inside onMessage**
  Modify [ConnectionManager.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/services/ConnectionManager.kt) to log all incoming WebSocket messages.
  
  Target Content (around line 83-86):
  ```kotlin
              override fun onMessage(webSocket: WebSocket, text: String) {
                  if (webSocket !== this@ConnectionManager.webSocket) return
                  try {
  ```
  
  Replacement Content:
  ```kotlin
              override fun onMessage(webSocket: WebSocket, text: String) {
                  if (webSocket !== this@ConnectionManager.webSocket) return
                  Log.d("ConnectionManager", "Received WebSocket message: $text")
                  try {
  ```

- [ ] **Step 2: Verify Android code builds**
  Run: `mise run android:test`
  Expected: BUILD SUCCESSFUL.

- [ ] **Step 3: Commit debug logging changes**
  Run:
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/services/ConnectionManager.kt
  git commit -m "debug: add log statement for incoming WebSocket messages on Android"
  ```

---

### Task 8: Spacing Between Top-Right Header Icons

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt`

- [ ] **Step 1: Increase spacing to 12.dp for header icons**
  Modify [HomeScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt) (around line 149) to change `Arrangement.spacedBy(8.dp)` to `Arrangement.spacedBy(12.dp)`.
  
  Target Content (around line 147-151):
  ```kotlin
                  Spacer(modifier = Modifier.weight(1f))
  
                  Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                      // Theme toggle button
  ```
  
  Replacement Content:
  ```kotlin
                  Spacer(modifier = Modifier.weight(1f))
  
                  Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                      // Theme toggle button
  ```

- [ ] **Step 2: Verify Android code builds**
  Run: `mise run android:test`
  Expected: BUILD SUCCESSFUL.

- [ ] **Step 3: Commit spacing changes**
  Run:
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt
  git commit -m "style: increase horizontal spacing between top-right header icons to 12.dp"
  ```

---

### Task 9: Refactor Send Button to Custom Box Layout

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt`

- [ ] **Step 1: Refactor Button to Box in HomeScreen.kt**
  Modify [HomeScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt) to replace `Button` with a custom flat `Box` using `clickable`.
  
  Target Content (around line 266-291):
  ```kotlin
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
                                contentColor = primaryTextColor(),
                                disabledContainerColor = sendBtnBg.copy(alpha = 0.5f),
                                disabledContentColor = secondaryTextColor().copy(alpha = 0.5f)
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
  ```
  
  Replacement Content:
  ```kotlin
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(44.dp)
                                .clip(RoundedCornerShape(22.dp))
                                .background(if (isConnected && !inFlight) sendBtnBg else sendBtnBg.copy(alpha = 0.5f))
                                .clickable(enabled = isConnected && !inFlight) { viewModel.manualSend() },
                            contentAlignment = Alignment.Center
                        ) {
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.Center
                            ) {
                                Icon(
                                    imageVector = Icons.AutoMirrored.Filled.Send,
                                    contentDescription = null,
                                    modifier = Modifier.size(16.dp),
                                    tint = if (isConnected && !inFlight) primaryTextColor() else secondaryTextColor().copy(alpha = 0.5f)
                                )
                                Spacer(modifier = Modifier.width(8.dp))
                                Text(
                                    "发送到电脑",
                                    fontSize = 15.sp,
                                    fontWeight = FontWeight.SemiBold,
                                    color = if (isConnected && !inFlight) primaryTextColor() else secondaryTextColor().copy(alpha = 0.5f)
                                )
                            }
                        }
  ```

- [ ] **Step 2: Verify Android code builds**
  Run: `mise run android:test`
  Expected: BUILD SUCCESSFUL.

- [ ] **Step 3: Commit custom button changes**
  Run:
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/ui/screens/HomeScreen.kt
  git commit -m "style: replace material Button with custom flat Box pill button for iOS parity"
  ```

---

### Task 10: Add Collector Debug Logging in AutoSendController

**Files:**
- Modify: `android/app/src/main/java/com/yule/airvoice/services/AutoSendController.kt`

- [ ] **Step 1: Import Log and add print statement inside incomingMessages collector**
  Modify [AutoSendController.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/services/AutoSendController.kt) to log all messages passing through the collector flow.
  
  Target Content (around line 3-4):
  ```kotlin
  import com.yule.airvoice.models.ProtocolMessage
  import java.util.UUID
  ```
  
  Replacement Content:
  ```kotlin
  import com.yule.airvoice.models.ProtocolMessage
  import java.util.UUID
  import android.util.Log
  ```
  
  Target Content (around line 70-74):
  ```kotlin
          scope.launch {
              connectionManager.incomingMessages.collect { msg ->
                  if (msg.type == "ack" && msg.id == pendingMessageId) {
  ```
  
  Replacement Content:
  ```kotlin
          scope.launch {
              connectionManager.incomingMessages.collect { msg ->
                  Log.d("AutoSendController", "Collector received msg: type=${msg.type}, id=${msg.id}, pendingMessageId=$pendingMessageId, success=${msg.ok}")
                  if (msg.type == "ack" && msg.id == pendingMessageId) {
  ```

- [ ] **Step 2: Verify Android code builds and tests pass**
  Run: `mise run android:test`
  Expected: BUILD SUCCESSFUL.

- [ ] **Step 3: Commit collector logging changes**
  Run:
  ```bash
  git add android/app/src/main/java/com/yule/airvoice/services/AutoSendController.kt
  git commit -m "debug: log WebSocket messages inside AutoSendController collector"
  ```


