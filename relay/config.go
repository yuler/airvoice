package relay

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultAddr               = ":8787"
	DefaultMaxMessageBytes    = 16384 // 16 KiB
	DefaultRatePerMinute      = 30
	DefaultRateBurst          = 10
	DefaultDailyMessageLimit  = 1000
	DefaultMaxRoomsPerIP      = 5
	DefaultMaxGlobalRooms     = 500
	DefaultOrphanIdleTimeout  = 15 * time.Minute
	DefaultSessionIdleTimeout = 2 * time.Hour

	defaultPingInterval  = 30 * time.Second
	defaultPongWait      = 90 * time.Second
	defaultWriteTimeout  = 10 * time.Second
	defaultEvictInterval = 30 * time.Second
	defaultAuthFailLimit = 60
)

// Config configures the Airvoice zero-knowledge WebSocket relay server.
type Config struct {
	Addr               string
	AuthKey            string
	MaxMessageBytes    int64
	RatePerMinute      int
	RateBurst          int
	DailyMessageLimit  int
	MaxRoomsPerIP      int
	MaxGlobalRooms     int
	OrphanIdleTimeout  time.Duration
	SessionIdleTimeout time.Duration
	TrustProxy         bool

	// Transport & testing tuning knobs (zero uses defaults).
	PingInterval   time.Duration
	PongWait       time.Duration
	WriteTimeout   time.Duration
	EvictInterval  time.Duration
	AuthFailLimit  int
	AuthFailWindow time.Duration
	Now            func() time.Time
}

// DefaultConfig returns the official quota-limited relay server defaults.
func DefaultConfig() Config {
	return Config{
		Addr:               DefaultAddr,
		MaxMessageBytes:    DefaultMaxMessageBytes,
		RatePerMinute:      DefaultRatePerMinute,
		RateBurst:          DefaultRateBurst,
		DailyMessageLimit:  DefaultDailyMessageLimit,
		MaxRoomsPerIP:      DefaultMaxRoomsPerIP,
		MaxGlobalRooms:     DefaultMaxGlobalRooms,
		OrphanIdleTimeout:  DefaultOrphanIdleTimeout,
		SessionIdleTimeout: DefaultSessionIdleTimeout,
		TrustProxy:         false,
	}
}

