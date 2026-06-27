import Foundation

struct OutboundHello: Encodable {
    let type = "hello"
    let device: String
    let app: String
}

struct OutboundText: Encodable {
    let type = "text"
    let id: String
    let content: String
    let ts: Int
}

struct InboundMessage: Decodable {
    let type: String
    let host: String?
    let message: String?
    let id: String?
    let ok: Bool?
}
