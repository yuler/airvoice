import XCTest
import Security
@testable import Airvoice

private final class InMemoryTokenStore: TokenStore {
    var storedToken: String?

    func loadToken() -> String? {
        storedToken
    }

    @discardableResult
    func saveToken(_ token: String) -> Bool {
        storedToken = token
        return true
    }

    func deleteToken() {
        storedToken = nil
    }
}

@MainActor
final class ConnectionManagerTests: XCTestCase {
    private var suiteName: String!
    private var defaults: UserDefaults!
    private var tokenStore: InMemoryTokenStore!

    override func setUp() {
        super.setUp()
        suiteName = "cc.yuler.airvoice.tests.\(UUID().uuidString)"
        defaults = UserDefaults(suiteName: suiteName)!
        defaults.removePersistentDomain(forName: suiteName)
        tokenStore = InMemoryTokenStore()
    }

    override func tearDown() {
        defaults.removePersistentDomain(forName: suiteName)
        defaults = nil
        suiteName = nil
        tokenStore = nil
        super.tearDown()
    }

    private func makePayload(
        ws: String = "ws://127.0.0.1:7655/ws",
        token: String = "tok-1"
    ) -> PairingPayload {
        PairingPayload(v: 1, ws: ws, token: token)
    }

    func testConnectPersistsPayloadAcrossInstancesUsingTokenStore() {
        let payload = makePayload(token: "persisted-token")
        let manager1 = ConnectionManager(userDefaults: defaults, tokenStore: tokenStore)
        var connectedPayloads: [PairingPayload] = []
        manager1.connectHandlerForTesting = { connectedPayloads.append($0) }

        manager1.connect(payload: payload)
        XCTAssertEqual(connectedPayloads, [payload])
        XCTAssertEqual(manager1.lastPayload, payload)
        XCTAssertTrue(manager1.canReconnect)
        XCTAssertEqual(tokenStore.storedToken, "persisted-token")

        // Token must NOT be stored in plaintext UserDefaults.
        XCTAssertNil(defaults.data(forKey: ConnectionManager.savedPayloadKey))
        if let metaData = defaults.data(forKey: ConnectionManager.savedMetadataKey),
           let metaString = String(data: metaData, encoding: .utf8) {
            XCTAssertFalse(metaString.contains("persisted-token"))
        } else {
            XCTFail("Expected pairing metadata to be saved in UserDefaults")
        }

        // Simulate cold app restart with a fresh ConnectionManager backed by the same stores.
        let manager2 = ConnectionManager(userDefaults: defaults, tokenStore: tokenStore)
        var autoConnectedPayloads: [PairingPayload] = []
        manager2.connectHandlerForTesting = { autoConnectedPayloads.append($0) }

        XCTAssertEqual(manager2.lastPayload, payload)
        XCTAssertTrue(manager2.canReconnect)

        manager2.connectSavedIfAvailable()
        XCTAssertEqual(autoConnectedPayloads, [payload])
        XCTAssertEqual(manager2.state, .connecting)
    }

    func testLegacyPlaintextUserDefaultsMigratesToTokenStoreAndPurgesOldKey() throws {
        let legacy = makePayload(token: "legacy-secret-token")
        let data = try JSONEncoder().encode(legacy)
        defaults.set(data, forKey: ConnectionManager.savedPayloadKey)

        let manager = ConnectionManager(userDefaults: defaults, tokenStore: tokenStore)
        XCTAssertEqual(manager.lastPayload, legacy)
        XCTAssertEqual(tokenStore.storedToken, "legacy-secret-token")
        XCTAssertNil(defaults.data(forKey: ConnectionManager.savedPayloadKey))
    }

    func testKeychainTokenStoreUsesDeviceOnlyAfterFirstUnlockAccessibility() {
        let store = KeychainTokenStore(service: "cc.yuler.airvoice.test", account: "token")
        let query = store.saveQuery(for: Data("secret".utf8))
        let accessibility = query[kSecAttrAccessible as String] as? String
        XCTAssertEqual(accessibility, kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly as String)
    }

