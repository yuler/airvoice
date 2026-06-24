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

    private var webSocketTask: URLSessionWebSocketTask?
    private var session: URLSession?
    private var isDisconnecting = false

    func connect(payload: PairingPayload) {
        disconnect()

        isDisconnecting = false
        state = .connecting
        hostName = nil

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

        let session = URLSession(configuration: .default)
        self.session = session
        let task = session.webSocketTask(with: url)
        webSocketTask = task

        task.resume()
        receiveMessage(task: task)
        sendHello(task: task)
    }

    func disconnect() {
        isDisconnecting = true
        webSocketTask?.cancel(with: .normalClosure, reason: nil)
        webSocketTask = nil
        session?.invalidateAndCancel()
        session = nil
        state = .disconnected
        hostName = nil
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
                    }
                }
            }
        } catch {
            state = .error("Hello encoding failed: \(error.localizedDescription)")
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
