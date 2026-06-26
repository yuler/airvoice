import Foundation

enum SendTrigger {
    case auto
    case manual
}

@MainActor final class AutoSendController: ObservableObject {
    /// Return `true` when a send actually started (connection ready, etc.).
    var onSend: ((String, SendTrigger) -> Bool)?
    /// Auto-send fires only after this many seconds of no new input. Streamed
    /// voice-input characters keep resetting this window (and the countdown
    /// animation) so we only send once dictation goes quiet.
    let autoSendDelay: TimeInterval = 1.5
    private var debounceTask: Task<Void, Never>?
    private var lastAcked: String?
    @Published private(set) var inFlight = false
    /// True while the idle countdown toward an auto-send is running.
    @Published private(set) var countdownActive = false
    /// Bumped every time the countdown (re)starts so the UI can restart its animation.
    @Published private(set) var countdownToken = 0

    func textDidChange(_ text: String) {
        debounceTask?.cancel()

        let trimmed = text.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty, !inFlight else {
            stopCountdown()
            return
        }

        startCountdown()
        debounceTask = Task { [weak self] in
            do {
                guard let self else { return }
                try await Task.sleep(nanoseconds: UInt64(self.autoSendDelay * 1_000_000_000))
                guard !Task.isCancelled else { return }
                self.countdownActive = false
                _ = self.attemptSend(text)
            } catch {
                // Task cancelled
            }
        }
    }

    func resetOnFocus() {
        stopCountdown()
    }

    func beginSend() {
        stopCountdown()
        inFlight = true
    }

    func markAcked(_ content: String) {
        lastAcked = content
        inFlight = false
    }

    func clearInFlight() {
        inFlight = false
    }

    /// Manual send for debugging; bypasses dedup when `force` is true.
    @discardableResult
    func sendNow(_ raw: String, force: Bool = false) -> Bool {
        let trimmed = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return false }
        guard !inFlight else { return false }

        if !force {
            let lastAckedTrimmed = lastAcked?.trimmingCharacters(in: .whitespacesAndNewlines)
            guard trimmed != lastAckedTrimmed else { return false }
        }

        stopCountdown()
        return onSend?(raw, .manual) ?? false
    }

    @discardableResult
    func attemptSend(_ raw: String) -> Bool {
        let trimmed = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return false }
        guard !inFlight else { return false }

        let lastAckedTrimmed = lastAcked?.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmed != lastAckedTrimmed else { return false }

        stopCountdown()
        return onSend?(raw, .auto) ?? false
    }

    private func startCountdown() {
        countdownToken &+= 1
        countdownActive = true
    }

    private func stopCountdown() {
        debounceTask?.cancel()
        countdownActive = false
    }
}
