import Foundation
import SwiftUI

/// Manages app-wide configurations and preferences, persisting them via UserDefaults.
final class AppSettings: ObservableObject {
    static let shared = AppSettings()

    /// Timeout in milliseconds for clipboard transmission acknowledgment from CLI.
    @AppStorage("sendTimeoutMs") var sendTimeoutMs: Double = 2000.0

    private init() {}
}
