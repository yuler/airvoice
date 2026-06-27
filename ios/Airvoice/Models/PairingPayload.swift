import Foundation

struct PairingPayload: Codable {
    let v: Int
    let ws: String
    let token: String
    
    static func decode(from string: String) throws -> PairingPayload {
        guard let data = string.data(using: .utf8) else {
            throw DecodingError.dataCorrupted(DecodingError.Context(
                codingPath: [],
                debugDescription: "Failed to convert string to UTF-8 data"
            ))
        }
        return try JSONDecoder().decode(PairingPayload.self, from: data)
    }
}