    func testMakeWebSocketRequestUsesBearerHeaderWithoutQueryToken() throws {
        let payload = makePayload(ws: "ws://192.168.1.20:7655/ws", token: "secret-header-token")
        let request = try ConnectionManager.makeWebSocketRequest(for: payload)
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer secret-header-token")
        XCTAssertNil(request.url?.query)
        XCTAssertFalse(request.url?.absoluteString.contains("secret-header-token") ?? true)
    }

    func testRemoteWssBlockedUnlessExplicitlyAllowed() {
        let remotePayload = makePayload(ws: "wss://remote.example.com/ws", token: "tok-remote")
        let manager = ConnectionManager(userDefaults: defaults, tokenStore: tokenStore)
        var connectedPayloads: [PairingPayload] = []
        manager.connectHandlerForTesting = { connectedPayloads.append($0) }

        manager.connect(payload: remotePayload)
        XCTAssertTrue(connectedPayloads.isEmpty)
        XCTAssertNil(manager.lastPayload)
        if case .error = manager.state {
            // expected
        } else {
            XCTFail("Expected error state for unconfirmed remote host, got \(manager.state)")
        }

        manager.connect(payload: remotePayload, allowRemote: true)
        XCTAssertEqual(connectedPayloads, [remotePayload])
        XCTAssertEqual(manager.lastPayload, remotePayload)
        XCTAssertEqual(manager.state, .connecting)
    }

    func testClearStoredPayloadRemovesFromUserDefaultsAndTokenStore() {
        let manager = ConnectionManager(userDefaults: defaults, tokenStore: tokenStore)
        manager.connectHandlerForTesting = { _ in }
        manager.connect(payload: makePayload())

        manager.clearStoredPayload()

        XCTAssertNil(manager.lastPayload)
        XCTAssertFalse(manager.canReconnect)
        XCTAssertNil(tokenStore.storedToken)
        XCTAssertNil(ConnectionManager.loadSavedPayload(from: defaults, tokenStore: tokenStore))
    }

    func testDisconnectAndClearResetsConnectionAndStoredAddress() {
        let manager = ConnectionManager(userDefaults: defaults, tokenStore: tokenStore)
        manager.connectHandlerForTesting = { _ in }
        manager.connect(payload: makePayload(ws: "ws://192.168.1.9:7655/ws"))
        XCTAssertEqual(manager.savedServerAddress, "ws://192.168.1.9:7655/ws")

        manager.disconnectAndClear()

        XCTAssertEqual(manager.state, .disconnected)
        XCTAssertNil(manager.savedServerAddress)
        XCTAssertNil(manager.lastPayload)
        XCTAssertNil(tokenStore.storedToken)
    }

    func testExponentialBackoffProgressionAndCap() {
        let manager = ConnectionManager(
            userDefaults: defaults,
            tokenStore: tokenStore,
            initialBackoffNanoseconds: 2_000_000_000,
            maxBackoffNanoseconds: 30_000_000_000
        )
        manager.connectHandlerForTesting = { _ in }
        manager.connect(payload: makePayload())

        // 1st failure schedules 2s, next becomes 4s
        manager.handleConnectionFailure(for: nil, statusCode: nil, message: "offline")
        XCTAssertEqual(manager.state, .error("offline"))
        XCTAssertTrue(manager.isReconnectScheduled)
        XCTAssertEqual(manager.currentBackoffNanoseconds, 4_000_000_000)

        // 2nd failure: 4s -> 8s
        manager.handleConnectionFailure(for: nil, statusCode: nil, message: "offline")
        XCTAssertEqual(manager.currentBackoffNanoseconds, 8_000_000_000)

        // 3rd failure: 8s -> 16s
        manager.handleConnectionFailure(for: nil, statusCode: nil, message: "offline")
        XCTAssertEqual(manager.currentBackoffNanoseconds, 16_000_000_000)

        // 4th failure: 16s -> capped at 30s
        manager.handleConnectionFailure(for: nil, statusCode: nil, message: "offline")
        XCTAssertEqual(manager.currentBackoffNanoseconds, 30_000_000_000)

        // 5th failure stays capped at 30s
        manager.handleConnectionFailure(for: nil, statusCode: nil, message: "offline")
        XCTAssertEqual(manager.currentBackoffNanoseconds, 30_000_000_000)

        manager.disconnect()
    }

