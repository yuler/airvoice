# Design System Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Update the iOS client SwiftUI views and Go CLI server print layouts to adhere strictly to the new monochrome and Vercel Blue (accent-blue) design system.

**Architecture:** Standardize UI surfaces (such as iOS screen background to pure black, button to solid accent-blue, status connection indicators, and borders instead of shadow overlays) and refactor the terminal output layout in Go with precise margins, default non-inverted QR colors, and quiet zones.

**Tech Stack:** Go (for CLI), SwiftUI (for iOS client)

---

### Task 1: Go CLI QR Code Print Configuration

Configures the `qrterminal` generator to output with default colors and an explicit 2-unit quiet zone border.

**Files:**
- Modify: `cli/pairing/qr.go`
- Test: `cli/pairing/qr_test.go`

- [ ] **Step 1: Update qr.go with Configured GenerateWithConfig**
  Replace `PrintQR` to define a config structure with L error level, stderr writer, QuietZone 2, and HalfBlocks.
  
  Code change to `cli/pairing/qr.go`:
  ```go
  package pairing

  import (
  	"os"

  	"github.com/mdp/qrterminal/v3"
  )

  // PrintQR generates a QR code from the payload and prints it to os.Stderr using half-blocks,
  // enforcing a QuietZone of 2 and terminal default colors (preventing inverted scanner issues).
  func PrintQR(payload []byte) {
  	config := qrterminal.Config{
  		Level:      qrterminal.L,
  		Writer:     os.Stderr,
  		QuietZone:  2,
  		HalfBlocks: true,
  	}
  	qrterminal.GenerateWithConfig(string(payload), config)
  }
  ```

- [ ] **Step 2: Run tests to verify the CLI build and PrintQR works**
  Run: `go test ./cli/pairing -v`
  Expected: PASS

- [ ] **Step 3: Commit**
  ```bash
  git add cli/pairing/qr.go
  git commit -m "style: configure qrterminal with quiet zone 2 and default colors"
  ```

---

### Task 2: Go CLI Terminal Print Margins and Formatting

Update terminal logs and connection status printout in the Go binary to implement 1-space indentation margins, empty lines, and clean status representations.

**Files:**
- Modify: `cli/pairing/session.go`
- Modify: `cli/server/log.go`
- Modify: `cli/main.go`

- [ ] **Step 1: Indent session printouts and add spacing**
  Update `PrintPairing` in `cli/pairing/session.go` to add left margins and trailing newlines.
  
  Code change:
  ```go
  package pairing

  import (
  	"fmt"
  	"os"

  	"github.com/google/uuid"
  )

  // PrintPairing builds a fresh pairing payload, prints the QR code to stderr, and returns the token and WebSocket URL.
  func PrintPairing(port int, banner string) (token, wsURL string, err error) {
  	token = uuid.NewString()
  	ip, err := LocalIPv4()
  	if err != nil {
  		return "", "", err
  	}
  	wsURL = fmt.Sprintf("ws://%s:%d/ws", ip, port)

  	payload := &Payload{
  		Version: 1,
  		WS:      wsURL,
  		Token:   token,
  	}
  	payloadBytes, err := payload.Marshal()
  	if err != nil {
  		return "", "", err
  	}

  	if banner != "" {
  		fmt.Fprintf(os.Stderr, "\n [airvoice] %s\n\n", banner)
  	} else {
  		fmt.Fprintf(os.Stderr, "\n")
  	}
  	PrintQR(payloadBytes)
  	fmt.Fprintf(os.Stderr, "\n")
  	fmt.Fprintf(os.Stderr, "  Token: %s\n", token)
  	fmt.Fprintf(os.Stderr, "  WebSocket URL: %s\n\n", wsURL)
  	fmt.Fprintf(os.Stderr, "  [airvoice] waiting for iPhone connection...\n\n")
  	return token, wsURL, nil
  }
  ```

- [ ] **Step 2: Add left margin indentation to logs**
  Update `logStatus` in `cli/server/log.go` to add a space prefix before `[airvoice]`.
  
  Code change:
  ```go
  package server

  import (
  	"fmt"
  	"os"
  	"strings"
  	"unicode/utf8"
  )

  func logStatus(format string, args ...any) {
  	fmt.Fprintf(os.Stderr, " [airvoice] "+format+"\n", args...)
  }
  ```

- [ ] **Step 3: Update cli start messages spacing**
  Update main server start lines in `cli/main.go` lines 59-60 to use margins and newlines.
  
  Code change:
  ```go
  		fmt.Fprintf(os.Stderr, "  Paste backend: %s\n", paster.Name())
  		fmt.Fprintf(os.Stderr, "  [airvoice] listening on %s (health: /health, ws: /ws)\n\n", addr)
  ```

- [ ] **Step 4: Run Go CLI tests**
  Run: `go test ./cli/... -v`
  Expected: PASS

- [ ] **Step 5: Commit**
  ```bash
  git add cli/pairing/session.go cli/server/log.go cli/main.go
  git commit -m "style: add margins and line spacing to CLI terminal output"
  ```

---

### Task 3: iOS Client Onboarding Screen Styling

Update onboarding view to match the pure black background, single Vercel Blue accent color, and border-based cards.

