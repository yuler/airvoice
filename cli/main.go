package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/yuler/airvoice/cli/config"
	"github.com/yuler/airvoice/cli/pairing"
	"github.com/yuler/airvoice/cli/paste"
	"github.com/yuler/airvoice/cli/relayclient"
	"github.com/yuler/airvoice/cli/server"
	"github.com/yuler/airvoice/cli/upgrade"
)

// version is injected by release builds via -ldflags. go install leaves it as "dev"
// and productVersion falls back to the module version embedded in the binary.
var version = "dev"

var (
	serveFn   = runServer
	upgradeFn = runUpgrade
)

type serveOptions struct {
	port         int
	relay        bool
	relayURL     string
	relayAuthKey string
}

func (o serveOptions) useRelay() bool {
	return o.relay || strings.TrimSpace(o.relayURL) != ""
}

func parseServeFlags(args []string, errOut io.Writer) (serveOptions, error) {
	var opts serveOptions
	fs := flag.NewFlagSet("airvoice", flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.Usage = func() { writeUsage(errOut) }
	fs.IntVar(&opts.port, "port", 7654, "port to listen on (LAN mode)")
	fs.IntVar(&opts.port, "p", 7654, "port to listen on (shorthand)")
	fs.BoolVar(&opts.relay, "relay", false, "enable E2E cross-LAN relay mode")
	fs.StringVar(&opts.relayURL, "relay-url", "", "relay WebSocket URL (implicitly enables relay mode)")
	fs.StringVar(&opts.relayAuthKey, "relay-auth-key", "", "optional Bearer auth key for private relay server")
	if err := fs.Parse(args); err != nil {
		return serveOptions{}, err
	}
	return opts, nil
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) > 0 && isHelp(args[0]) {
		writeUsage(os.Stdout)
		return 0
	}

	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return serveFn(args)
	}

	switch args[0] {
	case "upgrade":
		return upgradeFn(args[1:])
	case "version":
		fmt.Println("airvoice " + productVersion())
		return 0
	case "doctor":
		return paste.PrintDoctor(os.Stdout)
	case "settings":
		return runSettings()
	case "token":
		if len(args) < 2 || args[1] != "refresh" {
			printUsage()
			return 1
		}
		return runTokenRefresh()
	default:
		printUsage()
		return 1
	}
}

