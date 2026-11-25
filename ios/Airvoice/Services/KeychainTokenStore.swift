import Foundation
import Security

protocol TokenStore: AnyObject {
    func loadToken() -> String?
    @discardableResult
    func saveToken(_ token: String) -> Bool
    func deleteToken()
}

final class KeychainTokenStore: TokenStore {
    let service: String
    let account: String

    init(service: String = "cc.yuler.airvoice", account: String = "pairingToken") {
        self.service = service
        self.account = account
    }

    func baseQuery() -> [String: Any] {
        [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
        ]
    }

    func saveQuery(for data: Data) -> [String: Any] {
        var query = baseQuery()
        query[kSecValueData as String] = data
        query[kSecAttrAccessible as String] = kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly
        return query
    }

    func loadToken() -> String? {
        var query = baseQuery()
        query[kSecReturnData as String] = true
        query[kSecMatchLimit as String] = kSecMatchLimitOne

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        guard status == errSecSuccess, let data = result as? Data else {
            return nil
        }
        return String(data: data, encoding: .utf8)
    }

    @discardableResult
    func saveToken(_ token: String) -> Bool {
        guard let data = token.data(using: .utf8) else { return false }
        SecItemDelete(baseQuery() as CFDictionary)
        let status = SecItemAdd(saveQuery(for: data) as CFDictionary, nil)
        return status == errSecSuccess
    }

    func deleteToken() {
        SecItemDelete(baseQuery() as CFDictionary)
    }
}
