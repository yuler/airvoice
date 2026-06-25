import Foundation
import UIKit

@MainActor
final class HomeViewModel: ObservableObject {
    @Published var text = ""
    @Published var toastMessage: String?
    @Published var isToastError = false

    private var sendTimeoutTask: Task<Void, Never>?
    private var keyboardObserver: NSObjectProtocol?
    /// Maps in-flight message IDs to the content actually sent (for dedup after ack).
    private var pendingSendContent: [String: String] = [:]
    private var pendingSendMsgId: String?
    /// Brief pause after keyboard hides so third-party IMEs can commit final transcription.
    private let imeCommitDelayNs: UInt64 = 250_000_000

    func wire(connection: ConnectionManager, autoSend: AutoSendController) {
        autoSend.onSend = { [weak self] content in
            Task { @MainActor in
                self?.handleSend(content: content, connection: connection, autoSend: autoSend)
            }
        }

        connection.onAck = { [weak self] id, ok, errMsg in
            Task { @MainActor in
                self?.handleAck(id: id, ok: ok, errMsg: errMsg, autoSend: autoSend)
            }
        }

        connection.onTransportError = { [weak self] message in
            Task { @MainActor in
                self?.handleTransportError(message, autoSend: autoSend)
            }
        }
    }

    func observeKeyboard(autoSend: AutoSendController) {
        if let keyboardObserver {
            NotificationCenter.default.removeObserver(keyboardObserver)
        }
        keyboardObserver = NotificationCenter.default.addObserver(
            forName: UIResponder.keyboardDidHideNotification,
            object: nil,
            queue: .main
        ) { [weak self] _ in
            Task { @MainActor in
                guard let self else { return }
                try? await Task.sleep(nanoseconds: self.imeCommitDelayNs)
                autoSend.keyboardDidHide(currentText: self.text)
            }
        }
    }

    func showToast(_ message: String, isError: Bool) {
        isToastError = isError
        toastMessage = message
    }

    func manualSend(connection: ConnectionManager, autoSend: AutoSendController) {
        guard connection.state == .connected else {
            showToast("请先连接电脑", isError: true)
            return
        }
        guard !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            showToast("请输入文字", isError: true)
            return
        }
        if autoSend.inFlight {
            showToast("上一条仍在发送中", isError: true)
            return
        }
        autoSend.sendNow(text, force: true)
    }

    private func handleSend(
        content: String,
        connection: ConnectionManager,
        autoSend: AutoSendController
    ) {
        let msgId = UUID().uuidString
        pendingSendMsgId = msgId
        pendingSendContent[msgId] = content
        guard connection.sendText(id: msgId, content: content) else {
            clearPendingSend(msgId: msgId)
            autoSend.clearInFlight()
            showToast("发送失败：未连接到电脑", isError: true)
            return
        }

        sendTimeoutTask?.cancel()
        sendTimeoutTask = Task {
            try? await Task.sleep(nanoseconds: 15_000_000_000)
            guard !Task.isCancelled else {
                sendTimeoutTask = nil
                return
            }
            if autoSend.inFlight {
                if let msgId = pendingSendMsgId {
                    clearPendingSend(msgId: msgId)
                }
                autoSend.clearInFlight()
                showToast("发送超时，请重试", isError: true)
            }
            sendTimeoutTask = nil
        }
    }

    private func handleTransportError(_ message: String, autoSend: AutoSendController) {
        sendTimeoutTask?.cancel()
        sendTimeoutTask = nil
        if let msgId = pendingSendMsgId {
            clearPendingSend(msgId: msgId)
        }
        autoSend.clearInFlight()
        showToast(message, isError: true)
    }

    private func clearPendingSend(msgId: String) {
        pendingSendContent.removeValue(forKey: msgId)
        if pendingSendMsgId == msgId {
            pendingSendMsgId = nil
        }
    }

    private func handleAck(id: String, ok: Bool, errMsg: String?, autoSend: AutoSendController) {
        sendTimeoutTask?.cancel()
        sendTimeoutTask = nil
        let sentContent = pendingSendContent.removeValue(forKey: id)
        if pendingSendMsgId == id {
            pendingSendMsgId = nil
        }

        if ok {
            UIImpactFeedbackGenerator(style: .medium).impactOccurred()
            showToast("已发送到电脑", isError: false)
            if let sentContent {
                autoSend.markAcked(sentContent)
            }
            text = ""
        } else {
            autoSend.clearInFlight()
            showToast(errMsg ?? "发送失败", isError: true)
        }
    }
}
