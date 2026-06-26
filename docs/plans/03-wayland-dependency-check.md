# Implementation Plan: Wayland ydotool Dependency Check

**Goal:** On Linux Wayland, validate that the required input emulation tool `ydotool` is installed before starting the server. If it is missing, print an installation prompt and exit.

---

### Task 1: Update Linux paste initialization

**Files:**
- [paste_linux.go](file:///home/yule/Sides/airvoice/cli/paste/paste_linux.go)

**Changes:**
1. Import `os/exec`.
2. Under the `SessionWayland` case, check if `ydotool` is available in the system PATH using `exec.LookPath("ydotool")`.
3. If it is not found, return a descriptive error prompt instructing the user to install `ydotool`.
