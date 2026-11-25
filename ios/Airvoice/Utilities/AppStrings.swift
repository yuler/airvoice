import Foundation

enum AppLanguage: String, CaseIterable {
    case system
    case zh
    case en

    func isChinese(preferredLanguages: [String] = Locale.preferredLanguages) -> Bool {
        switch self {
        case .zh:
            return true
        case .en:
            return false
        case .system:
            guard let first = preferredLanguages.first?.lowercased() else {
                return false
            }
            return first == "zh" || first.hasPrefix("zh-") || first.hasPrefix("zh_")
        }
    }

    static func fromStorage(_ raw: String?) -> AppLanguage {
        guard let raw, let language = AppLanguage(rawValue: raw) else {
            return .system
        }
        return language
    }
}

struct AppStrings: Equatable {
    let isChinese: Bool

    // Status bar
    let statusDisconnected: String
    let statusConnecting: String
    let statusError: String
    let defaultHostName: String
    let statusConnectedPrefix: String

    // Home screen
    let editorPlaceholder: String
    let sending: String
    let cancel: String
    let promptReconnect: String
    let promptScanQr: String
    let sendToPc: String
    let scannerTitle: String
    let invalidQrCode: String
    let cameraErrorPrefix: String

    // Input method tips
    let tipsTitle: String
    let tipsBody: String
    let tipsInstallHint: String
    let doubaoIme: String
    let wetypeIme: String

    // Remote pairing confirmation
    let remoteAlertTitle: String
    let remoteAlertConfirm: String
    let remoteAlertBodyTemplate: String

    // Pairing validation errors
    let errUnsupportedVersion: String
    let errEmptyToken: String
    let errInvalidUrl: String
    let errUnsupportedScheme: String
    let errCleartextRemote: String

    // Toasts
    let toastAuthExpired: String
    let toastEnterText: String
    let toastConnectFirst: String
    let toastSendInFlight: String
    let toastSendCancelled: String
    let toastSentToPc: String
    let toastSendFailedNotConnected: String
    let toastSendTimeout: String
    let toastSendFailedCheckConnection: String
    let toastSendFailedGeneric: String
    let toastPairingCleared: String
    let toastSettingsReset: String
    let toastVersionCopied: String

    // Settings sheet
    let settingsTitle: String
    let done: String
    let sectionAppearance: String
    let sectionGeneral: String
    let sectionConnection: String
    let sectionAbout: String
    let themeLabel: String
    let themeSystem: String
    let themeLight: String
    let themeDark: String
    let languageLabel: String
    let languageSystem: String
    let languageZh: String
    let languageEn: String
    let autoSendLabel: String
    let autoSendDelayLabel: String
    let hapticFeedbackLabel: String
    let keepScreenAwakeLabel: String
    let showInputTipsLabel: String
    let sendTimeoutLabel: String
    let pairedServerLabel: String
    let notPaired: String
    let disconnectAndClear: String
    let disconnectConfirmTitle: String
    let disconnectConfirmBody: String
    let disconnectConfirmAction: String
    let resetDefaults: String
    let resetConfirmTitle: String
    let resetConfirmBody: String
    let resetConfirmAction: String
    let appVersionLabel: String
    let systemVersionLabel: String
    let tapToCopyHint: String

    func statusConnected(_ host: String?) -> String {
        let trimmed = host?.trimmingCharacters(in: .whitespacesAndNewlines)
        let resolvedHost = (trimmed?.isEmpty == false) ? trimmed! : defaultHostName
        return "\(statusConnectedPrefix): \(resolvedHost)"
    }

    func remoteAlertBody(_ host: String) -> String {
        remoteAlertBodyTemplate.replacingOccurrences(of: "%HOST%", with: host)
    }

    func cameraError(_ description: String) -> String {
        "\(cameraErrorPrefix): \(description)"
    }

