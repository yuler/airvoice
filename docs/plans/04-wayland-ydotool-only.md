# Implementation Plan: Wayland ydotool-only Typing

**Goal:** Modify the Wayland paste implementation to type the content directly using `ydotool type` instead of copying it to the clipboard with `wl-copy` and simulating a Ctrl+V keystroke. This eliminates the dependency on `wl-clipboard` (`wl-copy`).

---

### Task 1: Update Wayland paste implementation

**Files:**
- [linux_wayland.go](file:///home/yule/Sides/airvoice/cli/paste/linux_wayland.go)

**Changes:**
1. Remove `wl-copy` invocation and the 80ms sleep.
2. Call `ydotool type --file -` with the text as stdin to type the text directly.

---

### Task 2: Update unit tests

**Files:**
- [linux_test.go](file:///home/yule/Sides/airvoice/cli/paste/linux_test.go)

**Changes:**
1. Update `TestLinuxPasters/waylandPaster success` to expect a single invocation of `ydotool` with `type --file -` arguments and the input text as stdin.
