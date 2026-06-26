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
    @State private var isBreathing = false

    private var theme: AppTheme {
        AppTheme(rawValue: appThemeRaw) ?? .light
    }

    var body: some View {
        ZStack(alignment: .top) {
            theme.background.ignoresSafeArea()

            VStack(spacing: 0) {
                // — Status bar (narrow, full width, at the very top of safe area)
                statusBar

                // — Countdown bar (full width, no spacing, directly below status bar)
                AutoSendCountdownBar(
                    active: autoSend.countdownActive,
                    token: autoSend.countdownToken,
                    duration: autoSend.autoSendDelay
                )

                // — Main content
                VStack(spacing: 20) {
                    editorSection
                    bottomControls
                }
                .padding(.top, 12)
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
            if newPhase == .active {
                if !showScanner {
                    focusEditorForKeyboard()
                }
                // Auto-reconnect when returning to foreground
                if connection.state != .connected,
                   connection.state != .connecting,
                   connection.canReconnect {
                    connection.reconnect()
                }
            }
        }
    }

    // MARK: - Status Bar

    private var statusBar: some View {
        HStack(spacing: 8) {
            statusIndicator

            Text(statusText)
                .font(.system(size: 13, weight: .medium, design: .rounded))
                .foregroundColor(theme.primaryText.opacity(0.8))

            Spacer()

            HStack(spacing: 8) {
                Button {
                    var next = theme
                    next.toggle()
                    appThemeRaw = next.rawValue
                } label: {
                    Image(systemName: theme == .light ? "moon.fill" : "sun.max.fill")
                        .font(.system(size: 13, weight: .bold))
                        .foregroundColor(theme.primaryText)
                        .frame(width: 28, height: 28)
                        .background(theme.chipBackground)
                        .clipShape(Circle())
                }

                Button {
                    showScanner = true
                } label: {
                    Image(systemName: "qrcode.viewfinder")
                        .font(.system(size: 13, weight: .bold))
                        .foregroundColor(theme.primaryText)
                        .frame(width: 28, height: 28)
                        .background(theme.chipBackground)
                        .clipShape(Circle())
                }
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 8)
        .background(theme.background)
        .animation(.easeInOut(duration: 0.3), value: connection.state)
    }

    @ViewBuilder
    private var statusIndicator: some View {
        if viewModel.sendTimedOut {
            Image(systemName: "exclamationmark.triangle.fill")
                .font(.system(size: 10, weight: .semibold))
                .foregroundColor(theme.statusBarConnecting)
                .frame(width: 8, height: 8)
        } else {
            Circle()
                .fill(stateDotColor)
                .frame(width: 8, height: 8)
                .scaleEffect(shouldBreatheDot && isBreathing ? 1.25 : 1.0)
                .opacity(shouldBreatheDot && isBreathing ? 0.4 : 1.0)
                .onAppear {
                    if shouldBreatheDot {
                        startBreathingAnimation()
                    }
                }
                .onChange(of: connection.state) { _, newState in
                    if newState == .connecting || newState == .connected {
                        startBreathingAnimation()
                    } else {
                        isBreathing = false
                    }
                }
        }
    }

    private var shouldBreatheDot: Bool {
        connection.state == .connecting || connection.state == .connected
    }

    private var stateDotColor: Color {
        switch connection.state {
        case .disconnected: return theme.statusBarDisconnected
        case .connecting: return theme.statusBarConnecting
        case .connected: return theme.statusBarConnected
        case .error: return theme.statusBarError
        }
    }

    private func startBreathingAnimation() {
        withAnimation(
            .easeInOut(duration: 1.0)
            .repeatForever(autoreverses: true)
        ) {
            isBreathing = true
        }
    }

    private var statusText: String {
        switch connection.state {
        case .disconnected: return "未连接"
        case .connecting: return "连接中..."
        case .connected: return "已连接: \(connection.hostName ?? "电脑")"
        case .error(let msg): return msg
        }
    }

    // MARK: - Editor

    private var editorSection: some View {
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
        .padding(.horizontal, 20)
    }

    // MARK: - Bottom Controls

    private var bottomControls: some View {
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
                Text(connection.canReconnect ? "连接中断，正在重连..." : "请先扫码连接电脑")
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

    // MARK: - Helpers

    private func focusEditorForKeyboard() {
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.35) {
            isEditorFocused = true
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

/// Thin bar that drains from full width to empty over `duration`, signalling
/// the idle countdown before an automatic send. Restarts whenever `token` changes.
private struct AutoSendCountdownBar: View {
    let active: Bool
    let token: Int
    let duration: Double

    @State private var progress: CGFloat = 0

    var body: some View {
        Rectangle()
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