    func testPauseAndResumeReconnect() {
        let manager = ConnectionManager(
            userDefaults: defaults,
            tokenStore: tokenStore,
            initialBackoffNanoseconds: 2_000_000_000,
            maxBackoffNanoseconds: 30_000_000_000
        )
        var attempts = 0
        manager.connectHandlerForTesting = { _ in attempts += 1 }
        manager.connect(payload: makePayload())
        XCTAssertEqual(attempts, 1)

        manager.handleConnectionFailure(for: nil, statusCode: nil, message: "dropped")
        XCTAssertTrue(manager.isReconnectScheduled)
        XCTAssertEqual(manager.currentBackoffNanoseconds, 4_000_000_000)

        // Going to background cancels pending reconnect timer.
        manager.pauseReconnect()
        XCTAssertFalse(manager.isReconnectScheduled)

        // Returning to foreground resets backoff and reconnects immediately.
        manager.resumeReconnect()
        XCTAssertEqual(attempts, 2)
        XCTAssertEqual(manager.state, .connecting)
        XCTAssertEqual(manager.currentBackoffNanoseconds, 2_000_000_000)

        manager.disconnect()
    }

    func testUnauthorized401ClearsPayloadStopsReconnectAndNotifies() {
        let manager = ConnectionManager(userDefaults: defaults, tokenStore: tokenStore)
        manager.connectHandlerForTesting = { _ in }
        manager.connect(payload: makePayload(token: "stale-token"))

        var authExpiredCount = 0
        var transportErrorCount = 0
        manager.onAuthExpired = { authExpiredCount += 1 }
        manager.onTransportError = { _ in transportErrorCount += 1 }

        manager.handleConnectionFailure(for: nil, statusCode: 401, message: "Unauthorized")

        XCTAssertEqual(authExpiredCount, 1)
        XCTAssertEqual(transportErrorCount, 0)
        XCTAssertEqual(manager.state, .disconnected)
        XCTAssertNil(manager.lastPayload)
        XCTAssertFalse(manager.canReconnect)
        XCTAssertFalse(manager.isReconnectScheduled)
        XCTAssertNil(tokenStore.storedToken)
        XCTAssertNil(ConnectionManager.loadSavedPayload(from: defaults, tokenStore: tokenStore))
    }

    func testHomeViewModelSilentReconnectErrorAndAuthExpiredToast() async {
        let settings = AppSettings(userDefaults: defaults)
        settings.language = .zh
        let manager = ConnectionManager(userDefaults: defaults, tokenStore: tokenStore)
        manager.connectHandlerForTesting = { _ in }
        let autoSend = AutoSendController()
        let viewModel = HomeViewModel(settings: settings)
        viewModel.wire(connection: manager, autoSend: autoSend)

        manager.connect(payload: makePayload())

        // Idle transport failure should not spam error toasts.
        manager.handleConnectionFailure(for: nil, statusCode: nil, message: "network down")
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertNil(viewModel.toastMessage)

        // Active in-flight send failure should show an error toast and clear inFlight.
        autoSend.beginSend()
        manager.handleConnectionFailure(for: nil, statusCode: nil, message: "network down")
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertFalse(autoSend.inFlight)
        XCTAssertEqual(viewModel.toastMessage, "发送失败，请检查连接")
        XCTAssertTrue(viewModel.isToastError)

        // 401 auth expiration should alert user to re-scan QR code.
        manager.handleConnectionFailure(for: nil, statusCode: 401, message: "Unauthorized")
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(viewModel.toastMessage, "配对已失效，请重新扫码")
        XCTAssertTrue(viewModel.isToastError)

        manager.disconnect()
    }

