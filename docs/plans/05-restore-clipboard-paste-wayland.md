# Implementation Plan: Restore Clipboard Copy and Paste via ydotool on Wayland

**Goal:** Modify the Wayland paste implementation to copy the text to the clipboard first (using `wl-copy`) and then trigger the paste operation by simulating a Ctrl+V keystroke using `ydotool key CTRL+v`, as done in the `typo` project. This ensures that non-ASCII and Chinese characters are input correctly at the cursor position.

---

### Task 1: Update Wayland paste implementation

**Files:**
- [linux_wayland.go](file:///home/yule/Sides/airvoice/cli/paste/linux_wayland.go)

**Changes:**
1. Restore the `wl-copy` command to write the text to the system clipboard first.
2. Sleep for 80 milliseconds to ensure the clipboard is updated.
3. Simulate `Ctrl+V` using `ydotool key CTRL+v` instead of direct typing or raw scancodes.

---

### Task 2: Update unit tests

**Files:**
- [linux_test.go](file:///home/yule/Sides/airvoice/cli/paste/linux_test.go)

**Changes:**
1. Update `TestLinuxPasters/waylandPaster success` to assert the sequence:
   * First, `wl-copy` with the text.
   * Second, `ydotool` with `key CTRL+v` arguments.
2. Update the failure tests to verify failure if `wl-copy` or `ydotool` fails.
