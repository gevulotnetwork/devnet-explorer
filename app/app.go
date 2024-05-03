// Package app provides self-contained application business logic and signal handling.
package app

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/api"
	"github.com/gevulotnetwork/devnet-explorer/model"
	"github.com/gevulotnetwork/devnet-explorer/signalhandler"
	"github.com/gevulotnetwork/devnet-explorer/store/cache"
	"github.com/gevulotnetwork/devnet-explorer/store/mock"
	"github.com/gevulotnetwork/devnet-explorer/store/pg"
	"github.com/kelseyhightower/envconfig"
)

type Store interface {
	Search(filter string) ([]model.Event, error)
	Stats(model.StatsRange) (model.Stats, error)
	Events() <-chan model.Event
	TxInfo(id string) (model.TxInfo, error)
	Runnable
}

type CachedStore interface {
	CachedStats(model.StatsRange) model.Stats
}

type CombinedStore struct {
	Store
	CachedStore
}

// Run starts the application and listens for OS signals to gracefully shutdown.
func Run(args ...string) error {
	slog.Info("starting application")
	conf, err := ParseConfig(args...)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: conf.LogLevel})))
	slog.Debug("Starting app with config", slog.Any("config", conf))

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

	c := cache.NewStatsCache(s, conf.StatsTTL)
	cs := CombinedStore{
		Store:       s,
		CachedStore: c,
	}

	brc := api.NewBroadcaster(cs, conf.SseRetryTimeout)
	srv, err := api.NewServer(conf.ServerListenAddr, cs, brc)
	if err != nil {
		return fmt.Errorf("failed to api server: %w", err)
	}

	sh := signalhandler.New(os.Interrupt)
	r := NewRunner(s, c, srv, brc, sh)
	return r.Run()
}

type Config struct {
	ServerListenAddr string        `envconfig:"SERVER_LISTEN_ADDR" default:"127.0.0.1:8383"`
	DSN              string        `envconfig:"DSN" default:"postgres://gevulot:gevulot@localhost:5432/gevulot"`
	MockStore        bool          `envconfig:"MOCK_STORE" default:"false"`
	StatsTTL         time.Duration `envconfig:"STATS_TTL" default:"5s"`
	SseRetryTimeout  time.Duration `envconfig:"SSE_RETRY_TIMEOUT" default:"10ms"`
	LogLevel         slog.Level    `envconfig:"LOG_LEVEL" default:"info"`
}

// TODO: Proper config parsing
func ParseConfig(args ...string) (Config, error) {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		return c, err
	}
	return c, nil
}
