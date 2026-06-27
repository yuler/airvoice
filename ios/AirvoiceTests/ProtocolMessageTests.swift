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
