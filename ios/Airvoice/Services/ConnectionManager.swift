import Foundation
import UIKit
import Combine

@MainActor
class ConnectionManager: ObservableObject {
    enum ConnectionState: Equatable {
        case disconnected
        case connecting
        case connected
        case error(String)
    }

    static let savedPayloadKey = "savedPairingPayload"

    @Published var state: ConnectionState = .disconnected
    @Published var hostName: String? = nil

    var onAck: ((String, Bool, String?) -> Void)?
    var onTransportError: ((String) -> Void)?
    var onAuthExpired: (() -> Void)?

    /// The most recent pairing payload, persisted to UserDefaults for reconnection across launches.
    private(set) var lastPayload: PairingPayload?

    private let userDefaults: UserDefaults
    private let initialBackoffNanoseconds: UInt64
    private let maxBackoffNanoseconds: UInt64

    private var webSocketTask: URLSessionWebSocketTask?
    private var session: URLSession?
    private var delegateProxy: WebSocketDelegate?
    private var reconnectTask: Task<Void, Never>?
    private var backoffNanoseconds: UInt64
    private var isDisconnecting = false
    private var isPaused = false

    /// Test hooks to intercept actual socket creation/sends if desired; nil uses URLSession.
    var connectHandlerForTesting: ((PairingPayload) -> Void)?
    var sendTextHandlerForTesting: ((String, String) -> Bool)?

    init(
        userDefaults: UserDefaults = .standard,
        initialBackoffNanoseconds: UInt64 = 2_000_000_000,
        maxBackoffNanoseconds: UInt64 = 30_000_000_000
    ) {
        self.userDefaults = userDefaults
        self.initialBackoffNanoseconds = initialBackoffNanoseconds
        self.maxBackoffNanoseconds = maxBackoffNanoseconds
        self.backoffNanoseconds = initialBackoffNanoseconds
        self.lastPayload = Self.loadSavedPayload(from: userDefaults)
    }

    /// Auto-connect using the previously saved pairing payload on app launch.
    func connectSavedIfAvailable() {
        guard state == .disconnected, let payload = lastPayload else { return }
        backoffNanoseconds = initialBackoffNanoseconds
        isDisconnecting = false
        isPaused = false
        state = .connecting
        hostName = nil
        connectWithPayload(payload)
    }

    func connect(payload: PairingPayload) {
        cancelActiveSocket()
        cancelReconnectTimer()

        lastPayload = payload
        savePayload(payload)
        backoffNanoseconds = initialBackoffNanoseconds
        isDisconnecting = false
        isPaused = false
        state = .connecting
        hostName = nil

        connectWithPayload(payload)
    }

    /// Reconnect using the stored payload. No-op if no payload or already connected/connecting.
    func reconnect() {
        guard let payload = lastPayload else { return }
        guard state != .connected, state != .connecting else { return }

        cancelReconnectTimer()
        isDisconnecting = false
        state = .connecting
        hostName = nil

        connectWithPayload(payload)
    }

    /// Pause scheduled reconnect attempts while the app is in the background.
    func pauseReconnect() {
        isPaused = true
        cancelReconnectTimer()
    }

    /// Resume reconnection when returning to foreground, resetting backoff and retrying immediately.
    func resumeReconnect() {
        isPaused = false
        guard canReconnect else { return }
        guard state != .connected, state != .connecting else { return }
        backoffNanoseconds = initialBackoffNanoseconds
        reconnect()
    }

    func disconnect() {
        isDisconnecting = true
        cancelReconnectTimer()
        cancelActiveSocket()
        state = .disconnected
        hostName = nil
    }

    /// Clear stored payload (e.g. on 401 when the host refreshed its pairing token).
    func clearStoredPayload() {
        lastPayload = nil
        userDefaults.removeObject(forKey: Self.savedPayloadKey)
    }

    /// Whether we have credentials to attempt reconnection.
    var canReconnect: Bool {
        lastPayload != nil
    }

    /// Whether a reconnect timer is currently scheduled.
    var isReconnectScheduled: Bool {
        reconnectTask != nil
    }

    /// Current backoff delay in nanoseconds (exposed for testing).
    var currentBackoffNanoseconds: UInt64 {
        backoffNanoseconds
    }