func runUpgrade(args []string) int {
	var checkOnly bool
	var targetVersion string

	fs := flag.NewFlagSet("airvoice upgrade", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  airvoice upgrade [--check] [--version X.Y.Z]\n")
	}
	fs.BoolVar(&checkOnly, "check", false, "check for updates without installing")
	fs.StringVar(&targetVersion, "version", "", "install a specific version (e.g. 0.4.8)")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() > 0 {
		fs.Usage()
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := upgrade.Run(ctx, upgrade.Options{
		CurrentVersion: productVersion(),
		TargetVersion:  targetVersion,
		CheckOnly:      checkOnly,
		Out:            os.Stdout,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func runSettings() int {
	path := config.Path()
	s, err := config.EnsureToken(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	display := s
	if strings.TrimSpace(display.RelayAuthKey) != "" {
		display.RelayAuthKey = "********"
	}
	data, err := json.MarshalIndent(display, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Println(path)
	fmt.Println(string(data))
	return 0
}

func runTokenRefresh() int {
	s, err := config.RotateToken(config.Path())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Println(s.Token)
	return 0
}

func runServer(args []string) int {
	opts, err := parseServeFlags(args, os.Stderr)
	if err != nil {
		return 1
	}
	if opts.useRelay() {
		return runRelayServer(opts)
	}
	return runLANServer(opts.port)
}

func runLANServer(port int) int {
	if err := server.CheckPortAvailable(port); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	paster, err := paste.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing paster: %v\n", err)
		return 1
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "PC"
	}

	s, err := config.EnsureToken(config.Path())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading settings: %v\n", err)
		return 1
	}
	token := s.Token

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	srv := server.New(server.Config{
		Addr:     addr,
		Port:     port,
		Hostname: hostname,
		Version:  productVersion(),
		Paster:   paster,
	})
	srv.SetToken(token)

	if _, err := pairing.PrintPairingWithToken(port, token, "", productVersion()); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pairing session: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "  Paste backend: %s\n", paster.Name())
	fmt.Fprintf(os.Stderr, "  [airvoice] listening on %s (health: /health, ws: /ws)\n\n", addr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startUpdateCheck(ctx)

	go config.WatchToken(ctx, config.Path(), time.Second, token, func(newToken string) {
		srv.SetToken(newToken)
		srv.DisconnectClients()
		fmt.Fprintf(os.Stderr, "  [airvoice] pairing token refreshed\n")
		if _, err := pairing.PrintPairingWithToken(port, newToken, "", productVersion()); err != nil {
			fmt.Fprintf(os.Stderr, "Error reprinting pairing: %v\n", err)
		}
	})

	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
		return 1
	}
	return 0
}

func runRelayServer(opts serveOptions) int {
	settingsPath := config.Path()
	s, err := config.EnsureRelayKey(settingsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading settings: %v\n", err)
		return 1
	}

	relayURL := config.ResolveRelayURL(opts.relayURL, s.RelayURL)
	if err := relayclient.ValidateRelayURL(relayURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	authKey := strings.TrimSpace(opts.relayAuthKey)
	if authKey == "" {
		authKey = strings.TrimSpace(s.RelayAuthKey)
	}

	paster, err := paste.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing paster: %v\n", err)
		return 1
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "PC"
	}

	client, err := relayclient.New(relayclient.Config{
		RelayURL: relayURL,
		RelayKey: s.RelayKey,
		AuthKey:  authKey,
		Hostname: hostname,
		Version:  productVersion(),
		Paster:   paster,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing relay client: %v\n", err)
		return 1
	}

	if _, err := pairing.PrintRelayPairing(relayURL, s.RelayKey, authKey, "", productVersion()); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating relay pairing session: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "  Paste backend: %s\n\n", paster.Name())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startUpdateCheck(ctx)

	currentKey := s.RelayKey
	go config.WatchCredentials(ctx, settingsPath, time.Second, s.Token, currentKey, func(updated config.Settings) {
		newKey := strings.TrimSpace(updated.RelayKey)
		if newKey == "" || newKey == currentKey {
			return
		}
		if err := client.UpdateRelayKey(newKey); err != nil {
			fmt.Fprintf(os.Stderr, "Error applying refreshed relay key: %v\n", err)
			return
		}
		currentKey = newKey
		fmt.Fprintf(os.Stderr, "  [airvoice] pairing credentials refreshed\n")
		if _, err := pairing.PrintRelayPairing(relayURL, newKey, authKey, "", productVersion()); err != nil {
			fmt.Fprintf(os.Stderr, "Error reprinting relay pairing: %v\n", err)
		}
	})

	if err := client.Run(ctx); err != nil {
		var authErr *relayclient.AuthError
		if errors.As(err, &authErr) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", authErr)
		} else {
			fmt.Fprintf(os.Stderr, "Relay client failed: %v\n", err)
		}
		return 1
	}
	return 0
}

func startUpdateCheck(ctx context.Context) {
	execPath, _ := os.Executable()
	upgrade.StartBackgroundCheck(ctx, upgrade.BackgroundOptions{
		CurrentVersion: productVersion(),
		ExecPath:       execPath,
		Notify: func(current, latest, hint string) {
			fmt.Fprintf(os.Stderr, "  [airvoice] update available: v%s -> v%s (run `%s`)\n", current, latest, hint)
		},
	})
}

func isHelp(arg string) bool {
	return arg == "help" || arg == "--help" || arg == "-h"
}

func printUsage() {
	writeUsage(os.Stderr)
}

func writeUsage(w io.Writer) {
	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "  airvoice [--port 7654]\n")
	fmt.Fprintf(w, "  airvoice --relay [--relay-url wss://...] [--relay-auth-key <key>]\n")
	fmt.Fprintf(w, "  airvoice --relay-url wss://... [--relay-auth-key <key>]\n")
	fmt.Fprintf(w, "  airvoice upgrade [--check] [--version X.Y.Z]\n")
	fmt.Fprintf(w, "  airvoice settings\n")
	fmt.Fprintf(w, "  airvoice token refresh\n")
	fmt.Fprintf(w, "  airvoice doctor\n")
	fmt.Fprintf(w, "  airvoice version\n")
	fmt.Fprintf(w, "  airvoice help\n")
}
