package cache

import (
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/model"
)

type Store interface {
	Stats() (model.Stats, error)
}

type Cache struct {
	s        Store
	stats    atomic.Value
	interval time.Duration
	stop     chan struct{}
}

func New(s Store, interval time.Duration) *Cache {
	c := &Cache{
		s:        s,
		interval: interval,
		stop:     make(chan struct{}),
	}
	c.refreshStats()
	return c
}

func (c *Cache) Stats() (model.Stats, error) {
	stats := c.stats.Load().(model.Stats)
	return stats, nil
}

func (c *Cache) Run() {
	t := time.NewTicker(c.interval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			c.refreshStats()
		case <-c.stop:
			return
		}
	}
}

func (c *Cache) refreshStats() {
	stats, err := c.s.Stats()
	if err != nil {
		slog.Error("cache failed to refresh stats", slog.Any("error", err))
		return
	}
	c.stats.Store(stats)
}

func (c *Cache) Stop() {
	close(c.stop)
}
