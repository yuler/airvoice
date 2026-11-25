import Foundation
import Darwin

enum PairingValidationError: LocalizedError, Equatable {
    case unsupportedVersion(Int)
    case emptyToken
    case invalidURL
    case unsupportedScheme(String)
    case cleartextRemoteNotAllowed(host: String)

    var errorDescription: String? {
        switch self {
        case .unsupportedVersion:
            return "不支持的二维码协议版本"
        case .emptyToken:
            return "二维码缺少认证令牌"
        case .invalidURL:
            return "无效的服务器地址"
        case .unsupportedScheme:
            return "仅支持 ws:// 或 wss:// 协议"
        case .cleartextRemoteNotAllowed:
            return "拒绝连接：非局域网地址不支持明文 ws://"
        }
    }
}

struct ValidatedPairing: Equatable {
    let payload: PairingPayload
    let url: URL
    let host: String
    let isLocal: Bool
    let isSecure: Bool
}

struct PairingPayload: Codable, Equatable {
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
        let payload = try JSONDecoder().decode(PairingPayload.self, from: data)
        _ = try payload.validate()
        return payload
    }

    func validate() throws -> ValidatedPairing {
        guard v == 1 else {
            throw PairingValidationError.unsupportedVersion(v)
        }

        let trimmedToken = token.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedToken.isEmpty else {
            throw PairingValidationError.emptyToken
        }

        let trimmedWS = ws.trimmingCharacters(in: .whitespacesAndNewlines)
        guard let components = URLComponents(string: trimmedWS),
              let rawScheme = components.scheme?.lowercased(),
              !rawScheme.isEmpty else {
            throw PairingValidationError.invalidURL
        }

        guard rawScheme == "ws" || rawScheme == "wss" else {
            throw PairingValidationError.unsupportedScheme(rawScheme)
        }

        guard components.user == nil, components.password == nil else {
            throw PairingValidationError.invalidURL
        }

        guard let host = components.host, !host.isEmpty else {
            throw PairingValidationError.invalidURL
        }

        guard let url = components.url else {
            throw PairingValidationError.invalidURL
        }

        let isLocal = Self.isLocalHost(host)
        let isSecure = (rawScheme == "wss")
        if !isSecure && !isLocal {
            throw PairingValidationError.cleartextRemoteNotAllowed(host: host)
        }

        return ValidatedPairing(
            payload: self,
            url: url,
            host: host,
            isLocal: isLocal,
            isSecure: isSecure
        )
    }

    static func isLocalHost(_ rawHost: String) -> Bool {
        var host = rawHost.trimmingCharacters(in: .whitespacesAndNewlines)
        if host.hasPrefix("[") && host.hasSuffix("]") && host.count >= 2 {
            host = String(host.dropFirst().dropLast())
        }
        if let zoneIndex = host.firstIndex(of: "%") {
            host = String(host[..<zoneIndex])
        }
        host = host.lowercased()
        while host.hasSuffix(".") {
            host.removeLast()
        }
        guard !host.isEmpty else { return false }

        if host == "localhost"
            || host.hasSuffix(".localhost")
            || host.hasSuffix(".local")
            || host.hasSuffix(".lan")
            || host.hasSuffix(".home.arpa") {
            return true
        }

        var ipv4 = in_addr()
        if inet_pton(AF_INET, host, &ipv4) == 1 {
            let addr = UInt32(bigEndian: ipv4.s_addr)
            let b0 = UInt8((addr >> 24) & 0xFF)
            let b1 = UInt8((addr >> 16) & 0xFF)
            let b2 = UInt8((addr >> 8) & 0xFF)
            let b3 = UInt8(addr & 0xFF)
            return isLocalIPv4(b0: b0, b1: b1, b2: b2, b3: b3)
        }

        var ipv6 = in6_addr()
        if inet_pton(AF_INET6, host, &ipv6) == 1 {
            let bytes = withUnsafeBytes(of: &ipv6) { Array($0) }
            guard bytes.count == 16 else { return false }

            // ::1 loopback
            if bytes[0..<15].allSatisfy({ $0 == 0 }) && bytes[15] == 1 {
                return true
            }
            // fc00::/7 unique local addresses (includes Tailscale fd7a:115c:a1e0::/48)
            if (bytes[0] & 0xFE) == 0xFC {
                return true
            }
            // fe80::/10 link-local
            if bytes[0] == 0xFE && (bytes[1] & 0xC0) == 0x80 {
                return true
            }
            // ::ffff:0:0/96 IPv4-mapped IPv6
            if bytes[0..<10].allSatisfy({ $0 == 0 }) && bytes[10] == 0xFF && bytes[11] == 0xFF {
                return isLocalIPv4(b0: bytes[12], b1: bytes[13], b2: bytes[14], b3: bytes[15])
            }
            return false
        }

        return false
    }

    private static func isLocalIPv4(b0: UInt8, b1: UInt8, b2: UInt8, b3: UInt8) -> Bool {
        if b0 == 127 { return true }                       // 127.0.0.0/8 loopback
        if b0 == 10 { return true }                        // 10.0.0.0/8 RFC 1918
        if b0 == 172 && (16...31).contains(b1) { return true } // 172.16.0.0/12 RFC 1918
        if b0 == 192 && b1 == 168 { return true }          // 192.168.0.0/16 RFC 1918
        if b0 == 169 && b1 == 254 { return true }          // 169.254.0.0/16 link-local
        if b0 == 100 && (64...127).contains(b1) { return true } // 100.64.0.0/10 CGNAT / Tailscale
        return false
    }
}
