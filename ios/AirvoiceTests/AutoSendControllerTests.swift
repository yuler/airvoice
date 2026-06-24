import XCTest
@testable import Airvoice

@MainActor
final class AutoSendControllerTests: XCTestCase {
    var controller: AutoSendController!
    var sentTexts: [String]!
    
    override func setUp() {
        super.setUp()
        controller = AutoSendController()
        sentTexts = []
        controller.onSend = { [weak self] text in
            self?.sentTexts.append(text)
        }
    }
    
    override func tearDown() {
        controller = nil
        sentTexts = nil
        super.tearDown()
    }
    
    func testTextDidChangeDebounceAndSend() async {
        controller.textDidChange("hello")
        
        // At this point, should not have sent because of 1.5s debounce
        XCTAssertTrue(sentTexts.isEmpty)
        
        // Wait 1.6s
        try? await Task.sleep(nanoseconds: 1_600_000_000)
        
        XCTAssertEqual(sentTexts.count, 1)
        XCTAssertEqual(sentTexts.first, "hello")
        XCTAssertTrue(controller.inFlight)
    }
    
    func testDebounceCancellation() async {
        controller.textDidChange("hello")
        try? await Task.sleep(nanoseconds: 500_000_000)
        
        // Update text again before 1.5s
        controller.textDidChange("hello world")
        
        // Wait another 1.2s (total 1.7s from first, but only 1.2s from second)
        try? await Task.sleep(nanoseconds: 1_200_000_000)
        XCTAssertTrue(sentTexts.isEmpty) // Not fired yet
        
        // Wait another 0.4s (to exceed 1.5s from second change)
        try? await Task.sleep(nanoseconds: 400_000_000)
        XCTAssertEqual(sentTexts.count, 1)
        XCTAssertEqual(sentTexts.first, "hello world")
    }
    
    func testKeyboardDidHideTriggersSendImmediately() async {
        controller.keyboardDidHide(currentText: "hello immediate")
        
        // Give runloop a moment to execute the task
        try? await Task.sleep(nanoseconds: 50_000_000)
        
        XCTAssertEqual(sentTexts.count, 1)
        XCTAssertEqual(sentTexts.first, "hello immediate")
    }
    
    func testDeduplicationAfterAck() async {
        controller.keyboardDidHide(currentText: "duplicate test")
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(sentTexts.count, 1)
        
        // Mark acked
        controller.markAcked("duplicate test")
        XCTAssertFalse(controller.inFlight)
        
        // Try sending same text again, should be skipped
        controller.keyboardDidHide(currentText: "duplicate test")
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(sentTexts.count, 1) // still 1
        
        // Try sending different text, should work
        controller.keyboardDidHide(currentText: "different test")
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(sentTexts.count, 2)
        XCTAssertEqual(sentTexts.last, "different test")
    }
    
    func testInFlightLock() async {
        controller.keyboardDidHide(currentText: "first")
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertTrue(controller.inFlight)
        XCTAssertEqual(sentTexts.count, 1)
        
        // Attempt to send "second" while in-flight is true
        controller.keyboardDidHide(currentText: "second")
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(sentTexts.count, 1) // second was ignored
        
        // Clear in-flight
        controller.clearInFlight()
        XCTAssertFalse(controller.inFlight)
        
        // Now it should send
        controller.keyboardDidHide(currentText: "second")
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(sentTexts.count, 2)
        XCTAssertEqual(sentTexts.last, "second")
    }
}
