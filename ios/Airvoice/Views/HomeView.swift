import SwiftUI
import Combine

struct HomeView: View {
    @AppStorage("appTheme") private var appThemeRaw = AppTheme.light.rawValue
    @Environment(\.scenePhase) private var scenePhase

    @StateObject private var connection = ConnectionManager()
    @StateObject private var autoSend = AutoSendController()
    @StateObject private var viewModel = HomeViewModel()

    @FocusState private var isEditorFocused: Bool

    @State private var showScanner = false

    private var theme: AppTheme {
        AppTheme(rawValue: appThemeRaw) ?? .light
    }

    var body: some View {
        ZStack {
            theme.background.ignoresSafeArea()

            VStack(spacing: 20) {
                HStack {
                    statusBadge

                    Spacer()

                    themeToggleButton

                    Button(action: {
                        showScanner = true
                    }) {
                        HStack(spacing: 6) {
                            Image(systemName: "qrcode.viewfinder")
                            Text(connection.state == .connected ? "重新配对" : "扫码配对")
                        }
                        .font(.subheadline)
                        .fontWeight(.medium)
                        .foregroundColor(theme.primaryText)
                        .padding(.horizontal, 16)
                        .padding(.vertical, 8)
                        .background(theme.chipBackground)
                        .cornerRadius(20)
                    }
                }
                .padding(.horizontal, 20)
                .padding(.top, 10)

                VStack(spacing: 8) {
                    ZStack(alignment: .topLeading) {
                        if viewModel.text.isEmpty {
                            Text("在此输入，或使用键盘麦克风语音输入...")
                                .foregroundColor(theme.placeholderText)
                                .padding(.horizontal, 16)
                                .padding(.vertical, 12)
                                .allowsHitTesting(false)
                        }

                        TextEditor(text: $viewModel.text)
                            .focused($isEditorFocused)
                            .scrollContentBackground(.hidden)
                            .foregroundColor(theme.primaryText)
                            .font(.body)
                            .padding(8)
                            .onChange(of: viewModel.text) { _, newValue in
                                autoSend.textDidChange(newValue)
                                // Text is cleared after a successful send; keep focus so
                                // the keyboard never drops.
                                if newValue.isEmpty, !showScanner {
                                    isEditorFocused = true
                                }
                            }
                    }
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .background(theme.secondaryBackground)
                    .clipShape(RoundedRectangle(cornerRadius: 16))
                    .overlay(
                        RoundedRectangle(cornerRadius: 16)
                            .stroke(theme.border, lineWidth: 1)
                    )

                    AutoSendCountdownBar(
                        active: autoSend.countdownActive,
                        token: autoSend.countdownToken,
                        duration: autoSend.autoSendDelay
                    )
                }
                .padding(.horizontal, 20)

                VStack(spacing: 12) {
                    if autoSend.inFlight {
                        HStack(spacing: 6) {
                            ProgressView()
                                .controlSize(.small)
                                .tint(theme.secondaryText)
                            Text("发送中")
                                .font(.caption)
                                .foregroundColor(theme.secondaryText)
                            Button {
                                viewModel.cancelSend(autoSend: autoSend)
                            } label: {
                                Text("取消")
                                    .font(.caption.weight(.semibold))
                                    .foregroundColor(theme.accent)
                            }
                        }
                        .padding(.horizontal, 12)
                        .padding(.vertical, 6)
                        .background(theme.chipBackground)
                        .cornerRadius(14)
                    } else if connection.state != .connected {
                        Text("请先扫码连接电脑")
                            .font(.caption)
                            .foregroundColor(theme.secondaryText)
                    }

                    Button(action: {
                        viewModel.manualSend(connection: connection, autoSend: autoSend)
                    }) {
                        HStack(spacing: 8) {
                            Image(systemName: "paperplane.fill")
                            Text("发送到电脑")
                        }
                        .font(.subheadline.weight(.semibold))
                        .foregroundColor(theme.primaryText)
                        .frame(maxWidth: .infinity)
                        .frame(height: 44)
                        .background(
                            theme.sendButtonBackground.opacity(
                                connection.state == .connected ? 1 : 0.5
                            )
                        )
                        .cornerRadius(22)
                    }
                    .disabled(connection.state != .connected || autoSend.inFlight)

                    InputMethodTipsView(theme: theme)
                }
                .padding(.horizontal, 20)
                .padding(.bottom, 20)
            }
        }
        .toast(message: $viewModel.toastMessage, isError: $viewModel.isToastError, theme: theme)
        .sheet(isPresented: $showScanner) {
            scannerSheetView
        }
        .onAppear {
            viewModel.wire(connection: connection, autoSend: autoSend)
            focusEditorForKeyboard()
        }
        .onChange(of: showScanner) { _, isShowing in
            if isShowing {
                isEditorFocused = false
            } else {
                focusEditorForKeyboard()
            }
        }
        .onChange(of: connection.state) { _, newValue in
            UIApplication.shared.isIdleTimerDisabled = (newValue == .connected)
            if newValue == .connected {
                viewModel.flushPendingAutoSend(connection: connection, autoSend: autoSend)
            }
        }
        .onChange(of: scenePhase) { _, newPhase in
            if newPhase == .active, !showScanner {
                focusEditorForKeyboard()
            }
        }
    }

