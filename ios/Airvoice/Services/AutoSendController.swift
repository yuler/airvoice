import Foundation

@MainActor final class AutoSendController: ObservableObject {
    var onSend: ((String) -> Void)?
    private let idleSeconds = 1.5
    private var debounceTask: Task<Void, Never>?
    private var lastAcked: String?
    @Published private(set) var inFlight = false

    func textDidChange(_ text: String) {
        debounceTask?.cancel()
        debounceTask = Task {
            do {
                try await Task.sleep(nanoseconds: UInt64(idleSeconds * 1_000_000_000))
                guard !Task.isCancelled else { return }
                await attemptSend(text)
            } catch {
                // Task cancelled
            }
        }
    }

    func keyboardDidHide(currentText: String) {
        debounceTask?.cancel()
        Task {
            await attemptSend(currentText)
        }
    }

    func resetOnFocus() {
        debounceTask?.cancel()
    }

    func markAcked(_ content: String) {
        lastAcked = content
        inFlight = false
    }

    func clearInFlight() {
        inFlight = false
    }

    private func attemptSend(_ raw: String) async {
        let trimmed = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }
        guard !inFlight else { return }
        
        let lastAckedTrimmed = lastAcked?.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmed != lastAckedTrimmed else { return }
        
        inFlight = true
        onSend?(raw)
    }
}
