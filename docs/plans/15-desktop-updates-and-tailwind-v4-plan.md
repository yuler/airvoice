# Desktop Updates and Tailwind v4 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the theme toggle, fix the Linux system tray compilation, update the desktop log panel to display real-time CLI status logs, and upgrade Tailwind CSS to v4.

**Architecture:** 
1. Fix `tray_systray.go` compilation for Linux by setting the correct build tag.
2. Introduce a `LogHook` callback in the `cli/server` package and bind it in `desktop/app.go` to emit Wails log events.
3. Update `LogPanel.vue` to listen to these events in real-time.
4. Upgrade Tailwind CSS to v4 in Vite by installing `@tailwindcss/vite` and replacing the configuration files with a CSS-first approach.

**Tech Stack:** Go, Vue 3, Tailwind CSS v4

---

### Task 1: Fix Linux System Tray Build Tag & Theme Toggle Stub

**Files:**
- Modify: [desktop/tray_systray.go](file:///home/yule/Sides/airvoice/desktop/tray_systray.go)

- [ ] **Step 1: Update build tag in `tray_systray.go`**
  Change the build tag on line 1 of [desktop/tray_systray.go](file:///home/yule/Sides/airvoice/desktop/tray_systray.go) to target both Linux and Windows.
  
  Code change:
  ```go
  //go:build linux || windows
  ```

- [ ] **Step 2: Commit build tag change**
  ```bash
  git add desktop/tray_systray.go
  git commit -m "fix: change tray_systray build tag to include linux"
  ```

---

### Task 2: Implement Real-time Status Log Event Pipeline

**Files:**
- Modify: [cli/server/log.go](file:///home/yule/Sides/airvoice/cli/server/log.go)
- Modify: [desktop/app.go](file:///home/yule/Sides/airvoice/desktop/app.go)
- Modify: [desktop/frontend/src/components/LogPanel.vue](file:///home/yule/Sides/airvoice/desktop/frontend/src/components/LogPanel.vue)

- [ ] **Step 1: Add LogHook to `cli/server/log.go`**
  Modify [cli/server/log.go](file:///home/yule/Sides/airvoice/cli/server/log.go) to add a global callback function `LogHook`.

  Code to write:
  ```go
  package server

  import (
  	"fmt"
  	"os"
  	"strings"
  	"unicode/utf8"
  )

  var LogHook func(string)

  func logStatus(format string, args ...any) {
  	msg := fmt.Sprintf(format, args...)
  	fmt.Fprintf(os.Stderr, " [airvoice] "+msg+"\n")
  	if LogHook != nil {
  		LogHook(" [airvoice] " + msg)
  	}
  }

  func previewText(s string, maxRunes int) string {
  	s = strings.ReplaceAll(s, "\n", `\n`)
  	if utf8.RuneCountInString(s) <= maxRunes {
  		return s
  	}
  	return string([]rune(s)[:maxRunes]) + "..."
  }
  ```

- [ ] **Step 2: Bind LogHook and emit Wails events in `desktop/app.go`**
  Modify `startup` and `StartServer` in [desktop/app.go](file:///home/yule/Sides/airvoice/desktop/app.go) to wire the logs.

  Code changes:
  ```go
  // In startup():
  func (a *App) startup(ctx context.Context) {
  	a.ctx = ctx
  	server.LogHook = func(msg string) {
  		runtime.EventsEmit(a.ctx, "log_added", msg)
  	}
  	_ = a.StartServer(a.port)
  }
  ```
  ```go
  // In StartServer(), emit the initial listening log:
  	a.status = ConnectionStatus{
  		State: "waiting",
  		Port:  port,
  	}
  	if a.tray != nil {
  		a.tray.UpdateStatus()
  	}
  	a.mu.Unlock()

  	if a.ctx != nil {
  		runtime.EventsEmit(a.ctx, "log_added", fmt.Sprintf(" [airvoice] listening on :%d (health: /health, ws: /ws)", port))
  	}
  ```

- [ ] **Step 3: Update `LogPanel.vue` to listen to `'log_added'` events**
  Modify [desktop/frontend/src/components/LogPanel.vue](file:///home/yule/Sides/airvoice/desktop/frontend/src/components/LogPanel.vue) to update in real-time.

  Code to write:
  ```vue
  <script setup lang="ts">
  import { ref, onMounted, onUnmounted, nextTick } from 'vue'

  const logs = ref<string[]>([])
  const logContainer = ref<HTMLElement>()

  function addLog(message: string) {
    const time = new Date().toLocaleTimeString('en-US', { hour12: false })
    logs.value.push(`[${time}] ${message}`)
    nextTick(() => {
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }
    })
  }

  onMounted(() => {
    const runtime = (window as any).runtime
    if (runtime && runtime.EventsOn) {
      runtime.EventsOn('log_added', (msg: string) => {
        addLog(msg)
      })
    }
  })

  onUnmounted(() => {
    const runtime = (window as any).runtime
    if (runtime && runtime.EventsOff) {
      runtime.EventsOff('log_added')
    }
    logs.value = []
  })
  </script>

  <template>
    <div 
      ref="logContainer"
      class="w-full h-32 p-3 bg-bg-primary border border-border-default rounded-xl overflow-y-auto font-mono text-xs leading-relaxed"
    >
      <div v-for="(log, index) in logs" :key="index" class="text-text-muted">
        {{ log }}
      </div>
    </div>
  </template>
  ```

- [ ] **Step 4: Commit log updates**
  ```bash
  git add cli/server/log.go desktop/app.go desktop/frontend/src/components/LogPanel.vue
  git commit -m "feat: implement real-time log event forwarding and updates in desktop client"
  ```

---

### Task 3: Upgrade Tailwind CSS to v4

**Files:**
- Modify: [desktop/frontend/package.json](file:///home/yule/Sides/airvoice/desktop/frontend/package.json)
- Modify: [desktop/frontend/vite.config.ts](file:///home/yule/Sides/airvoice/desktop/frontend/vite.config.ts)
- Modify: [desktop/frontend/src/assets/styles.css](file:///home/yule/Sides/airvoice/desktop/frontend/src/assets/styles.css)
- Delete: `desktop/frontend/tailwind.config.js`
- Delete: `desktop/frontend/postcss.config.js`

- [ ] **Step 1: Install Tailwind v4 and the Vite plugin**
  Run in `desktop/frontend`:
  ```bash
  npm install tailwindcss@4 @tailwindcss/vite
  npm uninstall autoprefixer postcss
  ```

- [ ] **Step 2: Add Tailwind plugin to `vite.config.ts`**
  Modify [desktop/frontend/vite.config.ts](file:///home/yule/Sides/airvoice/desktop/frontend/vite.config.ts) to use `@tailwindcss/vite`.

  Code to write:
  ```typescript
  import { defineConfig } from 'vite'
  import vue from '@vitejs/plugin-vue'
  import tailwindcss from '@tailwindcss/vite'

  export default defineConfig({
    plugins: [
      vue(),
      tailwindcss(),
    ],
    server: {
      port: 5173,
    },
  })
  ```

- [ ] **Step 3: Update `styles.css` imports and theme**
  Modify [desktop/frontend/src/assets/styles.css](file:///home/yule/Sides/airvoice/desktop/frontend/src/assets/styles.css). Replace the `@tailwind` directives with `@import "tailwindcss";` and define the custom theme tokens in `@theme`.

  Code to write:
  ```css
  @import "tailwindcss";

  @theme {
    --color-bg-primary: var(--color-bg-primary);
    --color-bg-secondary: var(--color-bg-secondary);
    --color-border-default: var(--color-border-default);
    --color-accent-blue: var(--color-accent-blue);
    
    --color-text-primary: var(--color-primary-text);
    --color-text-secondary: var(--color-secondary-text);
    --color-text-muted: var(--color-muted-text);
    
    --color-primary-text: var(--color-primary-text);
    --color-secondary-text: var(--color-secondary-text);
    --color-muted-text: var(--color-muted-text);
    
    --color-status-success: var(--color-status-success);
    --color-status-warning: var(--color-status-warning);
    --color-status-error: var(--color-status-error);
    --color-status-neutral: var(--color-status-neutral);
    
    --radius-sm: 6px;
    --radius-md: 12px;
    --radius-lg: 16px;
    --radius-full: 9999px;
  }

  :root {
    --color-primary-text: #171717;
    --color-secondary-text: #666666;
    --color-muted-text: #888888;
    --color-bg-primary: #ffffff;
    --color-bg-secondary: #fafafa;
    --color-border-default: #eaeaea;
    --color-accent-blue: #006efe;
    --color-status-success: #28a948;
    --color-status-warning: #ffae00;
    --color-status-error: #fc0035;
    --color-status-neutral: #8f8f8f;
  }

  @media (prefers-color-scheme: dark) {
    :root:not(.light) {
      --color-primary-text: #ededed;
      --color-secondary-text: #a0a0a0;
      --color-muted-text: #666666;
      --color-bg-primary: #000000;
      --color-bg-secondary: #0d0e15;
      --color-border-default: #2e2e2e;
      --color-status-success: #00ac3a;
      --color-status-error: #e2162a;
    }
  }

  :root.dark {
    --color-primary-text: #ededed;
    --color-secondary-text: #a0a0a0;
    --color-muted-text: #666666;
    --color-bg-primary: #000000;
    --color-bg-secondary: #0d0e15;
    --color-border-default: #2e2e2e;
    --color-status-success: #00ac3a;
    --color-status-error: #e2162a;
  }

  :root.light {
    --color-primary-text: #171717;
    --color-secondary-text: #666666;
    --color-muted-text: #888888;
    --color-bg-primary: #ffffff;
    --color-bg-secondary: #fafafa;
    --color-border-default: #eaeaea;
    --color-status-success: #28a948;
    --color-status-warning: #ffae00;
    --color-status-error: #fc0035;
  }

  * {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }

  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background-color: var(--color-bg-primary);
    color: var(--color-primary-text);
    -webkit-font-smoothing: antialiased;
  }

  ::-webkit-scrollbar {
    width: 6px;
  }

  ::-webkit-scrollbar-track {
    background: var(--color-bg-secondary);
  }

  ::-webkit-scrollbar-thumb {
    background: var(--color-border-default);
    border-radius: 3px;
  }
  ```

- [ ] **Step 4: Delete config files**
  Remove the obsolete `tailwind.config.js` and `postcss.config.js` files:
  ```bash
  rm desktop/frontend/tailwind.config.js
  rm desktop/frontend/postcss.config.js
  ```

- [ ] **Step 5: Commit Tailwind v4 upgrade**
  ```bash
  git add desktop/frontend/package.json desktop/frontend/vite.config.ts desktop/frontend/src/assets/styles.css
  git rm desktop/frontend/tailwind.config.js desktop/frontend/postcss.config.js
  git commit -m "chore: upgrade tailwindcss to v4 using CSS-first configuration and @tailwindcss/vite"
  ```

---

### Task 4: Compilation and Verification

- [ ] **Step 1: Test build desktop frontend**
  Run in `desktop/frontend`:
  ```bash
  npm run build
  ```
  Expected: Successful production build of static assets with Tailwind v4.

- [ ] **Step 2: Run Go tests**
  Run: `mise run desktop:test`
  Expected: All Go tests pass.
