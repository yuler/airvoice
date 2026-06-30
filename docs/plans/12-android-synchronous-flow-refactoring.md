# Implementation Plan: Android Synchronous Flow Refactoring

This plan outlines the changes required to refactor the Android client's message sending, receiving, and UI updates to follow a synchronous suspending flow (using Kotlin Coroutines structured concurrency) rather than asynchronous callback-based event handlers.

## Proposed Changes

### Android Client

#### [MODIFY] [AutoSendController.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/services/AutoSendController.kt)
* Import `CompletableDeferred`, `withTimeout`, and `TimeoutCancellationException`.
* Declare `private val pendingAcks = ConcurrentHashMap<String, CompletableDeferred<Boolean>>()` to track active messages.
* Remove `pendingMessageId`, `timeoutJob`, `sentMessages`, and `sendingText` fields.
* Change `attemptSend` to a `suspend` function:
  * Generate `msgId`, create a `CompletableDeferred<Boolean>`, and insert it into `pendingAcks`.
  * Send the WebSocket text message.
  * Use `withTimeout(5000L)` to suspend and await the deferred result.
  * If a timeout occurs, return `false`.
  * In the `finally` block, ensure the ID is removed from `pendingAcks`.
  * Trigger `onSentAck(...)` with the result, update `lastAckedText` on success, and set `_inFlight` appropriately.
* In `startListening()`:
  * Simplify the collector: when a message of type `"ack"` is received, extract its `id`, lookup the corresponding `deferred` in `pendingAcks`, and call `deferred.complete(msg.ok == true)`.
  * On connection drop (`Disconnected` or `Error`), retrieve all active deferreds from `pendingAcks`, clear the map, set `_inFlight` to `false`, and complete them all with `false` (resuming the suspended sending coroutines with a failure result).
* Change `sendPendingText` to a `suspend` function so it can call `attemptSend` sequentially.
* In `triggerImmediateSend()`, wrap `attemptSend` in a `scope.launch { ... }` block.

#### [MODIFY] [AirvoiceViewModel.kt](file:///home/yule/Sides/airvoice/android/app/src/main/java/com/yule/airvoice/ui/viewmodel/AirvoiceViewModel.kt)
* In `manualSend()`, wrap the suspending call `autoSendController.attemptSend(...)` inside a `viewModelScope.launch` coroutine.

## Verification Plan

### Automated Tests
* Run unit tests on Android to verify compilation and baseline logic:
  ```bash
  JAVA_HOME=/opt/android-studio/jbr/ ANDROID_HOME=/home/yule/Android/Sdk ./gradlew test
  ```
