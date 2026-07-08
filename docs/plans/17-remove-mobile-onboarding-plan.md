# Remove Mobile Onboarding Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the first-time onboarding/initialization screen from both iOS and Android applications, ensuring the user immediately lands on the home view upon launching either app.

**Architecture:** The onboarding logic will be bypassed by directly returning the main view (`HomeView` on iOS, `HomeScreen` on Android) on startup. All unused UI views, ViewModel flows, and storage keys that reference onboarding progress will be completely deleted.

**Tech Stack:** Swift, Kotlin, Jetpack Compose, SwiftUI, Android DataStore

---

### Task 1: iOS Client Cleanup

**Files:**
- Delete: `ios/Airvoice/Views/OnboardingView.swift`
- Modify: `ios/Airvoice/Views/ContentView.swift`
- Modify: `ios/Airvoice.xcodeproj/project.pbxproj`

- [ ] **Step 1: Delete OnboardingView.swift**

Run:
```bash
rm ios/Airvoice/Views/OnboardingView.swift
```

- [ ] **Step 2: Update ContentView.swift to directly load HomeView**

Modify [ContentView.swift](file:///home/yule/Sides/airvoice/ios/Airvoice/Views/ContentView.swift) to:
```swift
import SwiftUI

struct ContentView: View {
    var body: some View {
        HomeView()
    }
}

#Preview {
    ContentView()
}
```

- [ ] **Step 3: Remove file references from project.pbxproj**

In [project.pbxproj](file:///home/yule/Sides/airvoice/ios/Airvoice.xcodeproj/project.pbxproj), delete the following 4 lines:
1. In `PBXBuildFile` section:
   ```swift
   6F6274C1DA308E1986F13E93 /* OnboardingView.swift in Sources */ = {isa = PBXBuildFile; fileRef = C653CDEEA8D67BDEFB5DBCE7 /* OnboardingView.swift */; };
   ```
2. In `PBXFileReference` section:
   ```swift
   C653CDEEA8D67BDEFB5DBCE7 /* OnboardingView.swift */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.swift; path = OnboardingView.swift; sourceTree = "<group>"; };
   ```
3. In `PBXGroup` section (under `Views` children):
   ```swift
   C653CDEEA8D67BDEFB5DBCE7 /* OnboardingView.swift */,
   ```
4. In `PBXSourcesBuildPhase` section:
   ```swift
   6F6274C1DA308E1986F13E93 /* OnboardingView.swift in Sources */,
   ```

- [ ] **Step 4: Commit iOS Changes**

Run:
```bash
git add ios/Airvoice/Views/ContentView.swift ios/Airvoice.xcodeproj/project.pbxproj
git commit -m "refactor(ios): remove onboarding view and references"
```

---

### Task 2: Android Client Storage and ViewModel Cleanup

**Files:**
- Modify: `android/app/src/main/java/cc/yuler/airvoice/services/StorageManager.kt`
- Modify: `android/app/src/main/java/cc/yuler/airvoice/ui/viewmodel/AirvoiceViewModel.kt`
- Modify: `android/app/src/test/java/cc/yuler/airvoice/services/StorageManagerTest.kt`

- [ ] **Step 1: Remove onboarding storage logic in StorageManager.kt**

In [StorageManager.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/cc/yuler/airvoice/services/StorageManager.kt):
1. Remove line 27:
   ```kotlin
           private val KEY_HAS_SEEN_ONBOARDING = booleanPreferencesKey("has_seen_onboarding")
   ```
2. Remove lines 38-40:
   ```kotlin
       val hasSeenOnboardingFlow: Flow<Boolean> = context.dataStore.data
           .handleIOException()
           .map { prefs -> prefs[KEY_HAS_SEEN_ONBOARDING] ?: false }
   ```
3. Remove lines 62-66:
   ```kotlin
       suspend fun saveHasSeenOnboarding(completed: Boolean) {
           context.dataStore.edit { prefs ->
               prefs[KEY_HAS_SEEN_ONBOARDING] = completed
           }
       }
   ```

- [ ] **Step 2: Remove onboarding test in StorageManagerTest.kt**

In [StorageManagerTest.kt](file:///home/yule/Sides/airvoice/android/app/src/test/java/cc/yuler/airvoice/services/StorageManagerTest.kt), remove the following test function (lines 51-56):
```kotlin
    @Test
    fun testSaveOnboarding() = runBlocking {
        assertFalse(storageManager.hasSeenOnboardingFlow.first())
        storageManager.saveHasSeenOnboarding(true)
        assertTrue(storageManager.hasSeenOnboardingFlow.first())
    }
```

- [ ] **Step 3: Remove onboarding references in AirvoiceViewModel.kt**

In [AirvoiceViewModel.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/cc/yuler/airvoice/ui/viewmodel/AirvoiceViewModel.kt):
1. Remove `ONBOARDING` from `Screen` enum (line 22):
   ```kotlin
   enum class Screen {
       SCANNER,
       HOME
   }
   ```
2. Remove state declarations (lines 33-34):
   ```kotlin
       private val _hasSeenOnboarding = MutableStateFlow(false)
       val hasSeenOnboarding: StateFlow<Boolean> = _hasSeenOnboarding.asStateFlow()
   ```
3. Remove storage load step in `init` block (line 59):
   ```kotlin
               _hasSeenOnboarding.value = storage.hasSeenOnboardingFlow.first()
   ```
4. Remove `completeOnboarding` method (lines 86-91):
   ```kotlin
       fun completeOnboarding() {
           viewModelScope.launch {
               storage.saveHasSeenOnboarding(true)
               _hasSeenOnboarding.value = true
           }
       }
   ```

- [ ] **Step 4: Commit Storage and ViewModel changes**

Run:
```bash
git add android/app/src/main/java/cc/yuler/airvoice/services/StorageManager.kt android/app/src/main/java/cc/yuler/airvoice/ui/viewmodel/AirvoiceViewModel.kt android/app/src/test/java/cc/yuler/airvoice/services/StorageManagerTest.kt
git commit -m "refactor(android): remove onboarding keys, state flow, and methods"
```

---

### Task 3: Android Client UI and Navigation Cleanup

**Files:**
- Delete: `android/app/src/main/java/cc/yuler/airvoice/ui/screens/OnboardingScreen.kt`
- Modify: `android/app/src/main/java/cc/yuler/airvoice/ui/screens/MainScreen.kt`

- [ ] **Step 1: Delete OnboardingScreen.kt**

Run:
```bash
rm android/app/src/main/java/cc/yuler/airvoice/ui/screens/OnboardingScreen.kt
```

- [ ] **Step 2: Remove onboarding view and transition logic in MainScreen.kt**

In [MainScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/cc/yuler/airvoice/ui/screens/MainScreen.kt):
1. Remove state collect (line 31):
   ```kotlin
       val hasSeenOnboarding by viewModel.hasSeenOnboarding.collectAsState()
   ```
2. Remove `LaunchedEffect` (lines 38-40):
   ```kotlin
       LaunchedEffect(hasSeenOnboarding) {
           currentScreen = if (hasSeenOnboarding) Screen.HOME else Screen.ONBOARDING
       }
   ```
3. Remove `Screen.ONBOARDING` from `when(currentScreen)` navigation block (lines 47-54):
   ```kotlin
                   Screen.ONBOARDING -> OnboardingScreen(
                       onStartScanning = {
                           viewModel.completeOnboarding()
                       },
                       onToggleTheme = {
                           viewModel.toggleTheme()
                       }
                   )
   ```

- [ ] **Step 3: Commit Navigation and UI changes**

Run:
```bash
git add android/app/src/main/java/cc/yuler/airvoice/ui/screens/MainScreen.kt
git commit -m "refactor(android): remove onboarding screen and transition logic"
```

---

### Task 4: Verification

- [ ] **Step 1: Run Android unit tests**

Run:
```bash
./gradlew :app:testDebugUnitTest
```
Expected: All tests pass.

- [ ] **Step 2: Commit verification changes**

No new commits needed if tests pass without edits.
