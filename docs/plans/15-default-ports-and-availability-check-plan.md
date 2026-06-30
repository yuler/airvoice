# 15-default-ports-and-availability-check-plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Modify the default ports for the Airvoice CLI (7654) and Desktop (7655), and add a port-in-use check before starting each server, displaying any occupancy errors to the user.

**Architecture:** We will implement a shared function `CheckPortAvailable(port int) error` in the `server` package. The CLI serve command will check this and exit on error, while the Desktop app will display a native Wails error dialog on startup server failure, and return validation errors in the settings save path.

**Tech Stack:** Go (net/http, net), Wails v2 (runtime.MessageDialog), Vue 3.

---

### Task 1: Add Port Occupancy Check Helper to Go server

**Files:**
- Modify: [cli/server/server.go](file:///home/yule/Sides/airvoice/cli/server/server.go)
- Modify: [cli/server/server_test.go](file:///home/yule/Sides/airvoice/cli/server/server_test.go)

- [ ] **Step 1: Write tests for CheckPortAvailable**
  Add `TestCheckPortAvailable` in `cli/server/server_test.go`:
  ```go
  func TestCheckPortAvailable(t *testing.T) {
  	// Test with a free port
  	ln, err := net.Listen("tcp", "127.0.0.1:0")
  	if err != nil {
  		t.Fatal(err)
  	}
  	addr := ln.Addr().(*net.TCPAddr)
  	port := addr.Port
  	ln.Close()

  	if err := CheckPortAvailable(port); err != nil {
  		t.Errorf("expected port %d to be available, got error: %v", port, err)
  	}

  	// Test with an occupied port
  	ln2, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
  	if err != nil {
  		t.Fatal(err)
  	}
  	defer ln2.Close()

  	if err := CheckPortAvailable(port); err == nil {
  		t.Errorf("expected port %d to be occupied, got nil error", port)
  	}
  }
  ```

- [ ] **Step 2: Run test to verify it fails compiling**
  Run: `mise x -- go test ./cli/server -run TestCheckPortAvailable`
  Expected: FAIL (compilation error: undefined: CheckPortAvailable)

- [ ] **Step 3: Implement CheckPortAvailable**
  Add the implementation to `cli/server/server.go`:
  ```go
  import "net"

  // CheckPortAvailable checks if a TCP port is available to listen on all interfaces.
  func CheckPortAvailable(port int) error {
  	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
  	if err != nil {
  		return fmt.Errorf("port %d is already in use", port)
  	}
  	ln.Close()
  	return nil
  }
  ```

- [ ] **Step 4: Run test to verify it passes**
  Run: `mise x -- go test ./cli/server -run TestCheckPortAvailable`
  Expected: PASS

- [ ] **Step 5: Commit**
  Run:
  ```bash
  git add cli/server/server.go cli/server/server_test.go
  git commit -m "feat: add CheckPortAvailable helper and unit tests"
  ```

---

### Task 2: Update CLI Default Port and Add Startup Occupancy Check

**Files:**
- Modify: [cli/main.go](file:///home/yule/Sides/airvoice/cli/main.go)
- Modify: [cli/server/hub_test.go](file:///home/yule/Sides/airvoice/cli/server/hub_test.go)
- Modify: [cli/server/server_test.go](file:///home/yule/Sides/airvoice/cli/server/server_test.go)

- [ ] **Step 1: Modify default port and add pre-startup check in CLI serve command**
  In `cli/main.go`, change references of `7383` to `7654` inside `serve` options and usage docs. Add the pre-startup check:
  ```go
  // Inside serve command switch
  var port int
  fs := flag.NewFlagSet("serve", flag.ExitOnError)
  fs.IntVar(&port, "port", 7654, "port to listen on")
  fs.IntVar(&port, "p", 7654, "port to listen on (shorthand)")
  _ = fs.Parse(os.Args[2:])

  if err := server.CheckPortAvailable(port); err != nil {
  	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
  	os.Exit(1)
  }
  ```
  Update `printUsage()`:
  ```go
  func printUsage() {
  	fmt.Fprintf(os.Stderr, "Usage:\n")
  	fmt.Fprintf(os.Stderr, "  airvoice serve [--port 7654]\n")
  	fmt.Fprintf(os.Stderr, "  airvoice doctor\n")
  	fmt.Fprintf(os.Stderr, "  airvoice version\n")
  }
  ```

- [ ] **Step 2: Update Go server tests port references**
  Update hardcoded references to port `7383` in `cli/server/hub_test.go` and `cli/server/server_test.go` to use `7654`.

- [ ] **Step 3: Run CLI unit tests to verify they pass**
  Run: `mise x -- go test ./cli/...`
  Expected: PASS

- [ ] **Step 4: Commit**
  Run:
  ```bash
  git add cli/main.go cli/server/hub_test.go cli/server/server_test.go
  git commit -m "feat: update CLI default port to 7654 and add serve command check"
  ```

---

### Task 3: Update Desktop Application Default Port and Checks

**Files:**
- Modify: [desktop/app.go](file:///home/yule/Sides/airvoice/desktop/app.go)
- Modify: [desktop/app_test.go](file:///home/yule/Sides/airvoice/desktop/app_test.go)

- [ ] **Step 1: Update default port to 7655 in App instantiation and loading**
  In `desktop/app.go`:
  - Replace `7383` with `7655` in `NewApp()` constructor.
  - Replace `7383` with `7655` in `loadSettings()` fallback port.

- [ ] **Step 2: Add check in StartServer**
  In `desktop/app.go`, check port availability before starting:
  ```go
  func (a *App) StartServer(port int) error {
  	if err := server.CheckPortAvailable(port); err != nil {
  		return err
  	}
  ```

- [ ] **Step 3: Add error popup dialog in startup**
  In `desktop/app.go`'s `startup` method, catch the server startup error and show a native dialog box:
  ```go
  	if err := a.StartServer(a.port); err != nil {
  		log.Printf("Failed to start server: %v", err)
  		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
  			Type:    runtime.ErrorDialog,
  			Title:   "Server Error",
  			Message: fmt.Sprintf("Failed to start server: %v\n\nPlease check your settings and port occupancy.", err),
  		})
  	}
  ```

- [ ] **Step 4: Check port occupancy before saving new settings**
  In `desktop/app.go`'s `SaveSettings` method:
  ```go
  	a.mu.RLock()
  	portChanged := a.port != s.Port
  	a.mu.RUnlock()

  	if portChanged {
  		if err := server.CheckPortAvailable(s.Port); err != nil {
  			return fmt.Errorf("port %d is already in use: %w", s.Port, err)
  		}
  	}
  ```

- [ ] **Step 5: Update App tests**
  In `desktop/app_test.go`, update default port assertions from `7383` to `7655`.

- [ ] **Step 6: Run Go tests to verify correctness**
  Run: `mise x -- go test ./desktop/...`
  Expected: PASS

- [ ] **Step 7: Commit**
  Run:
  ```bash
  git add desktop/app.go desktop/app_test.go
  git commit -m "feat: update Desktop default port to 7655 and add occupancy checks"
  ```

---

### Task 4: Update Desktop Frontend Defaults

**Files:**
- Modify: [desktop/frontend/src/components/SettingsPanel.vue](file:///home/yule/Sides/airvoice/desktop/frontend/src/components/SettingsPanel.vue)
- Modify: [desktop/frontend/src/composables/useConnection.ts](file:///home/yule/Sides/airvoice/desktop/frontend/src/composables/useConnection.ts)

- [ ] **Step 1: Update SettingsPanel default port**
  In `SettingsPanel.vue`, update default port fallback to `7655`:
  ```typescript
  const settings = ref<Settings>({
    port: 7655,
    autoStart: false,
    language: 'zh-CN',
  })
  ```

- [ ] **Step 2: Update useConnection default port**
  In `useConnection.ts`, update default port to `7655`:
  ```typescript
  const status = ref<ConnectionStatus>({
    state: 'disconnected',
    deviceName: '',
    host: '',
    port: 7655,
  })
  ```

- [ ] **Step 3: Run all Go unit tests to verify everything remains clean**
  Run: `mise x -- go test ./...`
  Expected: PASS

- [ ] **Step 4: Commit**
  Run:
  ```bash
  git add desktop/frontend/src/components/SettingsPanel.vue desktop/frontend/src/composables/useConnection.ts
  git commit -m "frontend: update default port fallbacks to 7655"
  ```
