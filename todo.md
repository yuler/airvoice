这个模式（**Wayland + 第一次只有 `text id`、第二次完整 OK**）基本可以把原因收窄到：**第一次 paste 卡在 `ydotool`，第二次时环境已经「热」起来了**。

---

## 为什么第二次一定 OK

Wayland 的 paste 链路是：

```11:16:cli/paste/linux_wayland.go
func (w *waylandPaster) Paste(text string) error {
	if err := runCommand("wl-copy", text); err != nil {
		return err
	}
	time.Sleep(80 * time.Millisecond)
	return runCommand("ydotool", "", "key", "CTRL+v")
}
```

第二次完整日志说明：

- `wl-copy` 正常
- `ydotool` 正常
- WebSocket 连接正常
- iOS 能收到 ack

所以不是协议或配置整体有问题，而是 **第一次 `ydotool` 调用特别慢或卡住**。

---

## 第一次为什么卡在 `text id` 后面

打出 `text id` 之后，handler 在 goroutine 里跑 `Paste()`。在打出 `paste ok` / `paste failed` 之前，**不会有任何后续日志**。

第一次常见情况：

### 1. `ydotool` 冷启动（最可能）

README 里也写了需要 `ydotoold` 在跑。第一次 `ydotool key CTRL+v` 可能要：

- 连 `ydotoold` socket
- 等 daemon 起来
- 等 `/dev/uinput` 权限

这段时间可能 **超过 iOS 的 3 秒 ack 超时**，甚至超过 CLI 的 4 秒命令超时（部分环境下 `CommandContext` 杀不掉卡住的 `ydotool` 子进程）。

你第二次手动发时，`ydotoold` 已经热了，整条链路 1 秒内跑完 → 日志完整、手机也收到。

### 2. 自动发送更容易撞上「刚连上 + 第一次 paste」

自动发送常在两种时机触发：

- 连上后立刻 `flushPendingAutoSend`
- 说完话 **1.5 秒**就发

这时往往是 **连接后第一条 `text`**，正好撞上 `ydotool` 冷启动。  
手动第二次发时，连接已稳定、`ydotool` 已热，所以几乎总是 OK。

### 3. 并发 paste（曾去掉 `pasteMu`）

之前有 commit 为了修 ydotool 问题 **去掉了 paste 串行锁**。如果第一次 goroutine 还卡在 `ydotool`，你又手动发了第二次，会出现 **两个 paste 并行**：

- 第一次：一直卡在 `ydotool`，日志停在 `text id`
- 第二次：`ydotool` 已热，完整 `paste ok` → `ack ok`

这和你描述的现象完全一致。

---

## 时间线（推测）

```
第一次（自动发）:
  text id=...          ← CLI 收到
  [ydotool 冷启动，卡住 3~N 秒]
  iOS: 「发送超时」      ← 3 秒到了
  [可能断连 / 切后台]
  [第一条 goroutine 可能还在卡，或很晚才 paste failed]

第二次（手动发）:
  text id=...
  paste ok id=...
  ack ok id=...        ← 完整链路
```

---

## 建议验证（2 分钟）

连上后 **先等 5 秒再自动发**，看第一次是否也变完整。  
若明显好转 → 就是冷启动 + 超时问题。

另在终端单独测：

```bash
wl-copy "test"
ydotool key CTRL:v
```

第一次慢、第二次快 → 确认是 `ydotool` 冷启动。

---

## 修复方向（按优先级）

| 方向 | 做什么 |
|------|--------|
| **CLI 预热** | `serve` 启动或收到 `hello` 时跑一次 `ydotool key`  noop，把 daemon 拉起来 |
| **串行 paste** | 恢复 `pasteMu`，避免第一次卡住时第二条并发进来 |
| **iOS 超时** | ack 超时调到 **6–8 秒**（大于 paste 最长 4×2 秒） |
| **连接后延迟** | 连上后 **500ms–1s** 再允许自动发，别立刻 `flushPendingAutoSend` |
| **日志** | goroutine 入口打 `paste started id=...`，方便区分「卡住」和「断连」 |

---

**结论**：Wayland 上这是 **`ydotool` 第一次冷启动慢/卡住** + **自动发送太早触发** + **iOS 3 秒超时太短** 叠在一起。第二次手动发时 `ydotool` 已热，所以链路完整。

需要的话我可以直接改：CLI 加 `ydotool` 预热、恢复 paste 串行、iOS 超时调到 6s。