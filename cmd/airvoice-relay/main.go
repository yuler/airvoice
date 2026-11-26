package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yuler/airvoice/relay"
)

var version = "dev"

func main() {
	cfg, err := relay.ConfigFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	showVersion := flag.Bool("version", false, "Print version and exit")
	relay.BindFlags(flag.CommandLine, &cfg)
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	srv, err := relay.NewServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	authMode := "open"
	if cfg.AuthKey != "" {
		authMode = "bearer"
	}
	log.Printf(
		"[airvoice-relay] version=%s addr=%s auth=%s max_bytes=%d rate=%d/m(burst=%d) daily=%d rooms_per_ip=%d global_rooms=%d orphan=%s session=%s trust_proxy=%v",
		version, cfg.Addr, authMode, cfg.MaxMessageBytes, cfg.RatePerMinute, cfg.RateBurst,
		cfg.DailyMessageLimit, cfg.MaxRoomsPerIP, cfg.MaxGlobalRooms,
		cfg.OrphanIdleTimeout, cfg.SessionIdleTimeout, cfg.TrustProxy,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Printf("[airvoice-relay] server error: %v", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		log.Printf("[airvoice-relay] shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Close(shutdownCtx); err != nil {
			log.Printf("[airvoice-relay] shutdown error: %v", err)
			os.Exit(1)
		}
	}
}
