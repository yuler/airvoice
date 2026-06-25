import XCTest
@testable import Airvoice

@MainActor
final class AutoSendControllerTests: XCTestCase {
    var controller: AutoSendController!
    var sentTexts: [(String, SendTrigger)]!

    override func setUp() {
        super.setUp()
        controller = AutoSendController()
        sentTexts = []
        controller.onSend = { [weak self] text, trigger in
            self?.sentTexts.append((text, trigger))
            self?.controller.beginSend()
            return true
        }
    }

    override func tearDown() {
        controller = nil
        sentTexts = nil
        super.tearDown()
    }

    private func sleep(_ fractionOfDelay: Double) async {
        let seconds = controller.autoSendDelay * fractionOfDelay
        try? await Task.sleep(nanoseconds: UInt64(seconds * 1_000_000_000))
    }

    func testTextDidChangeDebounceAndSend() async {
        controller.textDidChange("hello")

        XCTAssertTrue(sentTexts.isEmpty)

        await sleep(1.3)

        XCTAssertEqual(sentTexts.count, 1)
        XCTAssertEqual(sentTexts.first?.0, "hello")
        XCTAssertEqual(sentTexts.first?.1, .auto)
        XCTAssertTrue(controller.inFlight)
    }

    func testDebounceCancellation() async {
        controller.textDidChange("hello")
        await sleep(0.3)

        controller.textDidChange("hello world")

        await sleep(0.5)
        XCTAssertTrue(sentTexts.isEmpty)

        await sleep(0.8)
        XCTAssertEqual(sentTexts.count, 1)
        XCTAssertEqual(sentTexts.first?.0, "hello world")
    }

    func testDeduplicationAfterAck() async {
        controller.attemptSend("duplicate test")
        XCTAssertEqual(sentTexts.count, 1)

        controller.markAcked("duplicate test")
        XCTAssertFalse(controller.inFlight)

        controller.attemptSend("duplicate test")
        XCTAssertEqual(sentTexts.count, 1)

        controller.attemptSend("different test")
        XCTAssertEqual(sentTexts.count, 2)
        XCTAssertEqual(sentTexts.last?.0, "different test")
    }

    func testInFlightLock() async {
        controller.attemptSend("first")
        XCTAssertTrue(controller.inFlight)
        XCTAssertEqual(sentTexts.count, 1)

        controller.attemptSend("second")
        XCTAssertEqual(sentTexts.count, 1)

        controller.clearInFlight()

        controller.attemptSend("second")
        XCTAssertEqual(sentTexts.count, 2)
        XCTAssertEqual(sentTexts.last?.0, "second")
    }

    func testSendNowUsesManualTrigger() {
        XCTAssertTrue(controller.sendNow("manual send", force: true))
        XCTAssertEqual(sentTexts.count, 1)
        XCTAssertEqual(sentTexts.first?.1, .manual)
    }
}
