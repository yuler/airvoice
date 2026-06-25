import SwiftUI

struct OnboardingView: View {
    @AppStorage("hasSeenOnboarding") private var hasSeenOnboarding = false
    @AppStorage("appTheme") private var appThemeRaw = AppTheme.light.rawValue

    private var theme: AppTheme {
        AppTheme(rawValue: appThemeRaw) ?? .light
    }

    var body: some View {
        ZStack {
            theme.background.ignoresSafeArea()

            VStack(spacing: 24) {
                HStack {
                    Spacer()
                    Button {
                        var next = theme
                        next.toggle()
                        appThemeRaw = next.rawValue
                    } label: {
                        Image(systemName: theme == .light ? "moon.fill" : "sun.max.fill")
                            .font(.subheadline.weight(.semibold))
                            .foregroundStyle(theme.themeToggleForeground)
                            .frame(width: 36, height: 36)
                            .background(theme.themeToggleBackground)
                            .cornerRadius(18)
                    }
                }
                .padding(.horizontal, 24)
                .padding(.top, 8)

                Spacer()

                Image(systemName: "waveform.circle.fill")
                    .resizable()
                    .scaledToFit()
                    .frame(width: 80, height: 80)
                    .foregroundStyle(theme.accent)
                    .shadow(color: theme.accent.opacity(0.3), radius: 15)

                VStack(spacing: 8) {
                    Text("Airvoice")
                        .font(.system(size: 32, weight: .bold, design: .rounded))
                        .foregroundStyle(theme.primaryText)

                    Text("让手机语音输入无缝连接电脑")
                        .font(.subheadline)
                        .foregroundStyle(theme.secondaryText)
                }

                Spacer()

                VStack(alignment: .leading, spacing: 16) {
                    Text("输入法安装与配置指南")
                        .font(.headline)
                        .foregroundStyle(theme.primaryText)
                        .padding(.bottom, 4)

                    GuideStepView(
                        number: "1",
                        title: "安装推荐输入法",
                        desc: "推荐使用「豆包输入法」（主）或「微信输入法」（备选）",
                        theme: theme
                    )

                    GuideStepView(
                        number: "2",
                        title: "启用键盘",
                        desc: "前往「系统设置」→「通用」→「键盘」→「添加新键盘」",
                        theme: theme
                    )

                    GuideStepView(
                        number: "3",
                        title: "隐私安全",
                        desc: "无需启用「允许完全访问」，Airvoice 仅读取本 App 内的输入框",
                        theme: theme
                    )
                }
                .padding(20)
                .background(
                    RoundedRectangle(cornerRadius: 20)
                        .fill(theme.secondaryBackground)
                        .overlay(
                            RoundedRectangle(cornerRadius: 20)
                                .stroke(theme.border, lineWidth: 1)
                        )
                )
                .padding(.horizontal, 24)

                Spacer()

                Button(action: {
                    hasSeenOnboarding = true
                }) {
                    Text("开始使用")
                        .font(.headline)
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .frame(height: 56)
                        .background(theme.accent)
                        .cornerRadius(28)
                        .shadow(color: theme.accent.opacity(0.3), radius: 10)
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
    let theme: AppTheme

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            Text(number)
                .font(.system(.body, design: .monospaced))
                .fontWeight(.bold)
                .foregroundStyle(theme.primaryText)
                .frame(width: 24, height: 24)
                .background(theme.accent.opacity(0.2))
                .cornerRadius(12)
                .overlay(
                    Circle().stroke(theme.accent.opacity(0.5), lineWidth: 1)
                )

            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.subheadline)
                    .fontWeight(.semibold)
                    .foregroundStyle(theme.primaryText)
                Text(desc)
                    .font(.caption)
                    .foregroundStyle(theme.secondaryText)
                    .lineLimit(nil)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }
}

#Preview {
    OnboardingView()
}
