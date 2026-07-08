# 18-add-mobile-version-badge-design

## Description
This design document details adding a version badge to the home screens of both the iOS and Android applications. The badge will display the current application version (e.g. `v0.3.1`) dynamically retrieved from the application bundles/packages to avoid hardcoding. Following user feedback, the badge will be displayed at the bottom of the home screen, centered, below the input method tips card.

## Proposed Solution

### iOS (SwiftUI)
- In [HomeView.swift](/ios/Airvoice/Views/HomeView.swift), inside the `bottomControls` `VStack`, append a centered `Text` below `InputMethodTipsView`.
- Retrieve the version dynamically using `Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "0.3.1"`.

### Android (Jetpack Compose)
- In [HomeScreen.kt](/android/app/src/main/java/cc/yuler/airvoice/ui/screens/HomeScreen.kt), inside the bottom controls `Column`, append a centered `Text` below `InputMethodTipsView()`.
- Retrieve the version dynamically using the Android `packageManager` to get `getPackageInfo(packageName, 0).versionName`.

---

## Architectural & Code Changes

### Files to Modify
- [HomeView.swift](/ios/Airvoice/Views/HomeView.swift)
- [HomeScreen.kt](/android/app/src/main/java/cc/yuler/airvoice/ui/screens/HomeScreen.kt)

---

## Verification Plan

### Automated Tests
- None required since these are minor presentation/UI changes.
- Ensure the project builds successfully on both iOS and Android platforms.

### Manual Verification
- Compile and run both applications.
- Verify that a centered `v0.3.1` badge is visible below the tips card on the home screen.
