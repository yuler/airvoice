package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yuler/airvoice/cli/config"
	"github.com/yuler/airvoice/cli/pairing"
	"github.com/yuler/airvoice/cli/paste"
	"github.com/yuler/airvoice/cli/relayclient"
	"github.com/yuler/airvoice/cli/server"
)

func (a *App) setStatus(status ConnectionStatus) {
	a.mu.Lock()
	a.status = status
	ctx := a.ctx
	a.mu.Unlock()
	if a.tray != nil {
		a.tray.UpdateStatus()
	}
	if ctx != nil {
		runtime.EventsEmit(ctx, "status_changed", status)
	}
}

func (a *App) ensurePaster() (paste.Paster, error) {
	a.mu.RLock()
	p := a.paster
	a.mu.RUnlock()
	if p != nil {
		return p, nil
	}

	created, err := paste.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize paster: %w", err)
	}
	a.mu.Lock()
	if a.paster == nil {
		a.paster = created
	} else {
		created = a.paster
	}
	a.mu.Unlock()
	return created, nil
}

func (a *App) startActiveTransport() error {
	a.mu.RLock()
	mode := config.NormalizeConnectionMode(a.settings.ConnectionMode)
	port := a.port
	a.mu.RUnlock()

	if mode == config.ModeRelay {
		return a.startRelay()
	}
	return a.StartServer(port)
}

func (a *App) applyCredentials(newToken, newRelayKey string, disconnect bool) {
	newToken = strings.TrimSpace(newToken)
	newRelayKey = strings.TrimSpace(newRelayKey)

	a.mu.Lock()
	tokenChanged := newToken != "" && newToken != a.token
	relayChanged := newRelayKey != "" && newRelayKey != a.relayKey
	if !tokenChanged && !relayChanged {
		a.mu.Unlock()
		return
	}
	if tokenChanged {
		a.token = newToken
		a.settings.Token = newToken
	}
	if relayChanged {
		a.relayKey = newRelayKey
		a.settings.RelayKey = newRelayKey
	}
	srv := a.server
	rc := a.relayClient
	ctx := a.ctx
	a.mu.Unlock()

	if srv != nil && tokenChanged {
		srv.SetToken(newToken)
		if disconnect {
			srv.DisconnectClients()
		}
	}
	if rc != nil && relayChanged {
		_ = rc.UpdateRelayKey(newRelayKey)
	}

	if ctx != nil {
		runtime.EventsEmit(ctx, "log_added", " [airvoice] pairing credentials refreshed")
		runtime.EventsEmit(ctx, "pairing_changed")
	}
}

func (a *App) StartServer(port int) error {
	if err := server.CheckPortAvailable(port); err != nil {
		return err
	}

	a.mu.Lock()
	if a.server != nil {
		a.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	a.port = port
	a.status = ConnectionStatus{
		State: "waiting",
		Mode:  config.ModeLAN,
		Port:  port,
	}
	ctx := a.ctx
	a.mu.Unlock()
	if a.tray != nil {
		a.tray.UpdateStatus()
	}
	if ctx != nil {
		runtime.EventsEmit(ctx, "status_changed", a.GetConnectionStatus())
	}

	paster, err := a.ensurePaster()
	if err != nil {
		a.setStatus(ConnectionStatus{State: "disconnected", Mode: config.ModeLAN})
		return err
	}

	if ctx != nil {
		runtime.EventsEmit(ctx, "log_added", fmt.Sprintf(" [airvoice] listening on :%d (health: /health, ws: /ws)", port))
	}

	srv := server.New(server.Config{
		Addr:     fmt.Sprintf(":%d", port),
		Port:     port,
		Hostname: getLocalHostname(),
		Paster:   paster,
		OnTextReceived: func(content, device string) {
			if a.history != nil {
				a.history.Add(content, device)
			}
		},
		OnConnect: func(device string) {
			host := ""
			if ip, err := pairing.LocalIPv4(); err == nil {
				host = ip
			}
			a.mu.RLock()
			curPort := a.port
			a.mu.RUnlock()
			a.setStatus(ConnectionStatus{
				State:      "connected",
				Mode:       config.ModeLAN,
				DeviceName: device,
				Host:       host,
				Port:       curPort,
			})
		},
		OnDisconnect: func() {
			a.mu.RLock()
			curPort := a.port
			a.mu.RUnlock()
			a.setStatus(ConnectionStatus{
				State: "waiting",
				Mode:  config.ModeLAN,
				Port:  curPort,
			})
		},
	})

	a.mu.Lock()
	if a.server != nil {
		a.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	token := a.token
	a.server = srv
	a.mu.Unlock()
	srv.SetToken(token)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.mu.Lock()
			if a.server == srv {
				a.server = nil
			}
			curPort := a.port
			a.mu.Unlock()
			a.setStatus(ConnectionStatus{
				State: "disconnected",
				Mode:  config.ModeLAN,
				Port:  curPort,
			})
		}
	}()

	return nil
}

