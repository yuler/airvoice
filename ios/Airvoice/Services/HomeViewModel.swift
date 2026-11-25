import Foundation
import UIKit

@MainActor
final class HomeViewModel: ObservableObject {
    @Published var text = ""
    @Published var toastMessage: String?
    @Published var isToastError = false
    @Published var sendTimedOut = false

    private let settings: AppSettings
    private var sendTimeoutTask: Task<Void, Never>?
    private var retryTask: Task<Void, Never>?
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

    init(settings: AppSettings? = nil) {
        self.settings = settings ?? AppSettings.shared
    }

    func wire(connection: ConnectionManager, autoSend: AutoSendController) {
        autoSend.currentText = { [weak self] in
            self?.text ?? ""
        }

        autoSend.onSend = { [weak self, weak connection, weak autoSend] content, trigger in
            guard let self, let connection, let autoSend else { return false }
            return self.handleSend(
                content: content,
                trigger: trigger,
                connection: connection,
                autoSend: autoSend
            )
        }

        connection.onAck = { [weak self, weak connection, weak autoSend] id, ok, errMsg in
            Task { @MainActor in
                guard let self, let connection, let autoSend else { return }
                self.handleAck(id: id, ok: ok, errMsg: errMsg, connection: connection, autoSend: autoSend)
            }
        }

        connection.onTransportError = { [weak self, weak autoSend] message in
            Task { @MainActor in
                guard let self, let autoSend else { return }
                self.handleTransportError(message, autoSend: autoSend)
            }
        }

        connection.onAuthExpired = { [weak self, weak autoSend] in
            Task { @MainActor in
                guard let self, let autoSend else { return }
                self.handleAuthExpired(autoSend: autoSend)
            }
        }
    }

