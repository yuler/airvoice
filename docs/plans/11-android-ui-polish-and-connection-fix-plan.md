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
