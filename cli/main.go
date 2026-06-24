package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/airvoice/airvoice/cli/pairing"
	"github.com/airvoice/airvoice/cli/paste"
	"github.com/airvoice/airvoice/cli/server"
	"github.com/google/uuid"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "version":
		fmt.Println("airvoice " + version)
	case "serve":
		port := 7383
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		portPtr := fs.Int("port", 7383, "port to listen on")
		if err := fs.Parse(os.Args[2:]); err == nil {
			port = *portPtr
		}

		token := uuid.NewString()
		ip, err := pairing.LocalIPv4()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting local IP: %v\n", err)
			os.Exit(1)
		}

		wsURL := fmt.Sprintf("ws://%s:%d/ws", ip, port)

		payload := &pairing.Payload{
			Version: 1,
			WS:      wsURL,
			Token:   token,
		}

		payloadBytes, err := payload.Marshal()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling pairing payload: %v\n", err)
			os.Exit(1)
		}

		// Print QR code to stderr
		pairing.PrintQR(payloadBytes)

		paster, err := paste.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing paster: %v\n", err)
			os.Exit(1)
		}

		// Log backend, token, and ws URL to stderr
		fmt.Fprintf(os.Stderr, "Paste backend: %s\n", paster.Name())
		fmt.Fprintf(os.Stderr, "Token: %s\n", token)
		fmt.Fprintf(os.Stderr, "WebSocket URL: %s\n", wsURL)

		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "PC"
		}

		addr := fmt.Sprintf("0.0.0.0:%d", port)
		srv := server.New(server.Config{
			Addr:     addr,
			Token:    token,
			Hostname: hostname,
			Version:  version,
			Paster:   paster,
		})

		fmt.Fprintf(os.Stderr, "[airvoice] listening on %s (health: /health, ws: /ws)\n", addr)
		fmt.Fprintf(os.Stderr, "[airvoice] waiting for iPhone connection...\n")

		if err := srv.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
			os.Exit(1)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  airvoice serve [--port 7383]\n")
	fmt.Fprintf(os.Stderr, "  airvoice version\n")
}
