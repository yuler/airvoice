import SwiftUI
import Combine

struct HomeView: View {
    @StateObject private var connection = ConnectionManager()
    @StateObject private var autoSend = AutoSendController()
    @StateObject private var viewModel = HomeViewModel()

    @FocusState private var isEditorFocused: Bool

    @State private var showScanner = false

    var body: some View {
        ZStack {
            Color(hex: "0D0E15").ignoresSafeArea()

            VStack(spacing: 20) {
                HStack {
                    statusBadge

                    Spacer()

                    Button(action: {
                        showScanner = true
                    }) {
                        HStack(spacing: 6) {
                            Image(systemName: "qrcode.viewfinder")
                            Text(connection.state == .connected ? "重新配对" : "扫码配对")
                        }
                        .font(.subheadline)
                        .fontWeight(.medium)
                        .foregroundColor(.white)
                        .padding(.horizontal, 16)
                        .padding(.vertical, 8)
                        .background(Color.white.opacity(0.1))
                        .cornerRadius(20)
                    }
                }
                .padding(.horizontal, 20)
                .padding(.top, 10)

                ZStack(alignment: .topLeading) {
                    if viewModel.text.isEmpty {
                        Text("点击下方「说话」或直接在此处输入...")
                            .foregroundColor(.gray)
                            .padding(.horizontal, 16)
                            .padding(.vertical, 12)
                            .allowsHitTesting(false)
                    }

                    TextEditor(text: $viewModel.text)
                        .focused($isEditorFocused)
                        .scrollContentBackground(.hidden)
                        .foregroundColor(.white)
                        .font(.body)
                        .padding(8)
                        .onChange(of: viewModel.text) { _, newValue in
                            autoSend.textDidChange(newValue)
                        }
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .background(Color.white.opacity(0.03))
                .cornerRadius(16)
                .overlay(
                    RoundedRectangle(cornerRadius: 16)
                        .stroke(Color.white.opacity(0.05), lineWidth: 1)
                )
                .padding(.horizontal, 20)

                VStack(spacing: 12) {
                    if autoSend.inFlight {
                        HStack(spacing: 8) {
                            ProgressView()
                                .tint(.white)
                            Text("正在发送至电脑...")
                                .font(.subheadline)
                                .foregroundColor(.secondary)
                        }
                    } else {
                        Text(connection.state == .connected ? "使用豆包或微信输入法，点击键盘麦克风开始说话" : "请先扫码连接电脑")
                            .font(.caption)
                            .foregroundColor(.secondary)
                    }

                    Button(action: {
                        autoSend.resetOnFocus()
                        isEditorFocused = true
                    }) {
                        HStack(spacing: 10) {
                            Image(systemName: "mic.fill")
                            Text("说话")
                        }
                        .font(.headline)
                        .foregroundColor(.white)
                        .frame(maxWidth: .infinity)
                        .frame(height: 56)
                        .background(
                            connection.state == .connected ?
                            LinearGradient(colors: [Color(hex: "3B82F6"), Color(hex: "8B5CF6")], startPoint: .leading, endPoint: .trailing) :
                            LinearGradient(colors: [Color.gray.opacity(0.3), Color.gray.opacity(0.3)], startPoint: .leading, endPoint: .trailing)
                        )
                        .cornerRadius(28)
                        .shadow(color: connection.state == .connected ? Color(hex: "3B82F6").opacity(0.3) : Color.clear, radius: 10)
                    }
                    .disabled(connection.state != .connected)

                    Button(action: {
                        viewModel.manualSend(connection: connection, autoSend: autoSend)
                    }) {
                        HStack(spacing: 8) {
                            Image(systemName: "paperplane.fill")
                            Text("发送到电脑")
                        }
                        .font(.subheadline.weight(.semibold))
                        .foregroundColor(.white)
                        .frame(maxWidth: .infinity)
                        .frame(height: 44)
                        .background(Color.white.opacity(connection.state == .connected ? 0.12 : 0.06))
                        .cornerRadius(22)
                    }
                    .disabled(connection.state != .connected || autoSend.inFlight)
                }
                .padding(.horizontal, 20)
                .padding(.bottom, 20)
            }
        }
        .toast(message: $viewModel.toastMessage, isError: $viewModel.isToastError)
        .sheet(isPresented: $showScanner) {
            scannerSheetView
        }
        .onAppear {
            viewModel.wire(connection: connection, autoSend: autoSend)
            viewModel.observeKeyboard(autoSend: autoSend)
        }
        .onChange(of: connection.state) { _, newValue in
            UIApplication.shared.isIdleTimerDisabled = (newValue == .connected)
        }
    }

    private var statusBadge: some View {
        HStack(spacing: 8) {
            Circle()
                .fill(statusColor)
                .frame(width: 8, height: 8)

            Text(statusText)
                .font(.subheadline)
                .fontWeight(.semibold)
                .foregroundColor(.white)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 6)
        .background(statusColor.opacity(0.15))
        .cornerRadius(12)
    }

    private var statusColor: Color {
        switch connection.state {
        case .disconnected: return .gray
        case .connecting: return .orange
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
