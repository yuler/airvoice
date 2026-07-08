# 17-remove-mobile-onboarding-design

## Description
This design document details the steps to completely remove the first-time onboarding/initialization screen on both the iOS and Android applications. By bypassing this flow, users will immediately enter the main/home interface upon launching either app for the first time. We will also clean up all code related to the onboarding view files, VM states, storage keys, and tests to prevent any zombie code.

## Proposed Solution

### iOS (SwiftUI)
1. **Remove UI**: Delete [OnboardingView.swift](file:///home/yule/Sides/airvoice/ios/Airvoice/Views/OnboardingView.swift).
2. **Update ContentView**: Modify [ContentView.swift](file:///home/yule/Sides/airvoice/ios/Airvoice/Views/ContentView.swift) to directly render `HomeView()`, removing the `@AppStorage("hasSeenOnboarding")` state check.
3. **Clean Project References**: Remove all references to `OnboardingView.swift` in the Xcode project definition [project.pbxproj](file:///home/yule/Sides/airvoice/ios/Airvoice.xcodeproj/project.pbxproj) so compilation succeeds.

### Android (Jetpack Compose)
1. **Remove UI**: Delete [OnboardingScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/cc/yuler/airvoice/ui/screens/OnboardingScreen.kt).
2. **Update AirvoiceViewModel**:
   - In [AirvoiceViewModel.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/cc/yuler/airvoice/ui/viewmodel/AirvoiceViewModel.kt):
     - Remove `ONBOARDING` from the `Screen` enum.
     - Remove `_hasSeenOnboarding` state flow, `hasSeenOnboarding` public flow, and the `completeOnboarding()` function.
     - Remove the preference check `_hasSeenOnboarding.value = storage.hasSeenOnboardingFlow.first()` in the initialization block (`init`).
3. **Update Navigation**:
   - In [MainScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/cc/yuler/airvoice/ui/screens/MainScreen.kt):
     - Remove `hasSeenOnboarding` state collection.
     - Remove `LaunchedEffect(hasSeenOnboarding)` that switches the screen to onboarding.
     - Ensure the default screen remains `Screen.HOME`.
     - Remove the `Screen.ONBOARDING -> OnboardingScreen(...)` branch from the `when(currentScreen)` block.
4. **Update Storage Manager**:
   - In [StorageManager.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/cc/yuler/airvoice/services/StorageManager.kt):
     - Remove the `KEY_HAS_SEEN_ONBOARDING` preference key definition.
     - Remove the `hasSeenOnboardingFlow` property.
     - Remove the `saveHasSeenOnboarding()` method.
5. **Update Storage Manager Tests**:
   - In [StorageManagerTest.kt](file:///home/yule/Sides/airvoice/android/app/src/test/java/cc/yuler/airvoice/services/StorageManagerTest.kt):
     - Delete the test function `testSaveOnboarding()`.

---

## Architectural & Code Changes

### Files to Delete
- [OnboardingView.swift](file:///home/yule/Sides/airvoice/ios/Airvoice/Views/OnboardingView.swift)
- [OnboardingScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/cc/yuler/airvoice/ui/screens/OnboardingScreen.kt)

### Files to Modify
- [ContentView.swift](file:///home/yule/Sides/airvoice/ios/Airvoice/Views/ContentView.swift)
- [project.pbxproj](file:///home/yule/Sides/airvoice/ios/Airvoice.xcodeproj/project.pbxproj)
- [AirvoiceViewModel.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/cc/yuler/airvoice/ui/viewmodel/AirvoiceViewModel.kt)
- [MainScreen.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/cc/yuler/airvoice/ui/screens/MainScreen.kt)
- [StorageManager.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/cc/yuler/airvoice/services/StorageManager.kt)
- [StorageManagerTest.kt](file:///home/yule/Sides/airvoice/android/app/src/test/java/cc/yuler/airvoice/services/StorageManagerTest.kt)

---

## Verification Plan

### Automated Tests
- For Android: Run unit tests to verify storage manager changes do not break other flows.
  - Run `./gradlew :app:testDebugUnitTest --tests cc.yuler.airvoice.services.StorageManagerTest` (or equivalent via Gradle).
- For iOS: Build and compile the Xcode workspace to verify compilation succeeds without syntax or linker errors.

### Manual Verification
- Launch both iOS and Android apps on a clean environment/fresh install.
- Verify that the application boots directly into the `Home` view instead of presenting the guide steps.
