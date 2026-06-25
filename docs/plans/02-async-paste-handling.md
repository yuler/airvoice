# Implementation Plan: Asynchronous Paste Handling

**Goal:** Modify the server connection handler to process text paste actions asynchronously. This prevents clipboard operations (which run external tools like `xclip`, `xdotool`, etc.) from blocking the main WebSocket connection read loop, avoiding client timeouts and connection drops.

---

### Task 1: Update WebSocket handler

**Files:**
- [handler.go](file:///home/yule/Sides/airvoice/cli/server/handler.go)

**Changes:**
1. Introduce a connection-level `sync.Mutex` or synchronized helper function to protect concurrent writes to the WebSocket (`conn.WriteMessage`), since Gorilla WebSocket does not support concurrent writes.
2. In the `text` case, spawn a new goroutine to perform the paste operation and send the response ack. This keeps the main WebSocket read loop unblocked so it can immediately receive subsequent frames or close messages.

---

### Task 2: Verify Go tests

**Changes:**
1. Run `go test ./cli/...` to verify that the implementation works correctly and does not break existing test cases.