**Files:**
- Modify: `ios/Airvoice/Views/OnboardingView.swift`

- [ ] **Step 1: Set onboarding view background, colors, and button**
  Apply the new tokens to `OnboardingView.swift`:
  - Pure black background.
  - Vercel Blue logo.
  - Lighter container color and dark border (`#2e2e2e`) for steps card.
  - Solid Vercel Blue (`#006efe`) button without gradient.
  
  Code change in `ios/Airvoice/Views/OnboardingView.swift`:
  ```swift
  // Replace body contents in OnboardingView:
      var body: some View {
          ZStack {
              // Pure black background
              Color(hex: "000000").ignoresSafeArea()
              
              VStack(spacing: 24) {
                  Spacer()
                  
                  // Icon/App Logo
                  Image(systemName: "waveform.circle.fill")
                      .resizable()
                      .scaledToFit()
                      .frame(width: 80, height: 80)
                      .foregroundStyle(Color(hex: "006efe")) // Accent-blue
                      .shadow(color: Color(hex: "006efe").opacity(0.3), radius: 15)
                  
                  VStack(spacing: 8) {
                      Text("Airvoice")
                          .font(.system(size: 32, weight: .bold, design: .rounded))
                          .foregroundStyle(.white)
                      
                      Text("让手机语音输入无缝连接电脑")
                          .font(.subheadline)
                          .foregroundStyle(.secondary)
                  }
                  
                  Spacer()
                  
                  // Guide steps card
                  VStack(alignment: .leading, spacing: 16) {
                      Text("输入法安装与配置指南")
                          .font(.headline)
                          .foregroundStyle(.white)
                          .padding(.bottom, 4)
                      
                      GuideStepView(
                          number: "1",
                          title: "安装推荐输入法",
                          desc: "推荐使用「豆包输入法」（主）或「微信输入法」（备选）"
                      )
                      
                      GuideStepView(
                          number: "2",
                          title: "启用键盘",
                          desc: "前往「系统设置」→「通用」→「键盘」→「添加新键盘」"
                      )
                      
                      GuideStepView(
                          number: "3",
                          title: "隐私安全",
                          desc: "无需启用「允许完全访问」，Airvoice 仅读取本 App 内的输入框"
                      )
                  }
                  .padding(20)
                  .background(
                      RoundedRectangle(cornerRadius: 20)
                          .fill(Color(hex: "0d0e15")) // background-secondary
                          .overlay(
                              RoundedRectangle(cornerRadius: 20)
                                  .stroke(Color(hex: "2e2e2e"), lineWidth: 1) // border-default
                          )
                  )
                  .padding(.horizontal, 24)
                  
                  Spacer()
                  
                  // Start Button
                  Button(action: {
                      hasSeenOnboarding = true
                  }) {
                      Text("开始使用")
                          .font(.headline)
                          .foregroundStyle(.white)
                          .frame(maxWidth: .infinity)
                          .frame(height: 56)
                          .background(Color(hex: "006efe")) // accent-blue
                          .cornerRadius(28)
                          .shadow(color: Color(hex: "006efe").opacity(0.3), radius: 10)
                  }
  ```

- [ ] **Step 2: Commit**
  ```bash
  git add ios/Airvoice/Views/OnboardingView.swift
  git commit -m "style: update onboarding screen background and button colors"
  ```

---

### Task 4: iOS Client Home Screen Styling

Apply unified tokens to the primary utility screen, including connecting indicators, text inputs, status bar, and active buttons.

**Files:**
- Modify: `ios/Airvoice/Views/HomeView.swift`

- [ ] **Step 1: Apply style updates to HomeView**
  Updates include:
  - Pure black background.
  - TextEditor frame background using `0d0e15` with `#2e2e2e` border default.
  - Active button background changes from gradient to solid `#006efe` with matching shadow.
  - State color updates: connecting status switches from orange to yellow.
  
  Code changes:
  ```swift
  // Replace ZStack color in HomeView.swift (around line 15):
              Color(hex: "000000").ignoresSafeArea() // Pure black background
  ```
  ```swift
  // Replace TextEditor container styling (around line 61-67):
                  .frame(maxWidth: .infinity, maxHeight: .infinity)
                  .background(Color(hex: "0d0e15")) // background-secondary
                  .cornerRadius(16)
                  .overlay(
                      RoundedRectangle(cornerRadius: 16)
                          .stroke(Color(hex: "2e2e2e"), lineWidth: 1) // border-default
                  )
  ```
  ```swift
  // Replace Mic button background & shadow (around line 97-103):
                          .background(
                              connection.state == .connected ?
                              Color(hex: "006efe") : // accent-blue
                              Color.gray.opacity(0.3)
                          )
                          .cornerRadius(28)
                          .shadow(color: connection.state == .connected ? Color(hex: "006efe").opacity(0.3) : Color.clear, radius: 10)
  ```
  ```swift
  // Update statusColor return for connecting (around line 160):
          case .connecting: return .yellow
  ```

- [ ] **Step 2: Commit changes**
  ```bash
  git add ios/Airvoice/Views/HomeView.swift
  git commit -m "style: apply pure black background, accent-blue button, and yellow connecting state to HomeView"
  ```
