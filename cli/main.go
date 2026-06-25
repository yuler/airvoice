package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/airvoice/airvoice/cli/paste"
	"github.com/airvoice/airvoice/cli/server"
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
		port := 7383
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		portPtr := fs.Int("port", 7383, "port to listen on")
		if err := fs.Parse(os.Args[2:]); err == nil {
			port = *portPtr
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

		addr := fmt.Sprintf("0.0.0.0:%d", port)
		srv := server.New(server.Config{
			Addr:     addr,
			Port:     port,
			Hostname: hostname,
			Version:  version,
			Paster:   paster,
		})

		if err := srv.RotatePairing(""); err != nil {
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
	fmt.Fprintf(os.Stderr, "  airvoice serve [--port 7383]\n")
	fmt.Fprintf(os.Stderr, "  airvoice doctor\n")
	fmt.Fprintf(os.Stderr, "  airvoice version\n")
}
