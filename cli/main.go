package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yuler/airvoice/cli/pairing"
	"github.com/yuler/airvoice/cli/paste"
	"github.com/yuler/airvoice/cli/server"
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
	case "doctor":
		os.Exit(paste.PrintDoctor(os.Stdout))
	case "serve":
		var port int
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		fs.IntVar(&port, "port", 7654, "port to listen on")
		fs.IntVar(&port, "p", 7654, "port to listen on (shorthand)")
		_ = fs.Parse(os.Args[2:])

		if err := server.CheckPortAvailable(port); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}


		paster, err := paste.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing paster: %v\n", err)
			os.Exit(1)
		}

		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "PC"
		}

		// Generate a stable token for the entire process lifetime.
		token := uuid.NewString()

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
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "  Paste backend: %s\n", paster.Name())
		fmt.Fprintf(os.Stderr, "  [airvoice] listening on %s (health: /health, ws: /ws)\n\n", addr)

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
	fmt.Fprintf(os.Stderr, "  airvoice serve [--port 7654]\n")
	fmt.Fprintf(os.Stderr, "  airvoice doctor\n")
	fmt.Fprintf(os.Stderr, "  airvoice version\n")
}
