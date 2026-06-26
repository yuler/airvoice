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

    @Published var state: ConnectionState = .disconnected
    @Published var hostName: String? = nil

    var onAck: ((String, Bool, String?) -> Void)?
    var onTransportError: ((String) -> Void)?

    /// The most recent pairing payload, kept in memory for reconnection.
    private(set) var lastPayload: PairingPayload?

    private var webSocketTask: URLSessionWebSocketTask?
    private var session: URLSession?
    private var isDisconnecting = false

    /// Whether a reconnect attempt has already been scheduled/in-progress.
    private var isReconnecting = false

    func connect(payload: PairingPayload) {
        disconnect()

        lastPayload = payload
        isDisconnecting = false
        isReconnecting = false
        state = .connecting
        hostName = nil

        connectWithPayload(payload)
    }

    /// Reconnect using the stored payload. No-op if no payload or already connected.
    func reconnect() {
        guard let payload = lastPayload else { return }
        guard state != .connected, state != .connecting else { return }
        guard !isReconnecting else { return }

        isReconnecting = true
        isDisconnecting = false
        state = .connecting
        hostName = nil

        connectWithPayload(payload)
    }

    func disconnect() {
        isDisconnecting = true
        isReconnecting = false
        webSocketTask?.cancel(with: .normalClosure, reason: nil)
        webSocketTask = nil
        session?.invalidateAndCancel()
        session = nil
        state = .disconnected
        hostName = nil
    }

    /// Clear stored payload (e.g. on 401 when CLI has restarted).
    func clearStoredPayload() {
        lastPayload = nil
    }

    /// Whether we have credentials to attempt reconnection.
    var canReconnect: Bool {
        lastPayload != nil
    }

    @discardableResult
    func sendText(id: String, content: String) -> Bool {
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
                        guard let self, !self.isDisconnecting else { return }
                        let message = "Send failed: \(error.localizedDescription)"
                        self.state = .error(message)
                        self.onTransportError?(message)
                    }
                }
            }
            return true
        } catch {
            state = .error("Encoding failed: \(error.localizedDescription)")
            onTransportError?("Encoding failed: \(error.localizedDescription)")
            return false
        }
    }

    // MARK: - Private

    private func connectWithPayload(_ payload: PairingPayload) {
        // Clean up any existing connection first.
        webSocketTask?.cancel(with: .normalClosure, reason: nil)
        webSocketTask = nil
        session?.invalidateAndCancel()
        session = nil

        guard var components = URLComponents(string: payload.ws) else {
            state = .error("Invalid WebSocket URL")
            isReconnecting = false
            return
        }

        var queryItems = components.queryItems ?? []
        queryItems.append(URLQueryItem(name: "token", value: payload.token))
        components.queryItems = queryItems

        guard let url = components.url else {
            state = .error("Invalid URL generated")
            isReconnecting = false
            return
        }

        let session = URLSession(configuration: .default)
        self.session = session
        let task = session.webSocketTask(with: url)
        webSocketTask = task

        task.resume()
        receiveMessage(task: task)
        sendHello(task: task)
    }

    private func sendHello(task: URLSessionWebSocketTask) {
        let device = UIDevice.current.name
        let hello = OutboundHello(device: device, app: "0.1.0")

        do {
            let data = try JSONEncoder().encode(hello)
            guard let jsonString = String(data: data, encoding: .utf8) else {
                return
            }
            task.send(.string(jsonString)) { [weak self] error in
                if let error {
                    Task { @MainActor [weak self] in
                        guard let self, !self.isDisconnecting else { return }
                        self.state = .error("Hello failed: \(error.localizedDescription)")
                        self.isReconnecting = false
                    }
                }
            }
        } catch {
            state = .error("Hello encoding failed: \(error.localizedDescription)")
            isReconnecting = false
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
                    if !self.isDisconnecting {
                        self.state = .error("Connection lost: \(error.localizedDescription)")
                        self.hostName = nil
                        self.isReconnecting = false
                        self.onTransportError?("Connection lost: \(error.localizedDescription)")
                    }
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
                state = .connected
                hostName = msg.host ?? "Unknown Server"
                isReconnecting = false
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
}
