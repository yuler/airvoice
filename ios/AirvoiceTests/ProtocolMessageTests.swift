import XCTest
@testable import Airvoice

final class ProtocolMessageTests: XCTestCase {
    
    func testPairingPayloadDecoding() throws {
        let jsonString = """
        {
            "v": 1,
            "ws": "ws://192.168.1.100:7383/ws",
            "token": "test-uuid-token"
        }
        """
        
        let payload = try PairingPayload.decode(from: jsonString)
        XCTAssertEqual(payload.v, 1)
        XCTAssertEqual(payload.ws, "ws://192.168.1.100:7383/ws")
        XCTAssertEqual(payload.token, "test-uuid-token")
    }

    func testPairingPayloadValidationAcceptsLocalAddresses() throws {
        let localURLs = [
            "ws://192.168.1.100:7383/ws",
            "ws://10.0.0.5:7655/ws",
            "ws://172.16.0.2:7655/ws",
            "ws://172.31.255.1:7655/ws",
            "ws://127.0.0.1:7655/ws",
            "ws://169.254.1.2:7655/ws",
            "ws://100.64.0.1:7655/ws",
            "ws://100.127.255.254:7655/ws",
            "ws://localhost:7655/ws",
            "ws://my-mac.local:7655/ws",
            "ws://[::1]:7655/ws",
            "ws://[fd7a:115c:a1e0::1]:7655/ws",
            "ws://[fe80::1]:7655/ws",
        ]

        for ws in localURLs {
            let payload = PairingPayload(v: 1, ws: ws, token: "valid-token")
            let validated = try payload.validate()
            XCTAssertTrue(validated.isLocal, "Expected \(ws) to be recognized as local")
            XCTAssertFalse(validated.isSecure)
        }
    }

    func testPairingPayloadValidationRejectsInvalidPayloads() {
        let badVersion = PairingPayload(v: 2, ws: "ws://192.168.1.100:7383/ws", token: "tok")
        XCTAssertThrowsError(try badVersion.validate()) { error in
            XCTAssertEqual(error as? PairingValidationError, .unsupportedVersion(2))
        }

        let emptyToken = PairingPayload(v: 1, ws: "ws://192.168.1.100:7383/ws", token: "   ")
        XCTAssertThrowsError(try emptyToken.validate()) { error in
            XCTAssertEqual(error as? PairingValidationError, .emptyToken)
        }

        let httpScheme = PairingPayload(v: 1, ws: "http://192.168.1.100:7383/ws", token: "tok")
        XCTAssertThrowsError(try httpScheme.validate()) { error in
            XCTAssertEqual(error as? PairingValidationError, .unsupportedScheme("http"))
        }

        let remoteCleartext = [
            ("ws://8.8.8.8:7655/ws", "8.8.8.8"),
            ("ws://evil.example.com/ws", "evil.example.com"),
            ("ws://172.32.0.1:7655/ws", "172.32.0.1"),
            ("ws://100.128.0.1:7655/ws", "100.128.0.1"),
        ]
        for (ws, expectedHost) in remoteCleartext {
            let payload = PairingPayload(v: 1, ws: ws, token: "tok")
            XCTAssertThrowsError(try payload.validate(), "Expected \(ws) to be rejected") { error in
                XCTAssertEqual(error as? PairingValidationError, .cleartextRemoteNotAllowed(host: expectedHost))
            }
        }
    }

    func testPairingPayloadAllowsRemoteWssAndMarksNonLocal() throws {
        let payload = PairingPayload(v: 1, ws: "wss://tunnel.example.com:443/ws", token: "tok")
        let validated = try payload.validate()
        XCTAssertFalse(validated.isLocal)
        XCTAssertTrue(validated.isSecure)
        XCTAssertEqual(validated.host, "tunnel.example.com")
    }
    
    func testOutboundHelloEncoding() throws {
        let hello = OutboundHello(device: "iPhone", app: "0.1.0")
        let data = try JSONEncoder().encode(hello)
        let json = try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: Any])
        
        XCTAssertEqual(json["type"] as? String, "hello")
        XCTAssertEqual(json["device"] as? String, "iPhone")
        XCTAssertEqual(json["app"] as? String, "0.1.0")
    }
    
    func testOutboundTextEncoding() throws {
        let textMsg = OutboundText(id: "msg-123", content: "Hello World\nLine 2", ts: 1700000000)
        let data = try JSONEncoder().encode(textMsg)
        let json = try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: Any])
        
        XCTAssertEqual(json["type"] as? String, "text")
        XCTAssertEqual(json["id"] as? String, "msg-123")
        XCTAssertEqual(json["content"] as? String, "Hello World\nLine 2")
        XCTAssertEqual(json["ts"] as? Int, 1700000000)
    }
    
    func testInboundMessageDecoding() throws {
        let jsonHello = """
        {
            "type": "hello",
            "host": "my-mac",
            "version": "0.1.0"
        }
        """.data(using: .utf8)!
        
        let helloMsg = try JSONDecoder().decode(InboundMessage.self, from: jsonHello)
        XCTAssertEqual(helloMsg.type, "hello")
        XCTAssertEqual(helloMsg.host, "my-mac")
        XCTAssertNil(helloMsg.id)
        
        let jsonAck = """
        {
            "type": "ack",
            "id": "msg-123",
            "ok": true
        }
        """.data(using: .utf8)!
        
        let ackMsg = try JSONDecoder().decode(InboundMessage.self, from: jsonAck)
        XCTAssertEqual(ackMsg.type, "ack")
        XCTAssertEqual(ackMsg.id, "msg-123")
        XCTAssertEqual(ackMsg.ok, true)
        XCTAssertNil(ackMsg.message)
    }
}