// ConfigFromEnv populates Config from DefaultConfig overridden by AIRVOICE_RELAY_*
// environment variables. Invalid env values return an error.
func ConfigFromEnv() (Config, error) {
	cfg := DefaultConfig()

	if v, ok := os.LookupEnv("AIRVOICE_RELAY_ADDR"); ok && strings.TrimSpace(v) != "" {
		cfg.Addr = strings.TrimSpace(v)
	}
	if v, ok := os.LookupEnv("AIRVOICE_RELAY_AUTH_KEY"); ok {
		cfg.AuthKey = v
	}
	if err := envInt64("AIRVOICE_RELAY_MAX_MESSAGE_BYTES", &cfg.MaxMessageBytes); err != nil {
		return cfg, err
	}
	if err := envInt("AIRVOICE_RELAY_RATE_PER_MINUTE", &cfg.RatePerMinute); err != nil {
		return cfg, err
	}
	if err := envInt("AIRVOICE_RELAY_RATE_BURST", &cfg.RateBurst); err != nil {
		return cfg, err
	}
	if err := envInt("AIRVOICE_RELAY_DAILY_MESSAGE_LIMIT", &cfg.DailyMessageLimit); err != nil {
		return cfg, err
	}
	if err := envInt("AIRVOICE_RELAY_MAX_ROOMS_PER_IP", &cfg.MaxRoomsPerIP); err != nil {
		return cfg, err
	}
	if err := envInt("AIRVOICE_RELAY_MAX_GLOBAL_ROOMS", &cfg.MaxGlobalRooms); err != nil {
		return cfg, err
	}
	if err := envDuration("AIRVOICE_RELAY_ORPHAN_IDLE_TIMEOUT", &cfg.OrphanIdleTimeout); err != nil {
		return cfg, err
	}
	if err := envDuration("AIRVOICE_RELAY_SESSION_IDLE_TIMEOUT", &cfg.SessionIdleTimeout); err != nil {
		return cfg, err
	}
	if err := envBool("AIRVOICE_RELAY_TRUST_PROXY", &cfg.TrustProxy); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// BindFlags binds CLI flags onto fs, using the current fields in cfg as defaults.
func BindFlags(fs *flag.FlagSet, cfg *Config) {
	fs.StringVar(&cfg.Addr, "addr", cfg.Addr, "Listen address (env AIRVOICE_RELAY_ADDR)")
	fs.StringVar(&cfg.AuthKey, "auth-key", cfg.AuthKey, "Optional Bearer auth key for private relays (env AIRVOICE_RELAY_AUTH_KEY)")
	fs.Int64Var(&cfg.MaxMessageBytes, "max-message-bytes", cfg.MaxMessageBytes, "Maximum WebSocket text frame size in bytes (env AIRVOICE_RELAY_MAX_MESSAGE_BYTES)")
	fs.IntVar(&cfg.RatePerMinute, "rate-per-minute", cfg.RatePerMinute, "Per-room messages per minute, 0 disables (env AIRVOICE_RELAY_RATE_PER_MINUTE)")
	fs.IntVar(&cfg.RateBurst, "rate-burst", cfg.RateBurst, "Per-room rate limit burst size (env AIRVOICE_RELAY_RATE_BURST)")
	fs.IntVar(&cfg.DailyMessageLimit, "daily-message-limit", cfg.DailyMessageLimit, "Daily message cap per room and per IP, 0 disables (env AIRVOICE_RELAY_DAILY_MESSAGE_LIMIT)")
	fs.IntVar(&cfg.MaxRoomsPerIP, "max-rooms-per-ip", cfg.MaxRoomsPerIP, "Maximum active rooms per client IP, 0 disables (env AIRVOICE_RELAY_MAX_ROOMS_PER_IP)")
	fs.IntVar(&cfg.MaxGlobalRooms, "max-global-rooms", cfg.MaxGlobalRooms, "Maximum active rooms globally, 0 disables (env AIRVOICE_RELAY_MAX_GLOBAL_ROOMS)")
	fs.DurationVar(&cfg.OrphanIdleTimeout, "orphan-idle-timeout", cfg.OrphanIdleTimeout, "Evict single-peer orphan rooms after this duration (env AIRVOICE_RELAY_ORPHAN_IDLE_TIMEOUT)")
	fs.DurationVar(&cfg.SessionIdleTimeout, "session-idle-timeout", cfg.SessionIdleTimeout, "Evict inactive paired rooms after this duration (env AIRVOICE_RELAY_SESSION_IDLE_TIMEOUT)")
	fs.BoolVar(&cfg.TrustProxy, "trust-proxy", cfg.TrustProxy, "Trust X-Forwarded-For / X-Real-IP from reverse proxy (env AIRVOICE_RELAY_TRUST_PROXY)")
}

// Validate checks that all configured limits and durations are valid.
func (c Config) Validate() error {
	if strings.TrimSpace(c.Addr) == "" {
		return fmt.Errorf("relay: addr must not be empty")
	}
	if c.MaxMessageBytes <= 0 {
		return fmt.Errorf("relay: max-message-bytes must be > 0")
	}
	if c.RatePerMinute < 0 || c.RateBurst < 0 || c.DailyMessageLimit < 0 || c.MaxRoomsPerIP < 0 || c.MaxGlobalRooms < 0 {
		return fmt.Errorf("relay: numeric limits must be >= 0")
	}
	if c.OrphanIdleTimeout < 0 || c.SessionIdleTimeout < 0 {
		return fmt.Errorf("relay: idle timeouts must be >= 0")
	}
	return nil
}

func (c Config) normalized() Config {
	if c.Addr == "" {
		c.Addr = DefaultAddr
	}
	if c.MaxMessageBytes <= 0 {
		c.MaxMessageBytes = DefaultMaxMessageBytes
	}
	if c.RatePerMinute > 0 && c.RateBurst <= 0 {
		c.RateBurst = 1
	}
	if c.PingInterval <= 0 {
		c.PingInterval = defaultPingInterval
	}
	if c.PongWait <= 0 {
		c.PongWait = defaultPongWait
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = defaultWriteTimeout
	}
	if c.EvictInterval <= 0 {
		c.EvictInterval = defaultEvictInterval
		if c.OrphanIdleTimeout > 0 && c.OrphanIdleTimeout/2 < c.EvictInterval {
			c.EvictInterval = max(c.OrphanIdleTimeout/2, 10*time.Millisecond)
		}
		if c.SessionIdleTimeout > 0 && c.SessionIdleTimeout/2 < c.EvictInterval {
			c.EvictInterval = max(c.SessionIdleTimeout/2, 10*time.Millisecond)
		}
	}
	if c.AuthFailLimit == 0 {
		c.AuthFailLimit = defaultAuthFailLimit
	}
	if c.AuthFailWindow <= 0 {
		c.AuthFailWindow = time.Minute
	}
	if c.Now == nil {
		c.Now = time.Now
	}
	return c
}

func envInt(name string, dst *int) error {
	v, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(v) == "" {
		return nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return fmt.Errorf("relay: invalid %s=%q: %w", name, v, err)
	}
	*dst = n
	return nil
}

func envInt64(name string, dst *int64) error {
	v, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(v) == "" {
		return nil
	}
	n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	if err != nil {
		return fmt.Errorf("relay: invalid %s=%q: %w", name, v, err)
	}
	*dst = n
	return nil
}

func envDuration(name string, dst *time.Duration) error {
	v, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(v) == "" {
		return nil
	}
	d, err := time.ParseDuration(strings.TrimSpace(v))
	if err != nil {
		return fmt.Errorf("relay: invalid %s=%q: %w", name, v, err)
	}
	*dst = d
	return nil
}

func envBool(name string, dst *bool) error {
	v, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(v) == "" {
		return nil
	}
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return fmt.Errorf("relay: invalid %s=%q: %w", name, v, err)
	}
	*dst = b
	return nil
}