func (a *App) startRelay() error {
	a.mu.RLock()
	if a.relayClient != nil {
		a.mu.RUnlock()
		return fmt.Errorf("relay client already running")
	}
	relayURL := config.ResolveRelayURL(a.settings.RelayURL, "")
	authKey := strings.TrimSpace(a.settings.RelayAuthKey)
	a.mu.RUnlock()

	relayKey, err := a.ensureRelayKey()
	if err != nil {
		return err
	}

	relayHost := relayclient.RelayHost(relayURL)
	a.setStatus(ConnectionStatus{
		State: "connecting",
		Mode:  config.ModeRelay,
		Host:  relayHost,
	})

	paster, err := a.ensurePaster()
	if err != nil {
		a.setStatus(ConnectionStatus{
			State: "error",
			Mode:  config.ModeRelay,
			Host:  relayHost,
			Error: err.Error(),
		})
		return err
	}

	var client *relayclient.Client
	client, err = relayclient.New(relayclient.Config{
		RelayURL: relayURL,
		RelayKey: relayKey,
		AuthKey:  authKey,
		Hostname: getLocalHostname(),
		Paster:   paster,
		OnTextReceived: func(content, device string) {
			if a.history != nil {
				a.history.Add(content, device)
			}
		},
		OnStateChange: func(state relayclient.RelayState, detail string) {
			a.mu.RLock()
			active := a.relayClient == client
			a.mu.RUnlock()
			if !active {
				return
			}
			status := ConnectionStatus{
				State: string(state),
				Mode:  config.ModeRelay,
				Host:  relayHost,
			}
			switch state {
			case relayclient.StateConnected:
				status.DeviceName = detail
			case relayclient.StateError:
				status.Error = detail
			}
			a.setStatus(status)
		},
		OnRelayError: func(code, message string) {
			if ctx := a.getContext(); ctx != nil {
				runtime.EventsEmit(ctx, "relay_error", map[string]string{
					"code":    code,
					"message": message,
				})
			}
		},
	})
	if err != nil {
		a.setStatus(ConnectionStatus{
			State: "error",
			Mode:  config.ModeRelay,
			Host:  relayHost,
			Error: err.Error(),
		})
		return err
	}

	a.mu.Lock()
	if a.relayClient != nil {
		a.mu.Unlock()
		return fmt.Errorf("relay client already running")
	}
	a.relayClient = client
	a.mu.Unlock()

	client.Start()
	return nil
}

// DisconnectClients disconnects active clients while keeping the active transport running.
func (a *App) DisconnectClients() {
	a.mu.RLock()
	srv := a.server
	rc := a.relayClient
	port := a.port
	relayHost := relayclient.RelayHost(config.ResolveRelayURL(a.settings.RelayURL, ""))
	a.mu.RUnlock()

	if srv != nil {
		srv.DisconnectClients()
		a.mu.RLock()
		changed := a.status.State != "waiting"
		a.mu.RUnlock()
		if changed {
			a.setStatus(ConnectionStatus{
				State: "waiting",
				Mode:  config.ModeLAN,
				Port:  port,
			})
		}
		return
	}

	if rc != nil {
		rc.Reconnect()
		a.setStatus(ConnectionStatus{
			State: "waiting",
			Mode:  config.ModeRelay,
			Host:  relayHost,
		})
	}
}

// StopServer stops whichever transport (LAN server or Relay client) is currently active.
func (a *App) StopServer() error {
	a.mu.Lock()
	srv := a.server
	rc := a.relayClient
	a.server = nil
	a.relayClient = nil
	mode := config.NormalizeConnectionMode(a.settings.ConnectionMode)
	a.status = ConnectionStatus{State: "disconnected", Mode: mode}
	status := a.status
	ctx := a.ctx
	a.mu.Unlock()

	if a.tray != nil {
		a.tray.UpdateStatus()
	}
	if ctx != nil {
		runtime.EventsEmit(ctx, "status_changed", status)
	}

	var firstErr error
	if rc != nil {
		if err := rc.Close(); err != nil {
			firstErr = err
		}
	}
	if srv != nil {
		if err := srv.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
