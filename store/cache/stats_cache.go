package cache

import (
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/model"
)

type StatsStore interface {
	Stats(model.StatsRange) (model.Stats, error)
}

type Cache struct {
	store    StatsStore
	interval time.Duration
	done     chan struct{}
	stats    atomic.Value
}

type statsMap map[model.StatsRange]model.Stats

func NewStatsCache(s StatsStore, interval time.Duration) *Cache {
	return &Cache{
		store:    s,
		interval: interval,
		done:     make(chan struct{}),
	}
}

func (s *Cache) Run() error {
	if err := s.refresh(); err != nil {
		return fmt.Errorf("cache initialization failed: %w", err)
	}

	t := time.NewTicker(s.interval)
	for {
		select {
		case <-t.C:
			if err := s.refresh(); err != nil {
				slog.Error("cache refresh failed", slog.String("error", err.Error()))
			}
		case <-s.done:
			return nil
		}
	}
}

func (s *Cache) Stop() error {
	close(s.done)
	return nil
}

func (s *Cache) refresh() error {
	statsMap := make(statsMap, 4)
	for _, r := range model.SupportedStatsRanges() {
		stats, err := s.store.Stats(r)
		if err != nil {
			return fmt.Errorf("failed to get stats for range %s: %w", r, err)
		}
		statsMap[r] = stats
	}

	s.stats.Store(statsMap)
	slog.Info("Stats cache updated")
	return nil
}

func (s *Cache) CachedStats(r model.StatsRange) model.Stats {
	return s.stats.Load().(statsMap)[r]
}
