# 23-mobile-app-settings-design

## Overview

Add a unified, minimalist mobile Settings sheet to both the iOS and Android apps so users can customize appearance, language, voice-sending behavior, connection persistence, and inspect app/OS diagnostics without bloating the main dictation screen.

## Design Decisions

| Area                     | Decision                                                                                                                                                                                                                                               |
| :----------------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Presentation             | Native bottom sheet (iOS `.sheet`, Android `ModalBottomSheet`) styled with `DESIGN.md` monochrome + `#006efe` tokens                                                                                                                                   |
| Home Entry & Cleanup     | Top bar right keeps two circular buttons: `[QR Scanner]` and `[Settings]`; remove the top theme button and bottom version text from Home                                                                                                               |
| Theme                    | 3-way segmented control: `System / Light / Dark` (fresh default: `System`; existing `light` / `dark` values preserved)                                                                                                                                 |
| Language                 | 3-way segmented control: `System / 简体中文 / English` (fresh default: `System`, Chinese locale → `zh`, others → `en`); full app UI & toasts switch immediately without restart                                                                         |
| Input Method Tips        | Toggleable in Settings; Chinese mode keeps Doubao & WeType links, English mode shows generic keyboard dictation guidance                                                                                                                               |
| General Controls         | Auto-send toggle (default on), auto-send delay (`0.8s / 1.5s / 2.5s / 4.0s`, default `1.5s`, hidden when auto-send is off), haptic feedback toggle (default on), keep screen awake toggle (default on, iOS + Android), tips toggle, send timeout       |
| Connection & Persistence | Display paired server address (or unpaired state); destructive "Disconnect & Clear Pairing" with confirmation alert; "Reset to Defaults" with confirmation alert (preserves pairing credentials)                                                       |
| About                    | Display dynamic App Version (`vX.Y.Z`) and OS Version (`iOS X.Y` / `Android X (API Y)`); tapping copies a diagnostic string to clipboard with toast confirmation                                                                                       |

## Persistence Schema

| Setting              | Default  | iOS (`UserDefaults`)    | Android (`DataStore`)     |
| :------------------- | :------- | :---------------------- | :------------------------ |
| Theme mode           | `system` | `appTheme`              | `app_theme`               |
| Language             | `system` | `appLanguage`           | `app_language`            |
| Auto-send enabled    | `true`   | `autoSendEnabled`       | `auto_send_enabled`       |
| Auto-send delay (ms) | `1500`   | `autoSendDelayMs`       | `auto_send_delay_ms`      |
| Haptic feedback      | `true`   | `hapticFeedbackEnabled` | `haptic_feedback_enabled` |
| Keep screen awake    | `true`   | `keepScreenAwake`       | `keep_screen_awake`       |
| Show input tips      | `true`   | `showInputTips`         | `show_input_tips`         |
| Send timeout (ms)    | `5000`   | `sendAckTimeoutMs`      | `send_timeout_ms`         |

## Architecture

### Android (`android/app/src/main/java/cc/yuler/airvoice/`)
- `models/AppSettings.kt`: `ThemeMode`, `AppLanguage`, `AppSettings`, discrete option constants.
- `ui/i18n/AppStrings.kt`: `AppStrings` with `ZhStrings` / `EnStrings` and `LocalAppStrings`.
- `services/StorageManager.kt`: DataStore preference keys, `settingsFlow`, per-field save helpers, and `resetSettings()` (leaves pairing credentials intact).
- `services/AutoSendController.kt`: Dynamic `isAutoSendEnabled`, `autoSendDelayMs`, `sendTimeoutMs` providers; disables debounce countdown and IME-dismiss immediate send when auto-send is off.
- `ui/viewmodel/AirvoiceViewModel.kt`: Exposes `settings`, `pairedWsUrl`, localized toasts, haptic gating, settings mutations, `resetSettings()`, and `disconnectAndClear()`.
- `ui/screens/SettingsBottomSheet.kt`: Material 3 `ModalBottomSheet` with Appearance & Language, General, Connection & Data, and About sections plus confirmation dialogs.
- `ui/screens/MainScreen.kt`, `HomeScreen.kt`, `QRScannerScreen.kt`: Wire `LocalAppStrings`, system theme resolution, `FLAG_KEEP_SCREEN_ON`, updated top bar, and Settings sheet trigger.

### iOS (`ios/Airvoice/`)
- `Utilities/AppTheme.swift`: `ThemeMode` (`system`, `light`, `dark`) and `AppTheme.resolved(mode:colorScheme:)`.
- `Utilities/AppStrings.swift`: `AppLanguage` (`system`, `zh`, `en`) and `AppStrings` (`zh` / `en`).
- `Utilities/AppSettings.swift`: Injectable `UserDefaults`-backed `ObservableObject` storing all settings plus `resetToDefaults()`.
- `Services/AutoSendController.swift`: Configurable `isAutoSendEnabled` and `autoSendDelay`.
- `Services/ConnectionManager.swift`: Expose `savedServerAddress` and `disconnectAndClear()`.
- `Services/HomeViewModel.swift`: Injectable `AppSettings`, localized toasts, haptic gating, configurable timeout.
- `Views/SettingsSheetView.swift`: Native SwiftUI `.sheet` with matching four sections, confirmation alerts, and clipboard copy.
- `Views/HomeView.swift`, `Views/InputMethodTipsView.swift`: Wire theme mode, language strings, keep-awake toggle, tips visibility, top bar cleanup, and Settings sheet.

## Verification Plan

- **Android**: Run `./gradlew testDebugUnitTest assembleDebug` covering `StorageManager`, `AutoSendController`, `AirvoiceViewModel`, and `AppSettings` / `AppStrings` unit tests.
- **iOS**: Add unit tests for `AppSettings` persistence/reset, `AppLanguage` resolution, and `AutoSendController` disabled/custom-delay behavior; verified by GitHub Actions `ios-test` (`xcodegen` + `xcodebuild test`).