    func testHomeViewModelAckRemovesOnlySentPrefixAndKeepsRemaining() async {
        let manager = ConnectionManager(userDefaults: defaults, tokenStore: tokenStore)
        var sentMessages: [(String, String)] = []
        manager.sendTextHandlerForTesting = { id, content in
            sentMessages.append((id, content))
            return true
        }
        manager.state = .connected

        let autoSend = AutoSendController()
        let viewModel = HomeViewModel()
        viewModel.wire(connection: manager, autoSend: autoSend)

        viewModel.text = "hello"
        viewModel.manualSend(connection: manager, autoSend: autoSend)
        XCTAssertEqual(sentMessages.count, 1)
        let msgId = sentMessages[0].0

        // User keeps dictating while the first segment is in flight.
        viewModel.text = "hello world"

        manager.onAck?(msgId, true, nil)
        try? await Task.sleep(nanoseconds: 50_000_000)

        XCTAssertEqual(viewModel.text, " world")
        XCTAssertFalse(autoSend.inFlight)
        XCTAssertTrue(autoSend.countdownActive)
    }

    func testHomeViewModelIgnoresLateAckAfterCancelWithoutClearingNewTextOrReleasingNewInFlight() async {
        let settings = AppSettings(userDefaults: defaults)
        settings.language = .zh
        let manager = ConnectionManager(userDefaults: defaults, tokenStore: tokenStore)
        var sentMessages: [(String, String)] = []
        manager.sendTextHandlerForTesting = { id, content in
            sentMessages.append((id, content))
            return true
        }
        manager.state = .connected

        let autoSend = AutoSendController()
        let viewModel = HomeViewModel(settings: settings)
        viewModel.wire(connection: manager, autoSend: autoSend)

        // Send first message, then cancel it.
        viewModel.text = "first message"
        viewModel.manualSend(connection: manager, autoSend: autoSend)
        XCTAssertEqual(sentMessages.count, 1)
        let cancelledMsgId = sentMessages[0].0

        viewModel.cancelSend(autoSend: autoSend)
        XCTAssertFalse(autoSend.inFlight)
        XCTAssertEqual(viewModel.toastMessage, "已取消发送")

        // User starts typing and sends a second message.
        viewModel.text = "second message"
        viewModel.manualSend(connection: manager, autoSend: autoSend)
        XCTAssertEqual(sentMessages.count, 2)
        let activeMsgId = sentMessages[1].0
        XCTAssertTrue(autoSend.inFlight)

        // Late ACK (both ok: true and ok: false) for the cancelled message must be ignored.
        manager.onAck?(cancelledMsgId, true, nil)
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(viewModel.text, "second message")
        XCTAssertTrue(autoSend.inFlight)

        manager.onAck?(cancelledMsgId, false, "late error")
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(viewModel.text, "second message")
        XCTAssertTrue(autoSend.inFlight)

        // Active message ACK still succeeds normally.
        manager.onAck?(activeMsgId, true, nil)
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(viewModel.text, "")
        XCTAssertFalse(autoSend.inFlight)
    }

    func testHomeViewModelIgnoresUnknownAckWhenIdleWithoutClearingDraft() async {
        let manager = ConnectionManager(userDefaults: defaults, tokenStore: tokenStore)
        manager.state = .connected

        let autoSend = AutoSendController()
        let viewModel = HomeViewModel()
        viewModel.wire(connection: manager, autoSend: autoSend)

        viewModel.text = "unsent draft"
        manager.onAck?("unknown-stale-id", true, nil)
        try? await Task.sleep(nanoseconds: 50_000_000)

        XCTAssertEqual(viewModel.text, "unsent draft")
        XCTAssertFalse(autoSend.inFlight)
    }
}