    private func focusEditorForKeyboard() {
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.35) {
            isEditorFocused = true
        }
    }

    private var themeToggleButton: some View {
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
        .padding(.trailing, 8)
    }

    private var statusBadge: some View {
        HStack(spacing: 8) {
            Circle()
                .fill(statusColor)
                .frame(width: 8, height: 8)

            Text(statusText)
                .font(.subheadline)
                .fontWeight(.semibold)
                .foregroundColor(theme.primaryText)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 6)
        .background(statusColor.opacity(0.15))
        .cornerRadius(12)
    }

    private var statusColor: Color {
        switch connection.state {
        case .disconnected: return .gray
        case .connecting: return .yellow
        case .connected: return .green
        case .error: return .red
        }
    }

    private var statusText: String {
        switch connection.state {
        case .disconnected: return "未连接"
        case .connecting: return "连接中..."
        case .connected: return "已连接: \(connection.hostName ?? "电脑")"
        case .error(let msg): return "连接错误: \(msg)"
        }
    }

    private var scannerSheetView: some View {
        NavigationStack {
            QRScannerView(
                onScan: { rawPayload in
                    do {
                        let payload = try PairingPayload.decode(from: rawPayload)
                        connection.connect(payload: payload)
                        showScanner = false
                    } catch {
                        viewModel.showToast("无效的二维码", isError: true)
                    }
                },
                onError: { error in
                    viewModel.showToast("相机错误: \(error.localizedDescription)", isError: true)
                }
            )
            .navigationTitle("扫描电脑端二维码")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("取消") {
                        showScanner = false
                    }
                }
            }
            .ignoresSafeArea(edges: .bottom)
        }
    }
}

/// Thin red bar that drains from full width to empty over `duration`, signalling
/// the idle countdown before an automatic send. Restarts whenever `token` changes.
/// Uses a horizontal scale (clamped 0...1) so it can never overflow the text box.
private struct AutoSendCountdownBar: View {
    let active: Bool
    let token: Int
    let duration: Double

    @State private var progress: CGFloat = 0

    var body: some View {
        Capsule()
            .fill(Color.red)
            .frame(height: 3)
            .scaleEffect(x: min(max(progress, 0), 1), y: 1, anchor: .leading)
            .frame(maxWidth: .infinity, alignment: .leading)
            .opacity(active ? 1 : 0)
            .animation(.easeOut(duration: 0.15), value: active)
            .onChange(of: token) { _, _ in
                progress = 1
                withAnimation(.linear(duration: duration)) {
                    progress = 0
                }
            }
    }
}
