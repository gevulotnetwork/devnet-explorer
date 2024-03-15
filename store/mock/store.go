// Package mock abstracts the storage layer and provides a simple mock storage.
package mock

import (
	"crypto/sha512"
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	stats  model.Stats
	events chan model.Event
	done   chan struct{}
}

func New() *Store {
	return &Store{
		stats:  model.Stats{},
		events: make(chan model.Event, 1000),
		done:   make(chan struct{}),
	}
}

func (s *Store) Stats() (model.Stats, error) {
	s.stats.ProversDeployed += rand.Int63n(10)
	s.stats.ProofsGenerated += rand.Int63n(10)
	s.stats.ProofsVerified += rand.Int63n(10)
	s.stats.RegisteredUsers += rand.Int63n(10)
	return s.stats, nil
}

func (s *Store) Run() error {
	defer close(s.events)
	for {
		select {
		case <-s.done:
			return nil
		case s.events <- randomEvent():
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *Store) Events() <-chan model.Event {
	return s.events
}

func (s *Store) Stop() error {
	close(s.done)
	close(s.events)
	return nil
}

func randomEvent() model.Event {
	txID := sha512.Sum512([]byte(time.Now().String()))
	proverID := sha512.Sum512([]byte(time.Now().String()))
	return model.Event{
		State:     []string{"submitted", "verifying", "proving", "complete"}[rand.Intn(4)],
		TxID:      hex.EncodeToString(txID[:]),
		ProverID:  hex.EncodeToString(proverID[:]),
		Timestamp: time.Now(),
	}
}