    static func forLanguage(
        _ language: AppLanguage,
        preferredLanguages: [String] = Locale.preferredLanguages
    ) -> AppStrings {
        language.isChinese(preferredLanguages: preferredLanguages) ? .zh : .en
    }

    static let zh = AppStrings(
        isChinese: true,
        statusDisconnected: "未连接",
        statusConnecting: "连接中...",
        statusError: "连接失败，正在重试",
        defaultHostName: "电脑",
        statusConnectedPrefix: "已连接",
        editorPlaceholder: "在此输入，或使用键盘麦克风语音输入...",
        sending: "发送中",
        cancel: "取消",
        promptReconnect: "连接中断，正在重连...",
        promptScanQr: "请先扫码连接电脑",
        sendToPc: "发送到电脑",
        scannerTitle: "扫描电脑端二维码",
        invalidQrCode: "无效的二维码",
        cameraErrorPrefix: "相机错误",
        tipsTitle: "语音输入需要第三方输入法",
        tipsBody: "优先使用豆包或微信输入法，在键盘中点击麦克风说话：",
        tipsInstallHint: "安装后前往 设置 → 通用 → 键盘 添加新键盘",
        doubaoIme: "豆包输入法",
        wetypeIme: "微信输入法",
        remoteAlertTitle: "确认连接外部服务器",
        remoteAlertConfirm: "确认连接",
        remoteAlertBodyTemplate: "该二维码指向非局域网地址：\n%HOST%\n\n确认连接后，您的输入将自动发送到该服务器。请仅在信任该服务器时继续。",
        errUnsupportedVersion: "不支持的二维码协议版本",
        errEmptyToken: "二维码缺少认证令牌",
        errInvalidUrl: "无效的服务器地址",
        errUnsupportedScheme: "仅支持 ws:// 或 wss:// 协议",
        errCleartextRemote: "拒绝连接：非局域网地址不支持明文 ws://",
        toastAuthExpired: "配对已失效，请重新扫码",
        toastEnterText: "请输入文字",
        toastConnectFirst: "请先连接电脑",
        toastSendInFlight: "上一条仍在发送中",
        toastSendCancelled: "已取消发送",
        toastSentToPc: "已发送到电脑",
        toastSendFailedNotConnected: "发送失败：未连接到电脑",
        toastSendTimeout: "发送超时，请重试",
        toastSendFailedCheckConnection: "发送失败，请检查连接",
        toastSendFailedGeneric: "发送失败",
        toastPairingCleared: "已断开并清除配对",
        toastSettingsReset: "已恢复默认设置",
        toastVersionCopied: "已复制版本信息",
        settingsTitle: "设置",
        done: "完成",
        sectionAppearance: "外观与语言",
        sectionGeneral: "常规",
        sectionConnection: "连接与数据",
        sectionAbout: "关于",
        themeLabel: "主题",
        themeSystem: "跟随系统",
        themeLight: "浅色",
        themeDark: "深色",
        languageLabel: "语言",
        languageSystem: "跟随系统",
        languageZh: "简体中文",
        languageEn: "English",
        autoSendLabel: "自动发送",
        autoSendDelayLabel: "停顿延迟",
        hapticFeedbackLabel: "发送震动反馈",
        keepScreenAwakeLabel: "连接时保持屏幕常亮",
        showInputTipsLabel: "显示输入法提示卡片",
        sendTimeoutLabel: "发送超时时间",
        pairedServerLabel: "已配对服务器",
        notPaired: "未配对",
        disconnectAndClear: "断开并清除配对",
        disconnectConfirmTitle: "清除已配对设备",
        disconnectConfirmBody: "将断开当前连接并删除本地保存的配对凭证，后续需重新扫码连接。",
        disconnectConfirmAction: "清除配对",
        resetDefaults: "恢复默认设置",
        resetConfirmTitle: "恢复默认设置",
        resetConfirmBody: "外观、语言与常规选项将恢复为默认值，已配对的设备信息不会被清除。",
        resetConfirmAction: "恢复默认",
        appVersionLabel: "应用版本",
        systemVersionLabel: "系统版本",
        tapToCopyHint: "点击复制版本信息"
    )