    @discardableResult
    func sendText(id: String, content: String) -> Bool {
        if let handler = sendTextHandlerForTesting {
            guard state == .connected else { return false }
            return handler(id, content)
        }
        guard state == .connected, let task = webSocketTask else {
            return false
        }

        let outbound = OutboundText(
            id: id,
            content: content,
            ts: Int(Date().timeIntervalSince1970)
        )

        do {
            let data = try JSONEncoder().encode(outbound)
            guard let jsonString = String(data: data, encoding: .utf8) else {
                return false
            }
            task.send(.string(jsonString)) { [weak self] error in
                if let error {
                    Task { @MainActor [weak self] in
                        guard let self else { return }
                        self.handleConnectionFailure(
                            for: task,
                            statusCode: Self.httpStatusCode(from: task.response),
                            error: error,
                            message: "Send failed: \(error.localizedDescription)"
                        )
                    }
                }
            }
            return true
        } catch {
            onTransportError?("Encoding failed: \(error.localizedDescription)")
            return false
        }
    }

    // MARK: - Persistence

    static func loadSavedPayload(from defaults: UserDefaults) -> PairingPayload? {
        guard let data = defaults.data(forKey: savedPayloadKey) else {
            return nil
        }
        return try? JSONDecoder().decode(PairingPayload.self, from: data)
    }

    private func savePayload(_ payload: PairingPayload) {
        if let data = try? JSONEncoder().encode(payload) {
            userDefaults.set(data, forKey: Self.savedPayloadKey)
        }
    }

    // MARK: - Failure & Reconnect Handling

    static func httpStatusCode(from response: URLResponse?) -> Int? {
        (response as? HTTPURLResponse)?.statusCode
    }

    static func isUnauthorized(statusCode: Int?, error: Error? = nil) -> Bool {
        if statusCode == 401 {
            return true
        }
        if let urlError = error as? URLError, urlError.code == .userAuthenticationRequired {
            return true
        }
        return false
    }

    func handleConnectionFailure(
        for task: URLSessionWebSocketTask?,
        statusCode: Int?,
        error: Error? = nil,
        message: String
    ) {
        if let task {
            guard self.webSocketTask === task else { return }
        }
        guard !isDisconnecting else { return }

        cancelActiveSocket()
        hostName = nil

        if Self.isUnauthorized(statusCode: statusCode, error: error) {
            cancelReconnectTimer()
            backoffNanoseconds = initialBackoffNanoseconds
            clearStoredPayload()
            state = .disconnected
            onAuthExpired?()
            return
        }

        state = .error(message)
        onTransportError?(message)
        scheduleReconnect()
    }

    private func handleWebSocketOpened(for task: URLSessionWebSocketTask) {
        guard self.webSocketTask === task, !isDisconnecting else { return }
        backoffNanoseconds = initialBackoffNanoseconds
        sendHello(task: task)
    }

    private func handleWebSocketClosed(for task: URLSessionWebSocketTask) {
        guard self.webSocketTask === task, !isDisconnecting else { return }
        cancelActiveSocket()
        hostName = nil
        state = .disconnected
        onTransportError?("Connection closed")
        scheduleReconnect()
    }

    private func scheduleReconnect() {
        guard !isDisconnecting, !isPaused, lastPayload != nil else { return }

        cancelReconnectTimer()
        let delay = backoffNanoseconds
        backoffNanoseconds = min(backoffNanoseconds * 2, maxBackoffNanoseconds)

        reconnectTask = Task { @MainActor [weak self] in
            do {
                try await Task.sleep(nanoseconds: delay)
            } catch {
                return
            }
            guard !Task.isCancelled, let self else { return }
            self.reconnectTask = nil
            guard !self.isDisconnecting, !self.isPaused else { return }
            self.reconnect()
        }
    }

    private func cancelReconnectTimer() {
        reconnectTask?.cancel()
        reconnectTask = nil
    }

    private func cancelActiveSocket() {
        webSocketTask?.cancel(with: .normalClosure, reason: nil)
        webSocketTask = nil
        session?.invalidateAndCancel()
        session = nil
        delegateProxy = nil
    }

    // MARK: - Private Networking

