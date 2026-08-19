package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yuler/airvoice/cli/config"
	"github.com/yuler/airvoice/cli/pairing"
	"github.com/yuler/airvoice/cli/paste"
	"github.com/yuler/airvoice/cli/server"
)

const version = "0.3.3"

var serveFn = runServer

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return serveFn(args)
	}

	switch args[0] {
	case "version":
		fmt.Println("airvoice " + version)
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

func runSettings() int {
	s, err := config.EnsureToken(config.Path())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
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
	var port int
	fs := flag.NewFlagSet("airvoice", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.IntVar(&port, "port", 7654, "port to listen on")
	fs.IntVar(&port, "p", 7654, "port to listen on (shorthand)")
	if err := fs.Parse(args); err != nil {
		return 1
	}

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
		Version:  version,
		Paster:   paster,
	})
	srv.SetToken(token)

	if _, err := pairing.PrintPairingWithToken(port, token, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pairing session: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "  Paste backend: %s\n", paster.Name())
	fmt.Fprintf(os.Stderr, "  [airvoice] listening on %s (health: /health, ws: /ws)\n\n", addr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go config.WatchToken(ctx, config.Path(), time.Second, token, func(newToken string) {
		srv.SetToken(newToken)
		srv.DisconnectClients()
		fmt.Fprintf(os.Stderr, "  [airvoice] pairing token refreshed\n")
		if _, err := pairing.PrintPairingWithToken(port, newToken, ""); err != nil {
			fmt.Fprintf(os.Stderr, "Error reprinting pairing: %v\n", err)
		}
	})

	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
		return 1
	}
	return 0
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  airvoice [--port 7654]\n")
	fmt.Fprintf(os.Stderr, "  airvoice settings\n")
	fmt.Fprintf(os.Stderr, "  airvoice token refresh\n")
	fmt.Fprintf(os.Stderr, "  airvoice doctor\n")
	fmt.Fprintf(os.Stderr, "  airvoice version\n")
}
