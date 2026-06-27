import UIKit

/// Picks a user-enabled third-party keyboard for voice input.
///
/// iOS does not let host apps force-switch to a named keyboard, but overriding
/// `textInputMode` requests the preferred mode when the field becomes first responder.
enum PreferredKeyboardResolver {
    /// Priority order: Doubao first, then WeChat.
    private static let priorityPrefixes = [
        "doubao",             // 豆包输入法
        "com.tencent.wetype"  // 微信输入法
    ]

    static func preferredInputMode() -> UITextInputMode? {
        let modes = UITextInputMode.activeInputModes
        for prefix in priorityPrefixes {
            if let match = modes.first(where: { mode in
                guard let identifier = mode.keyboardIdentifier else { return false }
                return identifier.lowercased().contains(prefix.lowercased())
            }) {
                return match
            }
        }
        return nil
    }
}

extension UITextInputMode {
    /// Not public API, but the only practical way to distinguish third-party keyboards.
    var keyboardIdentifier: String? {
        value(forKey: "identifier") as? String
    }
}
