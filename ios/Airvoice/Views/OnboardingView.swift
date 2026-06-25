import SwiftUI

struct OnboardingView: View {
    @AppStorage("hasSeenOnboarding") private var hasSeenOnboarding = false

    var body: some View {
        ZStack {
            // Pure black background
            Color(hex: "000000").ignoresSafeArea()
            
            VStack(spacing: 24) {
                Spacer()
                
                // Icon/App Logo
                Image(systemName: "waveform.circle.fill")
                    .resizable()
                    .scaledToFit()
                    .frame(width: 80, height: 80)
                    .foregroundStyle(Color(hex: "006efe")) // Accent-blue
                    .shadow(color: Color(hex: "006efe").opacity(0.3), radius: 15)
                
                VStack(spacing: 8) {
                    Text("Airvoice")
                        .font(.system(size: 32, weight: .bold, design: .rounded))
                        .foregroundStyle(.white)
                    
                    Text("让手机语音输入无缝连接电脑")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
                
                Spacer()
                
                // Guide steps card
                VStack(alignment: .leading, spacing: 16) {
                    Text("输入法安装与配置指南")
                        .font(.headline)
                        .foregroundStyle(.white)
                        .padding(.bottom, 4)
                    
                    GuideStepView(
                        number: "1",
                        title: "安装推荐输入法",
                        desc: "推荐使用「豆包输入法」（主）或「微信输入法」（备选）"
                    )
                    
                    GuideStepView(
                        number: "2",
                        title: "启用键盘",
                        desc: "前往「系统设置」→「通用」→「键盘」→「添加新键盘」"
                    )
                    
                    GuideStepView(
                        number: "3",
                        title: "隐私安全",
                        desc: "无需启用「允许完全访问」，Airvoice 仅读取本 App 内的输入框"
                    )
                }
                .padding(20)
                .background(
                    RoundedRectangle(cornerRadius: 20)
                        .fill(Color(hex: "0d0e15")) // background-secondary
                        .overlay(
                            RoundedRectangle(cornerRadius: 20)
                                .stroke(Color(hex: "2e2e2e"), lineWidth: 1) // border-default
                        )
                )
                .padding(.horizontal, 24)
                
                Spacer()
                
                // Start Button
                Button(action: {
                    hasSeenOnboarding = true
                }) {
                    Text("开始使用")
                        .font(.headline)
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .frame(height: 56)
                        .background(Color(hex: "006efe")) // accent-blue
                        .cornerRadius(28)
                        .shadow(color: Color(hex: "006efe").opacity(0.3), radius: 10)
                }
                .padding(.horizontal, 24)
                .padding(.bottom, 32)
            }
        }
    }
}

struct GuideStepView: View {
    let number: String
    let title: String
    let desc: String
    
    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            Text(number)
                .font(.system(.body, design: .monospaced))
                .fontWeight(.bold)
                .foregroundStyle(.white)
                .frame(width: 24, height: 24)
                .background(Color(hex: "006efe").opacity(0.2))
                .cornerRadius(12)
                .overlay(
                    Circle().stroke(Color(hex: "006efe").opacity(0.5), lineWidth: 1)
                )
            
            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.subheadline)
                    .fontWeight(.semibold)
                    .foregroundStyle(.white)
                Text(desc)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .lineLimit(nil)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }
}

// Utility extension for hex colors in SwiftUI
extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let a, r, g, b: UInt64
        switch hex.count {
        case 3: // RGB (12-bit)
            (a, r, g, b) = (255, (int >> 8) * 17, (int >> 4 & 0xF) * 17, (int & 0xF) * 17)
        case 6: // RGB (24-bit)
            (a, r, g, b) = (255, int >> 16, int >> 8 & 0xFF, int & 0xFF)
        case 8: // ARGB (32-bit)
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

#Preview {
    OnboardingView()
}
