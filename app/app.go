// Package app provides self-contained application business logic and signal handling.
package app

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/api"
	"github.com/gevulotnetwork/devnet-explorer/signalhandler"
	"github.com/gevulotnetwork/devnet-explorer/store/mock"
	"github.com/gevulotnetwork/devnet-explorer/store/pg"
)

type Store interface {
	api.Store
	Runnable
}

// Run starts the application and listens for OS signals to gracefully shutdown.
func Run(args ...string) error {
	conf := ParseConfig(args...)

	var s Store
	if conf.MockStore {
		s = mock.New()
	} else {
		var err error
		s, err = pg.New(conf.DSN)
		if err != nil {
			return fmt.Errorf("failed to create store: %w", err)
		}
	}

	brc := api.NewBroadcaster(s, conf.SseRetryTimeout, conf.CacheRefreshInterval)
	srv, err := api.NewServer(conf.ServerListenAddr, s, brc)
	if err != nil {
		return fmt.Errorf("failed to api server: %w", err)
	}

	sh := signalhandler.New(os.Interrupt)
	r := NewRunner(s, srv, brc, sh)
	return r.Run()
}

type Config struct {
	ServerListenAddr     string
	DSN                  string
	MockStore            bool
	CacheRefreshInterval time.Duration
	SseRetryTimeout      time.Duration
}

// TODO: Proper config parsing
func ParseConfig(args ...string) Config {
	addr := os.Getenv("SERVER_LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8383"
	}

	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "postgres://gevulot:gevulot@localhost:5432/gevulot"
	}

	cacheRefreshInterval := os.Getenv("CACHE_REFRESH_INTERVAL")
	if cacheRefreshInterval == "" {
		cacheRefreshInterval = "5s"
	}

	d1, err := time.ParseDuration(cacheRefreshInterval)
	if err != nil {
		slog.Error("failed to parse cache refresh interval, defaulting to 5s", slog.Any("error", err))
		d1 = 5 * time.Second
	}

	sseRetryTimeout := os.Getenv("SSE_RETRY_INTERVAL")
	if sseRetryTimeout == "" {
		sseRetryTimeout = "10ms"
	}

	d2, err := time.ParseDuration(sseRetryTimeout)
	if err != nil {
		slog.Error("failed to parse sse retry timeout, defaulting to 10ms", slog.Any("error", err))
		d2 = 10 * time.Millisecond
	}

	mockStore, _ := strconv.ParseBool(os.Getenv("MOCK_STORE"))

	return Config{
		ServerListenAddr:     addr,
		DSN:                  dsn,
		MockStore:            mockStore,
		CacheRefreshInterval: d1,
		SseRetryTimeout:      d2,
	}
}
