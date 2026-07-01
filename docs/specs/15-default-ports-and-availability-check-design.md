# 15-default-ports-and-availability-check-design

## Description
This design document addresses the requirement to update the default server ports for both the CLI and Desktop clients of Airvoice, and to check for port occupancy before starting:
1. **CLI Default Port**: Change from `7383` to `7654`.
2. **Desktop Default Port**: Change from `7383` to `7655`.
3. **Port Check**: Implement a pre-startup verification check. If the port is in use, throw an error.
   - For CLI: Display the error on the terminal and exit with status code `1`.
   - For Desktop (at startup): Show a native error dialog explaining that the port is occupied and keep the app running in the `disconnected` state.
   - For Desktop (during settings update): Return the validation error to the frontend settings modal, preventing the settings from being saved with an in-use port.

---

## Proposed Solution

### Shared Port Availability Helper
We will implement a helper function `CheckPortAvailable(port int) error` in the shared Go backend module `cli/server/server.go`. This function will attempt to listen on the TCP port on all interfaces (`fmt.Sprintf(":%d", port)`). If the port is already in use, binding will fail, and we will return the error; otherwise, we close the listener and return `nil`.

### CLI Integration
1. Parse flags in `cli/main.go` using the new default port `7654`.
2. Run `server.CheckPortAvailable(port)` immediately after parsing flags.
3. If it returns an error, output it to standard error and terminate with exit code `1`.

### Desktop Integration
1. Update port defaults to `7655` in the `App` initialization, loaded settings defaults, and the frontend fallback variables.
2. In `StartServer(port)`, run `server.CheckPortAvailable(port)` before establishing any resources. If it fails, return the error.
3. In `startup(ctx)`, if `StartServer` returns an error, intercept it and show a native error dialog box utilizing `runtime.MessageDialog`.
4. In `SaveSettings(s Settings)`, if the port is changed, run `CheckPortAvailable(s.Port)` first. If it is occupied, return the error immediately to the Wails frontend layer to display it on the Settings modal, without changing settings or stopping the running server.

---

## Architectural & Code Changes

### Files to Modify
- [cli/server/server.go](file:///home/yule/Sides/airvoice/cli/server/server.go): Implement `CheckPortAvailable(port int) error`.
- [cli/main.go](file:///home/yule/Sides/airvoice/cli/main.go): Update default port to `7654`, add pre-serve port check, and update usage docs.
- [cli/server/hub_test.go](file:///home/yule/Sides/airvoice/cli/server/hub_test.go): Update port `7383` references in test setups to `7654`.
- [cli/server/server_test.go](file:///home/yule/Sides/airvoice/cli/server/server_test.go): Update port `7383` references in test setups to `7654`.
- [desktop/app.go](file:///home/yule/Sides/airvoice/desktop/app.go): Update default port to `7655`, add port availability check in `StartServer` and `SaveSettings`, and display dialog on startup failure.
- [desktop/app_test.go](file:///home/yule/Sides/airvoice/desktop/app_test.go): Update expected default port to `7655`.
- [desktop/frontend/src/components/SettingsPanel.vue](file:///home/yule/Sides/airvoice/desktop/frontend/src/components/SettingsPanel.vue): Update default port `7383` to `7655`.
- [desktop/frontend/src/composables/useConnection.ts](file:///home/yule/Sides/airvoice/desktop/frontend/src/composables/useConnection.ts): Update default port `7383` to `7655`.

---

## Verification Plan

### Automated Tests
- Run `mise x -- go test ./...` to verify all CLI and Desktop unit tests compile and pass.

### Manual Verification
1. **CLI Port Check**:
   - Start the CLI: `mise run cli:dev` (should start on `7654`).
   - Start another instance in a separate shell on the same port. Verify it fails with a port occupied error and exits immediately.
2. **Desktop Startup Dialog**:
   - Occupy port `7655` (e.g., using python `python3 -m http.server 7655` or netcat).
   - Launch the Desktop app: `mise run desktop:dev`.
   - Verify that an error dialog is presented indicating port `7655` is already in use, and the app interface status is shown as "disconnected".
3. **Desktop Settings Error**:
   - Open the Desktop app (with a free `7655` port).
   - Occupy port `7656`.
   - Open settings, modify the port to `7656`, and click Save.
   - Verify that the settings modal reports a validation error and prevents saving or closing.