    func flushPendingAutoSend(connection: ConnectionManager, autoSend: AutoSendController) {
        guard settings.autoSendEnabled else {
            pendingAutoSendText = nil
            return
        }
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
        let strings = settings.strings
        guard connection.state == .connected else {
            showToast(strings.toastConnectFirst, isError: true)
            return
        }
        guard !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            showToast(strings.toastEnterText, isError: true)
            return
        }
        if autoSend.inFlight {
            showToast(strings.toastSendInFlight, isError: true)
            return
        }
        autoSend.sendNow(text, force: true)
    }

    /// Stop waiting on the in-flight send so the user can keep editing immediately.
    func cancelSend(autoSend: AutoSendController) {
        let hadPendingRetry = retryTask != nil
        retryTask?.cancel()
        retryTask = nil
        guard autoSend.inFlight || hadPendingRetry else { return }
        sendTimeoutTask?.cancel()
        sendTimeoutTask = nil
        if let msgId = pendingSendMsgId {
            clearPendingSend(msgId: msgId)
        }
        inFlightContent = nil
        inFlightIsRetry = false
        autoSend.clearInFlight()
        showToast(settings.strings.toastSendCancelled, isError: false)
    }

    func disconnectAndClear(
        connection: ConnectionManager,
        autoSend: AutoSendController,
        showToast: Bool = true
    ) {
        retryTask?.cancel()
        retryTask = nil
        sendTimeoutTask?.cancel()
        sendTimeoutTask = nil
        if let msgId = pendingSendMsgId {
            clearPendingSend(msgId: msgId)
        }
        pendingAutoSendText = nil
        inFlightContent = nil
        inFlightIsRetry = false
        sendTimedOut = false
        autoSend.resetOnFocus()
        autoSend.clearInFlight()
        autoSend.resetLastAcked()
        text = ""
        connection.disconnectAndClear()
        if showToast {
            self.showToast(settings.strings.toastPairingCleared, isError: false)
        }
    }

    @discardableResult
    private func handleSend(
        content: String,
        trigger: SendTrigger,
        connection: ConnectionManager,
        autoSend: AutoSendController,
        isRetry: Bool = false
    ) -> Bool {
        let strings = settings.strings
        let trimmed = content.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return false }

        if connection.state != .connected {
            if trigger == .auto {
                pendingAutoSendText = content
                return false
            }
            showToast(strings.toastConnectFirst, isError: true)
            return false
        }

        if autoSend.inFlight {
            if trigger == .manual {
                showToast(strings.toastSendInFlight, isError: true)
            }
            return false
        }

        retryTask?.cancel()
        retryTask = nil
        autoSend.beginSend()
        sendTimedOut = false

        inFlightContent = content
        inFlightTrigger = trigger
        inFlightIsRetry = isRetry

        let msgId = UUID().uuidString
        pendingSendMsgId = msgId
        pendingSendContent[msgId] = content
        guard connection.sendText(id: msgId, content: content) else {
            clearPendingSend(msgId: msgId)
            inFlightContent = nil
            inFlightIsRetry = false
            autoSend.clearInFlight()
            showToast(strings.toastSendFailedNotConnected, isError: true)
            return false
        }

        sendTimeoutTask?.cancel()
        sendTimeoutTask = Task { [weak self, weak autoSend] in
            guard let self else { return }
            let timeoutMs = self.settings.sendTimeoutMs
            let timeoutNs = UInt64(timeoutMs * 1_000_000)
            try? await Task.sleep(nanoseconds: timeoutNs)
            guard !Task.isCancelled else {
                self.sendTimeoutTask = nil
                return
            }
            guard let autoSend else { return }
            if autoSend.inFlight {
                if let msgId = self.pendingSendMsgId {
                    self.clearPendingSend(msgId: msgId)
                }
                autoSend.clearInFlight()
                self.inFlightContent = nil
                self.inFlightIsRetry = false
                self.sendTimedOut = true
                self.showToast(self.settings.strings.toastSendTimeout, isError: true)
            }
            self.sendTimeoutTask = nil
        }
        return true
    }

    private func handleTransportError(_ message: String, autoSend: AutoSendController) {
        let wasInFlight = autoSend.inFlight
        retryTask?.cancel()
        retryTask = nil
        sendTimeoutTask?.cancel()
        sendTimeoutTask = nil
        if let msgId = pendingSendMsgId {
            clearPendingSend(msgId: msgId)
        }
        autoSend.clearInFlight()
        inFlightContent = nil
        inFlightIsRetry = false
        // Keep background reconnect failures silent in the status bar; only alert
        // when an active send was interrupted.
        if wasInFlight {
            sendTimedOut = true
            showToast(settings.strings.toastSendFailedCheckConnection, isError: true)
        }
    }

    private func handleAuthExpired(autoSend: AutoSendController) {
        retryTask?.cancel()
        retryTask = nil
        sendTimeoutTask?.cancel()
        sendTimeoutTask = nil
        if let msgId = pendingSendMsgId {
            clearPendingSend(msgId: msgId)
        }
        pendingAutoSendText = nil
        autoSend.clearInFlight()
        inFlightContent = nil
        inFlightIsRetry = false
        sendTimedOut = false
        showToast(settings.strings.toastAuthExpired, isError: true)
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
        // Ignore stale, duplicate, or unknown ACKs (e.g. after timeout or user cancel)
        // so they never cancel an active send's timeout, clear newly typed text, or
        // release a newer in-flight lock.
        guard let sentContent = pendingSendContent.removeValue(forKey: id) else {
            return
        }
        if pendingSendMsgId == id {
            pendingSendMsgId = nil
        }
        sendTimeoutTask?.cancel()
        sendTimeoutTask = nil

        if ok {
            if settings.hapticFeedbackEnabled {
                UIImpactFeedbackGenerator(style: .medium).impactOccurred()
            }
            sendTimedOut = false
            showToast(settings.strings.toastSentToPc, isError: false)
            autoSend.markAcked(sentContent)
            let remaining: String
            if text.hasPrefix(sentContent) {
                remaining = String(text.dropFirst(sentContent.count))
            } else {
                remaining = text
            }
            text = remaining
            autoSend.textDidChange(remaining)
            inFlightContent = nil
            inFlightIsRetry = false
        } else {
            autoSend.clearInFlight()

            // The desktop's first paste right after (re)connecting can transiently
            // fail. Auto-retry an auto-send once before surfacing the error so the
            // user doesn't have to manually re-send.
            let retryContent = inFlightContent ?? sentContent
            let shouldRetry = (inFlightTrigger == .auto) && !inFlightIsRetry
            inFlightContent = nil

            if shouldRetry,
               !retryContent.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                retryTask?.cancel()
                retryTask = Task { @MainActor [weak self] in
                    do {
                        try await Task.sleep(nanoseconds: 400_000_000)
                    } catch {
                        return
                    }
                    guard !Task.isCancelled, let self else { return }
                    self.retryTask = nil
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
                inFlightIsRetry = false
                showToast(errMsg ?? settings.strings.toastSendFailedGeneric, isError: true)
            }
        }
    }
}