    private func connectWithPayload(_ payload: PairingPayload) {
        cancelActiveSocket()

        if let handler = connectHandlerForTesting {
            handler(payload)
            return
        }

        guard var components = URLComponents(string: payload.ws) else {
            state = .error("Invalid WebSocket URL")
            return
        }

        var queryItems = components.queryItems ?? []
        queryItems.append(URLQueryItem(name: "token", value: payload.token))
        components.queryItems = queryItems

        guard let url = components.url else {
            state = .error("Invalid URL generated")
            return
        }

        let proxy = WebSocketDelegate(owner: self)
        self.delegateProxy = proxy
        let session = URLSession(configuration: .default, delegate: proxy, delegateQueue: nil)
        self.session = session
        let task = session.webSocketTask(with: url)
        self.webSocketTask = task

        task.resume()
        receiveMessage(task: task)
    }

    private func sendHello(task: URLSessionWebSocketTask) {
        let device = UIDevice.current.name
        let appVersion = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "0.0.0"
        let hello = OutboundHello(device: device, app: appVersion)

        do {
            let data = try JSONEncoder().encode(hello)
            guard let jsonString = String(data: data, encoding: .utf8) else {
                return
            }
            task.send(.string(jsonString)) { [weak self] error in
                if let error {
                    Task { @MainActor [weak self] in
                        guard let self else { return }
                        self.handleConnectionFailure(
                            for: task,
                            statusCode: Self.httpStatusCode(from: task.response),
                            error: error,
                            message: "Hello failed: \(error.localizedDescription)"
                        )
                    }
                }
            }
        } catch {
            handleConnectionFailure(
                for: task,
                statusCode: nil,
                message: "Hello encoding failed: \(error.localizedDescription)"
            )
        }
    }

    private func receiveMessage(task: URLSessionWebSocketTask) {
        task.receive { [weak self] result in
            Task { @MainActor [weak self] in
                guard let self, self.webSocketTask === task else { return }

                switch result {
                case .success(let message):
                    switch message {
                    case .string(let text):
                        self.processIncomingText(text)
                    case .data(let data):
                        if let text = String(data: data, encoding: .utf8) {
                            self.processIncomingText(text)
                        }
                    @unknown default:
                        break
                    }
                    self.receiveMessage(task: task)

                case .failure(let error):
                    self.handleConnectionFailure(
                        for: task,
                        statusCode: Self.httpStatusCode(from: task.response),
                        error: error,
                        message: "Connection lost: \(error.localizedDescription)"
                    )
                }
            }
        }
    }

    private func processIncomingText(_ text: String) {
        guard let data = text.data(using: .utf8) else { return }

        do {
            let msg = try JSONDecoder().decode(InboundMessage.self, from: data)
            switch msg.type {
            case "hello":
                backoffNanoseconds = initialBackoffNanoseconds
                state = .connected
                hostName = msg.host ?? "Unknown Server"
            case "ack":
                if let id = msg.id {
                    onAck?(id, msg.ok ?? false, msg.message)
                }
            case "pong":
                break
            default:
                break
            }
        } catch {
            print("[airvoice] decode error: \(error) raw=\(text)")
        }
    }

    private final class WebSocketDelegate: NSObject, URLSessionWebSocketDelegate {
        weak var owner: ConnectionManager?

        init(owner: ConnectionManager) {
            self.owner = owner
        }

        func urlSession(
            _ session: URLSession,
            webSocketTask: URLSessionWebSocketTask,
            didOpenWithProtocol protocol: String?
        ) {
            let owner = self.owner
            Task { @MainActor [weak owner] in
                owner?.handleWebSocketOpened(for: webSocketTask)
            }
        }

        func urlSession(
            _ session: URLSession,
            webSocketTask: URLSessionWebSocketTask,
            didCloseWith closeCode: URLSessionWebSocketTask.CloseCode,
            reason: Data?
        ) {
            let owner = self.owner
            Task { @MainActor [weak owner] in
                owner?.handleWebSocketClosed(for: webSocketTask)
            }
        }

        func urlSession(
            _ session: URLSession,
            task: URLSessionTask,
            didCompleteWithError error: Error?
        ) {
            guard let wsTask = task as? URLSessionWebSocketTask else { return }
            let statusCode = (task.response as? HTTPURLResponse)?.statusCode
            let message = error?.localizedDescription ?? "Connection closed"
            let owner = self.owner
            Task { @MainActor [weak owner] in
                owner?.handleConnectionFailure(
                    for: wsTask,
                    statusCode: statusCode,
                    error: error,
                    message: message
                )
            }
        }
    }
}
