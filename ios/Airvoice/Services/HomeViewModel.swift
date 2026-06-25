import Foundation
import UIKit

@MainActor
final class HomeViewModel: ObservableObject {
    @Published var text = ""
    @Published var toastMessage: String?
    @Published var isToastError = false

    private var sendTimeoutTask: Task<Void, Never>?
    /// Maps in-flight message IDs to the content actually sent (for dedup after ack).
    private var pendingSendContent: [String: String] = [:]
    private var pendingSendMsgId: String?
    /// Auto-send queued while WebSocket hello handshake is still in progress.
    private var pendingAutoSendText: String?
    /// Details of the current in-flight send, used to auto-retry a failed auto-send
    /// once (the desktop's first paste right after connecting can transiently fail).
    private var inFlightContent: String?
    private var inFlightTrigger: SendTrigger = .manual
    private var inFlightIsRetry = false

    func wire(connection: ConnectionManager, autoSend: AutoSendController) {
        autoSend.onSend = { [weak self] content, trigger in
            self?.handleSend(
                content: content,
                trigger: trigger,
                connection: connection,
                autoSend: autoSend
            ) ?? false
        }

        connection.onAck = { [weak self] id, ok, errMsg in
            Task { @MainActor in
                self?.handleAck(id: id, ok: ok, errMsg: errMsg, connection: connection, autoSend: autoSend)
            }
        }

        connection.onTransportError = { [weak self] message in
            Task { @MainActor in
                self?.handleTransportError(message, autoSend: autoSend)
            }
        }
    }

    func flushPendingAutoSend(connection: ConnectionManager, autoSend: AutoSendController) {
        guard connection.state == .connected, let pending = pendingAutoSendText else { return }
        pendingAutoSendText = nil
        _ = handleSend(
            content: pending,
            trigger: .auto,
            connection: connection,
            autoSend: autoSend
        )
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

    /// Stop waiting on the in-flight send so the user can keep editing immediately.
    func cancelSend(autoSend: AutoSendController) {
        guard autoSend.inFlight else { return }
        sendTimeoutTask?.cancel()
        sendTimeoutTask = nil
        if let msgId = pendingSendMsgId {
            clearPendingSend(msgId: msgId)
        }
        autoSend.clearInFlight()
        showToast("已取消发送", isError: false)
    }

    @discardableResult
    private func handleSend(
        content: String,
        trigger: SendTrigger,
        connection: ConnectionManager,
        autoSend: AutoSendController,
        isRetry: Bool = false
    ) -> Bool {
        let trimmed = content.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return false }

        if connection.state != .connected {
            if trigger == .auto {
                pendingAutoSendText = content
                return false
            }
            showToast("请先连接电脑", isError: true)
            return false
        }

        if autoSend.inFlight {
            if trigger == .manual {
                showToast("上一条仍在发送中", isError: true)
            }
            return false
        }

        autoSend.beginSend()

        inFlightContent = content
        inFlightTrigger = trigger
        inFlightIsRetry = isRetry

        let msgId = UUID().uuidString
        pendingSendMsgId = msgId
        pendingSendContent[msgId] = content
        guard connection.sendText(id: msgId, content: content) else {
            clearPendingSend(msgId: msgId)
            autoSend.clearInFlight()
            showToast("发送失败：未连接到电脑", isError: true)
            return false
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
                inFlightContent = nil
                showToast("发送超时，请重试", isError: true)
            }
            sendTimeoutTask = nil
        }
        return true
    }

    private func handleTransportError(_ message: String, autoSend: AutoSendController) {
        sendTimeoutTask?.cancel()
        sendTimeoutTask = nil
        if let msgId = pendingSendMsgId {
            clearPendingSend(msgId: msgId)
        }
        autoSend.clearInFlight()
        inFlightContent = nil
        showToast(message, isError: true)
    }

    private func clearPendingSend(msgId: String) {
        pendingSendContent.removeValue(forKey: msgId)
        if pendingSendMsgId == msgId {
            pendingSendMsgId = nil
        }
    }

    private func handleAck(
        id: String,
        ok: Bool,
        errMsg: String?,
        connection: ConnectionManager,
        autoSend: AutoSendController
    ) {
        sendTimeoutTask?.cancel()
        sendTimeoutTask = nil
        let sentContent = pendingSendContent.removeValue(forKey: id)
        if pendingSendMsgId == id {
            pendingSendMsgId = nil
        }

        if ok {
            UIImpactFeedbackGenerator(style: .medium).impactOccurred()
            showToast("已发送到电脑", isError: false)
            // Always release the in-flight lock on success, even if the acked id
            // didn't match a tracked message (otherwise "发送中" sticks forever).
            if let sentContent {
                autoSend.markAcked(sentContent)
            } else {
                autoSend.clearInFlight()
            }
            text = ""
            inFlightContent = nil
        } else {
            autoSend.clearInFlight()

            // The desktop's first paste right after (re)connecting can transiently
            // fail. Auto-retry an auto-send once before surfacing the error so the
            // user doesn't have to manually re-send.
            let retryContent = inFlightContent
            let shouldRetry = (inFlightTrigger == .auto) && !inFlightIsRetry
            inFlightContent = nil

            if shouldRetry,
               let retryContent,
               !retryContent.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                Task { @MainActor in
                    try? await Task.sleep(nanoseconds: 400_000_000)
                    guard connection.state == .connected, !autoSend.inFlight else { return }
                    _ = self.handleSend(
                        content: retryContent,
                        trigger: .auto,
                        connection: connection,
                        autoSend: autoSend,
                        isRetry: true
                    )
                }
            } else {
                showToast(errMsg ?? "发送失败", isError: true)
            }
        }
    }
}
