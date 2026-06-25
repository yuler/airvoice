import SwiftUI

struct InputMethodTipsView: View {
    let theme: AppTheme
    @Environment(\.openURL) private var openURL

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            Label("语音输入需要第三方输入法", systemImage: "keyboard")
                .font(.caption.weight(.semibold))
                .foregroundStyle(theme.primaryText)

            Text("优先使用豆包或微信输入法，在键盘中点击麦克风说话：")
                .font(.caption)
                .foregroundStyle(theme.secondaryText)
                .fixedSize(horizontal: false, vertical: true)

            HStack(spacing: 10) {
                appStoreLink(title: "豆包输入法", appID: "6752316550")
                appStoreLink(title: "微信输入法", appID: "1618175312")
            }

            Text("安装后前往 设置 → 通用 → 键盘 添加新键盘")
                .font(.caption2)
                .foregroundStyle(theme.secondaryText)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(14)
        .background(theme.tipsBackground)
        .cornerRadius(12)
    }

    private func appStoreLink(title: String, appID: String) -> some View {
        Button {
            if let url = URL(string: "https://apps.apple.com/app/id\(appID)") {
                openURL(url)
            }
        } label: {
            HStack(spacing: 4) {
                Image(systemName: "arrow.up.right.square")
                    .font(.caption2)
                Text(title)
                    .font(.caption.weight(.medium))
            }
            .foregroundStyle(theme.accent)
            .padding(.horizontal, 10)
            .padding(.vertical, 6)
            .background(theme.accent.opacity(theme == .light ? 0.1 : 0.15))
            .cornerRadius(8)
        }
    }
}