    static let en = AppStrings(
        isChinese: false,
        statusDisconnected: "Disconnected",
        statusConnecting: "Connecting...",
        statusError: "Connection failed, retrying...",
        defaultHostName: "Computer",
        statusConnectedPrefix: "Connected",
        editorPlaceholder: "Type here, or tap the keyboard mic for voice input...",
        sending: "Sending...",
        cancel: "Cancel",
        promptReconnect: "Connection lost, reconnecting...",
        promptScanQr: "Scan QR code to connect",
        sendToPc: "Send to PC",
        scannerTitle: "Scan Desktop QR Code",
        invalidQrCode: "Invalid QR code",
        cameraErrorPrefix: "Camera error",
        tipsTitle: "Voice Input Tip",
        tipsBody: "Tap the microphone key on your keyboard to dictate directly into this box using Apple Dictation or your preferred voice keyboard.",
        tipsInstallHint: "Enable dictation in Settings → General → Keyboard.",
        doubaoIme: "Doubao IME",
        wetypeIme: "WeType",
        remoteAlertTitle: "Connect to External Server",
        remoteAlertConfirm: "Connect",
        remoteAlertBodyTemplate: "This QR code points to a non-LAN address:\n%HOST%\n\nOnce connected, your input will be sent to this server. Continue only if you trust it.",
        errUnsupportedVersion: "Unsupported QR protocol version",
        errEmptyToken: "QR code is missing authentication token",
        errInvalidUrl: "Invalid server address",
        errUnsupportedScheme: "Only ws:// and wss:// are supported",
        errCleartextRemote: "Blocked: cleartext ws:// is not allowed for remote hosts",
        toastAuthExpired: "Pairing expired, scan QR again",
        toastEnterText: "Enter text first",
        toastConnectFirst: "Connect to PC first",
        toastSendInFlight: "Previous message still sending",
        toastSendCancelled: "Send cancelled",
        toastSentToPc: "Sent to PC",
        toastSendFailedNotConnected: "Send failed: not connected",
        toastSendTimeout: "Send timed out, try again",
        toastSendFailedCheckConnection: "Send failed, check connection",
        toastSendFailedGeneric: "Send failed",
        toastPairingCleared: "Disconnected and cleared pairing",
        toastSettingsReset: "Settings reset to defaults",
        toastVersionCopied: "Version info copied",
        settingsTitle: "Settings",
        done: "Done",
        sectionAppearance: "Appearance & Language",
        sectionGeneral: "General",
        sectionConnection: "Connection & Data",
        sectionAbout: "About",
        themeLabel: "Theme",
        themeSystem: "System",
        themeLight: "Light",
        themeDark: "Dark",
        languageLabel: "Language",
        languageSystem: "System",
        languageZh: "简体中文",
        languageEn: "English",
        autoSendLabel: "Auto-send",
        autoSendDelayLabel: "Pause delay",
        hapticFeedbackLabel: "Haptic feedback",
        keepScreenAwakeLabel: "Keep screen awake",
        showInputTipsLabel: "Show input method tip",
        sendTimeoutLabel: "Send timeout",
        pairedServerLabel: "Paired server",
        notPaired: "Not paired",
        disconnectAndClear: "Disconnect & Clear Pairing",
        disconnectConfirmTitle: "Clear Paired Device",
        disconnectConfirmBody: "This disconnects the current session and removes saved pairing credentials. You will need to scan the QR code again.",
        disconnectConfirmAction: "Clear Pairing",
        resetDefaults: "Reset to Defaults",
        resetConfirmTitle: "Reset to Defaults",
        resetConfirmBody: "Appearance, language, and general preferences will return to defaults. Your paired device will remain saved.",
        resetConfirmAction: "Reset",
        appVersionLabel: "App version",
        systemVersionLabel: "System version",
        tapToCopyHint: "Tap to copy version info"
    )
}
