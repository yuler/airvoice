import SwiftUI

enum AppTheme: String, CaseIterable {
    case light
    case dark

    var background: Color {
        switch self {
        case .light: Color(hex: "F5F5F7")
        case .dark: Color(hex: "000000")
        }
    }

    var secondaryBackground: Color {
        switch self {
        case .light: Color.white
        case .dark: Color(hex: "0d0e15")
        }
    }

    var border: Color {
        switch self {
        case .light: Color(hex: "E5E5EA")
        case .dark: Color(hex: "2e2e2e")
        }
    }

    var primaryText: Color {
        switch self {
        case .light: Color(hex: "1C1C1E")
        case .dark: .white
        }
    }

    var secondaryText: Color {
        switch self {
        case .light: Color(hex: "6C6C70")
        case .dark: Color.secondary
        }
    }

    var accent: Color { Color(hex: "006efe") }

    var chipBackground: Color {
        switch self {
        case .light: Color.black.opacity(0.06)
        case .dark: Color.white.opacity(0.1)
        }
    }

    /// Theme toggle stands out more on dark backgrounds.
    var themeToggleBackground: Color {
        switch self {
        case .light: Color.black.opacity(0.06)
        case .dark: Color.white.opacity(0.22)
        }
    }

    var themeToggleForeground: Color {
        switch self {
        case .light: Color(hex: "3C3C43")
        case .dark: .white
        }
    }

    var sendButtonBackground: Color {
        switch self {
        case .light: Color.black.opacity(0.05)
        case .dark: Color.white.opacity(0.12)
        }
    }

    var tipsBackground: Color {
        switch self {
        case .light: Color(hex: "EFEFF4")
        case .dark: Color.white.opacity(0.06)
        }
    }

    var toastBackground: Color {
        switch self {
        case .light: Color(hex: "1C1C1E").opacity(0.92)
        case .dark: Color(hex: "1F2030").opacity(0.95)
        }
    }

    var placeholderText: Color {
        switch self {
        case .light: Color(hex: "AEAEB2")
        case .dark: .gray
        }
    }

    mutating func toggle() {
        self = self == .light ? .dark : .light
    }
}

extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let a, r, g, b: UInt64
        switch hex.count {
        case 3:
            (a, r, g, b) = (255, (int >> 8) * 17, (int >> 4 & 0xF) * 17, (int & 0xF) * 17)
        case 6:
            (a, r, g, b) = (255, int >> 16, int >> 8 & 0xFF, int & 0xFF)
        case 8:
            (a, r, g, b) = (int >> 24, int >> 16 & 0xFF, int >> 8 & 0xFF, int & 0xFF)
        default:
            (a, r, g, b) = (1, 1, 1, 1)
        }

        self.init(
            .sRGB,
            red: Double(r) / 255,
            green: Double(g) / 255,
            blue: Double(b) / 255,
            opacity: Double(a) / 255
        )
    }
}
