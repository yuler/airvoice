import SwiftUI
import Combine

struct HomeView: View {
    @ObservedObject private var settings = AppSettings.shared
    @Environment(\.colorScheme) private var systemColorScheme
    @Environment(\.scenePhase) private var scenePhase

    @StateObject private var connection = ConnectionManager()
    @StateObject private var autoSend = AutoSendController()
    @StateObject private var viewModel = HomeViewModel()

    @FocusState private var isEditorFocused: Bool

    @State private var showScanner = false
    @State private var showSettings = false
    @State private var pendingRemotePairing: ValidatedPairing?
    @State private var isBreathing = false

    private var theme: AppTheme {
        AppTheme.resolved(mode: settings.themeMode, colorScheme: systemColorScheme)
    }

    private var strings: AppStrings {
        settings.strings
    }

    var body: some View {
        ZStack(alignment: .top) {
            theme.background.ignoresSafeArea()

            VStack(spacing: 0) {
                // — Status bar (narrow, full width, at the very top of safe area)
                statusBar

                // — Countdown bar (full width, no spacing, directly below status bar)
                AutoSendCountdownBar(
                    active: settings.autoSendEnabled && autoSend.countdownActive,
                    token: autoSend.countdownToken,
                    duration: settings.autoSendDelaySeconds,
                    theme: theme
                )

                // — Main content
                VStack(spacing: 20) {
                    editorSection
                    bottomControls
                }
                .padding(.top, 12)
            }
        }
        .preferredColorScheme(settings.themeMode.preferredColorScheme)
        .toast(message: $viewModel.toastMessage, isError: $viewModel.isToastError, theme: theme)
        .sheet(isPresented: $showScanner) {
            scannerSheetView
        }
        .sheet(isPresented: $showSettings) {
            SettingsSheetView(
                settings: settings,
                connection: connection,
                theme: theme,
                onDisconnectAndClear: {
                    viewModel.disconnectAndClear(connection: connection, autoSend: autoSend, showToast: false)
                },
                onSettingsReset: {
                    syncSettingsToControllers()
                }
            )
        }
        .alert(
            strings.remoteAlertTitle,
            isPresented: Binding(
                get: { pendingRemotePairing != nil },
                set: { if !$0 { pendingRemotePairing = nil } }
            ),
            presenting: pendingRemotePairing
        ) { pending in
            Button(strings.cancel, role: .cancel) {
                pendingRemotePairing = nil
            }
            Button(strings.remoteAlertConfirm) {
                autoSend.resetLastAcked()
                connection.connect(payload: pending.payload, allowRemote: true)
                pendingRemotePairing = nil
            }
        } message: { pending in
            Text(strings.remoteAlertBody(pending.host))
        }
        .onAppear {
            viewModel.wire(connection: connection, autoSend: autoSend)
            syncSettingsToControllers()
            connection.connectSavedIfAvailable()
            focusEditorForKeyboard()
        }
        .onChange(of: settings.autoSendEnabled) { _, enabled in
            autoSend.setAutoSendEnabled(enabled)
        }
        .onChange(of: settings.autoSendDelayMs) { _, _ in
            autoSend.autoSendDelay = settings.autoSendDelaySeconds
        }
        .onChange(of: settings.keepScreenAwake) { _, _ in
            syncIdleTimer()
        }
        .onChange(of: showScanner) { _, isShowing in
            if isShowing {
                isEditorFocused = false
            } else if !showSettings {
                focusEditorForKeyboard()
            }
        }
        .onChange(of: showSettings) { _, isShowing in
            if isShowing {
                isEditorFocused = false
            } else if !showScanner {
                focusEditorForKeyboard()
            }
        }
        .onChange(of: connection.state) { _, newValue in
            syncIdleTimer()
            if newValue == .connected {
                viewModel.flushPendingAutoSend(connection: connection, autoSend: autoSend)
            }
        }
        .onChange(of: scenePhase) { _, newPhase in
            switch newPhase {
            case .active:
                if !showScanner && !showSettings {
                    focusEditorForKeyboard()
                }
                connection.resumeReconnect()
            case .background:
                connection.pauseReconnect()
            default:
                break
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
                    showScanner = true
                } label: {
                    Image(systemName: "qrcode.viewfinder")
                        .font(.system(size: 13, weight: .bold))
                        .foregroundColor(theme.primaryText)
                        .frame(width: 28, height: 28)
                        .background(theme.chipBackground)
                        .clipShape(Circle())
                }

                Button {
                    showSettings = true
                } label: {
                    Image(systemName: "gearshape.fill")
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
        Circle()
            .fill(statusDotColor)
            .frame(width: 8, height: 8)
            .scaleEffect(shouldBlinkStatusDot && isBreathing ? 1.25 : 1.0)
            .opacity(shouldBlinkStatusDot && isBreathing ? 0.4 : 1.0)
            .onAppear {
                syncStatusDotAnimation()
            }
            .onChange(of: connection.state) { _, _ in
                syncStatusDotAnimation()
            }
            .onChange(of: viewModel.sendTimedOut) { _, _ in
                syncStatusDotAnimation()
            }
    }

    /// Gray when offline, green blink when connected, yellow when connecting or timed out, red on error.
    private var statusDotColor: Color {
        switch connection.state {
        case .disconnected:
            return theme.statusBarDisconnected
        case .connected where !viewModel.sendTimedOut:
            return theme.statusBarConnected
        case .connecting, .connected:
            return theme.statusBarConnecting
        case .error:
            return theme.statusBarError
        }
    }

    private var shouldBlinkStatusDot: Bool {
        connection.state == .connected && !viewModel.sendTimedOut
    }

    private func syncStatusDotAnimation() {
        if shouldBlinkStatusDot {
            startBreathingAnimation()
        } else {
            isBreathing = false
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
        case .disconnected: return strings.statusDisconnected
        case .connecting: return strings.statusConnecting
        case .connected: return strings.statusConnected(connection.hostName)
        case .error: return strings.statusError
        }
    }

    // MARK: - Editor

    private var editorSection: some View {
        ZStack(alignment: .topLeading) {
            if viewModel.text.isEmpty {
                Text(strings.editorPlaceholder)
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
                    if newValue.isEmpty, !showScanner, !showSettings {
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
                    Text(strings.sending)
                        .font(.caption)
                        .foregroundColor(theme.secondaryText)
                    Button {
                        viewModel.cancelSend(autoSend: autoSend)
                    } label: {
                        Text(strings.cancel)
                            .font(.caption.weight(.semibold))
                            .foregroundColor(theme.accent)
                    }
                }
                .padding(.horizontal, 12)
                .padding(.vertical, 6)
                .background(theme.chipBackground)
                .cornerRadius(14)
            } else if connection.state != .connected {
                Text(connection.canReconnect ? strings.promptReconnect : strings.promptScanQr)
                    .font(.caption)
                    .foregroundColor(theme.secondaryText)
            }

            Button(action: {
                viewModel.manualSend(connection: connection, autoSend: autoSend)
            }) {
                HStack(spacing: 8) {
                    Image(systemName: "paperplane.fill")
                    Text(strings.sendToPc)
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

            if settings.showInputTips {
                InputMethodTipsView(theme: theme, strings: strings)
            }
        }
        .padding(.horizontal, 20)
        .padding(.bottom, 20)
    }

    // MARK: - Helpers

    private func syncSettingsToControllers() {
        autoSend.setAutoSendEnabled(settings.autoSendEnabled)
        autoSend.autoSendDelay = settings.autoSendDelaySeconds
        syncIdleTimer()
    }

    private func syncIdleTimer() {
        UIApplication.shared.isIdleTimerDisabled = (connection.state == .connected && settings.keepScreenAwake)
    }

    private func focusEditorForKeyboard() {
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.35) {
            if !showScanner && !showSettings {
                isEditorFocused = true
            }
        }
    }

    private var scannerSheetView: some View {
        NavigationStack {
            QRScannerView(
                onScan: { rawPayload in
                    do {
                        let payload = try PairingPayload.decode(from: rawPayload)
                        let validated = try payload.validate()
                        if validated.isLocal {
                            autoSend.resetLastAcked()
                            connection.connect(payload: validated.payload)
                            showScanner = false
                        } else {
                            pendingRemotePairing = validated
                            showScanner = false
                        }
                    } catch let error as PairingValidationError {
                        viewModel.showToast(error.localizedMessage(strings), isError: true)
                    } catch {
                        viewModel.showToast(strings.invalidQrCode, isError: true)
                    }
                },
                onError: { error in
                    viewModel.showToast(strings.cameraError(error.localizedDescription), isError: true)
                }
            )
            .navigationTitle(strings.scannerTitle)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(strings.cancel) {
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
    let theme: AppTheme

    @State private var progress: CGFloat = 0

    var body: some View {
        Rectangle()
            .fill(theme.countdownBar)
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
