import XCTest
import SwiftUI
@testable import Airvoice

@MainActor
final class AppSettingsTests: XCTestCase {
    private var suiteName: String!
    private var defaults: UserDefaults!

    override func setUp() {
        super.setUp()
        suiteName = "cc.yuler.airvoice.settings.tests.\(UUID().uuidString)"
        defaults = UserDefaults(suiteName: suiteName)!
        defaults.removePersistentDomain(forName: suiteName)
    }

    override func tearDown() {
        defaults.removePersistentDomain(forName: suiteName)
        defaults = nil
        suiteName = nil
        super.tearDown()
    }

    func testDefaultsAndLegacyThemeCompatibility() {
        let fresh = AppSettings(userDefaults: defaults)
        XCTAssertEqual(fresh.themeMode, .system)
        XCTAssertEqual(fresh.language, .system)
        XCTAssertTrue(fresh.autoSendEnabled)
        XCTAssertEqual(fresh.autoSendDelayMs, 1500)
        XCTAssertTrue(fresh.hapticFeedbackEnabled)
        XCTAssertTrue(fresh.keepScreenAwake)
        XCTAssertTrue(fresh.showInputTips)
        XCTAssertEqual(fresh.sendTimeoutMs, 5000.0)

        // Legacy light/dark values remain preserved.
        defaults.set("dark", forKey: AppSettings.themeModeKey)
        let migrated = AppSettings(userDefaults: defaults)
        XCTAssertEqual(migrated.themeMode, .dark)
        XCTAssertEqual(AppTheme.resolved(mode: .dark, colorScheme: .light), .dark)
        XCTAssertEqual(AppTheme.resolved(mode: .system, colorScheme: .dark), .dark)
        XCTAssertEqual(AppTheme.resolved(mode: .system, colorScheme: .light), .light)
    }

    func testPersistenceAndResetPreservesPairingMetadata() {
        defaults.set(Data("keep".utf8), forKey: ConnectionManager.savedMetadataKey)

        let settings = AppSettings(userDefaults: defaults)
        settings.themeMode = .dark
        settings.language = .en
        settings.autoSendEnabled = false
        settings.autoSendDelayMs = 2500
        settings.hapticFeedbackEnabled = false
        settings.keepScreenAwake = false
        settings.showInputTips = false
        settings.sendTimeoutMs = 10000

        let reloaded = AppSettings(userDefaults: defaults)
        XCTAssertEqual(reloaded.themeMode, .dark)
        XCTAssertEqual(reloaded.language, .en)
        XCTAssertFalse(reloaded.autoSendEnabled)
        XCTAssertEqual(reloaded.autoSendDelayMs, 2500)
        XCTAssertFalse(reloaded.hapticFeedbackEnabled)
        XCTAssertFalse(reloaded.keepScreenAwake)
        XCTAssertFalse(reloaded.showInputTips)
        XCTAssertEqual(reloaded.sendTimeoutMs, 10000)

        reloaded.resetToDefaults()
        XCTAssertEqual(reloaded.themeMode, .system)
        XCTAssertEqual(reloaded.language, .system)
        XCTAssertTrue(reloaded.autoSendEnabled)
        XCTAssertEqual(reloaded.autoSendDelayMs, 1500)
        XCTAssertTrue(reloaded.hapticFeedbackEnabled)
        XCTAssertTrue(reloaded.keepScreenAwake)
        XCTAssertTrue(reloaded.showInputTips)
        XCTAssertEqual(reloaded.sendTimeoutMs, 5000)

        // Saved pairing metadata must not be cleared on settings reset.
        XCTAssertEqual(defaults.data(forKey: ConnectionManager.savedMetadataKey), Data("keep".utf8))
    }

    func testLanguageResolution() {
        XCTAssertTrue(AppLanguage.zh.isChinese(preferredLanguages: ["en-US"]))
        XCTAssertFalse(AppLanguage.en.isChinese(preferredLanguages: ["zh-Hans-CN"]))
        XCTAssertTrue(AppLanguage.system.isChinese(preferredLanguages: ["zh-Hans-CN"]))
        XCTAssertTrue(AppLanguage.system.isChinese(preferredLanguages: ["zh-Hant-TW"]))
        XCTAssertFalse(AppLanguage.system.isChinese(preferredLanguages: ["en-US"]))
        XCTAssertFalse(AppLanguage.system.isChinese(preferredLanguages: ["ja-JP"]))

        XCTAssertEqual(AppStrings.forLanguage(.zh, preferredLanguages: ["en-US"]), .zh)
        XCTAssertEqual(AppStrings.forLanguage(.en, preferredLanguages: ["zh-Hans-CN"]), .en)
    }

    func testInvalidDelayAndTimeoutAreSanitized() {
        let settings = AppSettings(userDefaults: defaults)
        settings.autoSendDelayMs = 999
        XCTAssertEqual(settings.autoSendDelayMs, AppSettings.defaultAutoSendDelayMs)

        settings.sendTimeoutMs = 1234
        XCTAssertEqual(settings.sendTimeoutMs, AppSettings.defaultSendTimeoutMs)
    }
}
