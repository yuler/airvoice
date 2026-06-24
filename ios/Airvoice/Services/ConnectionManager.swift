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
    
    private var webSocketTask: URLSessionWebSocketTask?
    private var session: URLSession?
    private var isDisconnecting = false
    
    func connect(payload: PairingPayload) {
        disconnect()
        
        isDisconnecting = false
        state = .connecting
        hostName = nil
        
        guard var components = URLComponents(string: payload.ws) else {
            self.state = .error("Invalid WebSocket URL")
            return
        }
        
        var queryItems = components.queryItems ?? []
        queryItems.append(URLQueryItem(name: "token", value: payload.token))
        components.queryItems = queryItems
        
        guard let url = components.url else {
            self.state = .error("Invalid URL generated")
            return
        }
        
        let session = URLSession(configuration: .default)
        self.session = session
        let task = session.webSocketTask(with: url)
        self.webSocketTask = task
        
        task.resume()
        
        // Start receiving messages
        receiveMessage(task: task)
        
        // Send OutboundHello
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
    
    func sendText(id: String, content: String) {
        guard state == .connected, let task = webSocketTask else {
            return
        }
        
        let outbound = OutboundText(
            id: id,
            content: content,
            ts: Int(Date().timeIntervalSince1970)
        )
        
        do {
            let data = try JSONEncoder().encode(outbound)
            guard let jsonString = String(data: data, encoding: .utf8) else {
                return
            }
            task.send(.string(jsonString)) { [weak self] error in
                if let error = error {
                    Task { [weak self] @MainActor in
                        self?.state = .error("Send failed: \(error.localizedDescription)")
                    }
                }
            }
        } catch {
            self.state = .error("Encoding failed: \(error.localizedDescription)")
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
                if let error = error {
                    Task { [weak self] @MainActor in
                        if !(self?.isDisconnecting ?? false) {
                            self?.state = .error("Hello failed: \(error.localizedDescription)")
                        }
                    }
                }
            }
        } catch {
            self.state = .error("Hello encoding failed: \(error.localizedDescription)")
        }
    }
    
    private func receiveMessage(task: URLSessionWebSocketTask) {
        task.receive { [weak self] result in
            guard let self = self else { return }
            
            switch result {
            case .success(let message):
                switch message {
                case .string(let text):
                    self.handleIncomingText(text)
                case .data(let data):
                    if let text = String(data: data, encoding: .utf8) {
                        self.handleIncomingText(text)
                    }
                @unknown default:
                    break
                }
                
                // Continue receiving
                Task { [weak self] @MainActor in
                    guard let self = self else { return }
                    if self.webSocketTask === task {
                        self.receiveMessage(task: task)
                    }
                }
                
            case .failure(let error):
                Task { [weak self] @MainActor in
                    guard let self = self else { return }
                    if !self.isDisconnecting && self.webSocketTask === task {
                        self.state = .error("Connection lost: \(error.localizedDescription)")
                        self.hostName = nil
                    }
                }
            }
        }
    }
    
    private func handleIncomingText(_ text: String) {
        guard let data = text.data(using: .utf8) else { return }
        
        do {
            let msg = try JSONDecoder().decode(InboundMessage.self, from: data)
            
            Task { @MainActor in
                switch msg.type {
                case "hello":
                    self.state = .connected
                    self.hostName = msg.host ?? "Unknown Server"
                case "ack":
                    if let id = msg.id {
                        let ok = msg.ok ?? false
                        self.onAck?(id, ok, msg.message)
                    }
                case "pong":
                    break
                default:
                    break
                }
            }
        } catch {
            print("Failed to decode inbound message: \(error)")
        }
    }
}
