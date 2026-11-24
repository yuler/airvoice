import XCTest
@testable import Airvoice

@MainActor
final class ConnectionManagerTests: XCTestCase {
    private var suiteName: String!
    private var defaults: UserDefaults!

    override func setUp() {
        super.setUp()
        suiteName = "cc.yuler.airvoice.tests.\(UUID().uuidString)"
        defaults = UserDefaults(suiteName: suiteName)!
        defaults.removePersistentDomain(forName: suiteName)
    }

    override func tearDown() {
        defaults.removePersistentDomain(forName: suiteName)
        defaults = nil
        suiteName = nil
        super.tearDown()
    }

    private func makePayload(token: String = "tok-1") -> PairingPayload {
        PairingPayload(v: 1, ws: "ws://127.0.0.1:7655/ws", token: token)
    }

    func testConnectPersistsPayloadAcrossInstances() {
        let payload = makePayload(token: "persisted-token")
        let manager1 = ConnectionManager(userDefaults: defaults)
        var connectedPayloads: [PairingPayload] = []
        manager1.connectHandlerForTesting = { connectedPayloads.append($0) }

        manager1.connect(payload: payload)
        XCTAssertEqual(connectedPayloads, [payload])
        XCTAssertEqual(manager1.lastPayload, payload)
        XCTAssertTrue(manager1.canReconnect)

        // Simulate cold app restart with a fresh ConnectionManager backed by the same UserDefaults.
        let manager2 = ConnectionManager(userDefaults: defaults)
        var autoConnectedPayloads: [PairingPayload] = []
        manager2.connectHandlerForTesting = { autoConnectedPayloads.append($0) }

        XCTAssertEqual(manager2.lastPayload, payload)
        XCTAssertTrue(manager2.canReconnect)

        manager2.connectSavedIfAvailable()
        XCTAssertEqual(autoConnectedPayloads, [payload])
        XCTAssertEqual(manager2.state, .connecting)
    }

    func testClearStoredPayloadRemovesFromUserDefaults() {
        let manager = ConnectionManager(userDefaults: defaults)
        manager.connectHandlerForTesting = { _ in }
        manager.connect(payload: makePayload())

        manager.clearStoredPayload()

        XCTAssertNil(manager.lastPayload)
        XCTAssertFalse(manager.canReconnect)
        XCTAssertNil(ConnectionManager.loadSavedPayload(from: defaults))
    }

    func testExponentialBackoffProgressionAndCap() {
        let manager = ConnectionManager(
            userDefaults: defaults,
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
        let manager = ConnectionManager(userDefaults: defaults)
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
        XCTAssertNil(ConnectionManager.loadSavedPayload(from: defaults))
    }

    func testHomeViewModelSilentReconnectErrorAndAuthExpiredToast() async {
        let manager = ConnectionManager(userDefaults: defaults)
        manager.connectHandlerForTesting = { _ in }
        let autoSend = AutoSendController()
        let viewModel = HomeViewModel()
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
        let manager = ConnectionManager(userDefaults: defaults)
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
}
