package stats

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/model"
)

type Store interface {
	LatestDailyStats() (model.Stats, error)
	AggregateStats(time.Time) error
}

type Aggregator struct {
	store Store
	done  chan struct{}
}

func NewAggregator(store Store) *Aggregator {
	return &Aggregator{
		store: store,
		done:  make(chan struct{}),
	}
}

func (a *Aggregator) Run() error {
	s, err := a.store.LatestDailyStats()
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return fmt.Errorf("failed to get latest aggregated stats: %w", err)
	}
	lastRan := s.CreatedAt

	t := time.NewTicker(time.Minute)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			now := time.Now()
			if now.Day() > lastRan.Day() {
				slog.Info("aggregating stats", slog.Time("last_ran", lastRan), slog.Time("now", now))
				if err := a.store.AggregateStats(now); err != nil {
					slog.Error("failed to aggregate stats", slog.String("error", err.Error()))
					continue
				}
				lastRan = now
			}

		case <-a.done:
			return nil
		}
	}
}

func (a *Aggregator) Stop() error {
	close(a.done)
	return nil
}
