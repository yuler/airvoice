import Foundation
import SwiftUI

/// Manages app-wide configurations and preferences, persisting them via UserDefaults.
@MainActor
final class AppSettings: ObservableObject {
    static let shared = AppSettings()

    static let themeModeKey = "appTheme"
    static let languageKey = "appLanguage"
    static let autoSendEnabledKey = "autoSendEnabled"
    static let autoSendDelayMsKey = "autoSendDelayMs"
    static let hapticFeedbackEnabledKey = "hapticFeedbackEnabled"
    static let keepScreenAwakeKey = "keepScreenAwake"
    static let showInputTipsKey = "showInputTips"
    static let sendTimeoutMsKey = "sendAckTimeoutMs"

    static let defaultAutoSendDelayMs = 1500
    static let defaultSendTimeoutMs = 5000.0
    static let autoSendDelayOptions = [800, 1500, 2500, 4000]
    static let sendTimeoutOptions: [Double] = [3000, 5000, 10000]

    private let userDefaults: UserDefaults

    @Published var themeMode: ThemeMode {
        didSet { userDefaults.set(themeMode.rawValue, forKey: Self.themeModeKey) }
    }

    @Published var language: AppLanguage {
        didSet { userDefaults.set(language.rawValue, forKey: Self.languageKey) }
    }

    @Published var autoSendEnabled: Bool {
        didSet { userDefaults.set(autoSendEnabled, forKey: Self.autoSendEnabledKey) }
    }

    @Published var autoSendDelayMs: Int {
        didSet {
            let sanitized = Self.sanitizeDelayMs(autoSendDelayMs)
            if sanitized != autoSendDelayMs {
                autoSendDelayMs = sanitized
            }
            userDefaults.set(sanitized, forKey: Self.autoSendDelayMsKey)
        }
    }

    @Published var hapticFeedbackEnabled: Bool {
        didSet { userDefaults.set(hapticFeedbackEnabled, forKey: Self.hapticFeedbackEnabledKey) }
    }

    @Published var keepScreenAwake: Bool {
        didSet { userDefaults.set(keepScreenAwake, forKey: Self.keepScreenAwakeKey) }
    }

    @Published var showInputTips: Bool {
        didSet { userDefaults.set(showInputTips, forKey: Self.showInputTipsKey) }
    }

    /// Timeout in milliseconds for clipboard transmission acknowledgment from CLI.
    @Published var sendTimeoutMs: Double {
        didSet {
            let sanitized = Self.sanitizeTimeoutMs(sendTimeoutMs)
            if sanitized != sendTimeoutMs {
                sendTimeoutMs = sanitized
            }
            userDefaults.set(sanitized, forKey: Self.sendTimeoutMsKey)
        }
    }

    var autoSendDelaySeconds: TimeInterval {
        Double(autoSendDelayMs) / 1000.0
    }

    var strings: AppStrings {
        AppStrings.forLanguage(language)
    }

    init(userDefaults: UserDefaults = .standard) {
        self.userDefaults = userDefaults
        self.themeMode = ThemeMode.fromStorage(userDefaults.string(forKey: Self.themeModeKey))
        self.language = AppLanguage.fromStorage(userDefaults.string(forKey: Self.languageKey))
        self.autoSendEnabled = userDefaults.object(forKey: Self.autoSendEnabledKey) as? Bool ?? true
        self.autoSendDelayMs = Self.sanitizeDelayMs(userDefaults.object(forKey: Self.autoSendDelayMsKey) as? Int)
        self.hapticFeedbackEnabled = userDefaults.object(forKey: Self.hapticFeedbackEnabledKey) as? Bool ?? true
        self.keepScreenAwake = userDefaults.object(forKey: Self.keepScreenAwakeKey) as? Bool ?? true
        self.showInputTips = userDefaults.object(forKey: Self.showInputTipsKey) as? Bool ?? true
        self.sendTimeoutMs = Self.sanitizeTimeoutMs(userDefaults.object(forKey: Self.sendTimeoutMsKey) as? Double)
    }

    /// Resets user preferences to defaults without clearing saved pairing credentials.
    func resetToDefaults() {
        themeMode = .system
        language = .system
        autoSendEnabled = true
        autoSendDelayMs = Self.defaultAutoSendDelayMs
        hapticFeedbackEnabled = true
        keepScreenAwake = true
        showInputTips = true
        sendTimeoutMs = Self.defaultSendTimeoutMs

        userDefaults.removeObject(forKey: Self.themeModeKey)
        userDefaults.removeObject(forKey: Self.languageKey)
        userDefaults.removeObject(forKey: Self.autoSendEnabledKey)
        userDefaults.removeObject(forKey: Self.autoSendDelayMsKey)
        userDefaults.removeObject(forKey: Self.hapticFeedbackEnabledKey)
        userDefaults.removeObject(forKey: Self.keepScreenAwakeKey)
        userDefaults.removeObject(forKey: Self.showInputTipsKey)
        userDefaults.removeObject(forKey: Self.sendTimeoutMsKey)
    }

    static func sanitizeDelayMs(_ value: Int?) -> Int {
        guard let value, autoSendDelayOptions.contains(value) else {
            return defaultAutoSendDelayMs
        }
        return value
    }

    static func sanitizeTimeoutMs(_ value: Double?) -> Double {
        guard let value, sendTimeoutOptions.contains(value) else {
            return defaultSendTimeoutMs
        }
        return value
    }
}
