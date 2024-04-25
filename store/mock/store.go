// Package mock abstracts the storage layer and provides a simple mock storage.
package mock

import (
	"crypto/sha512"
	"encoding/hex"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	eventsMu sync.RWMutex
	events   []model.Event
	stats    model.Stats
	eventsCh chan model.Event
	done     chan struct{}
}

func New() *Store {
	return &Store{
		stats:    model.Stats{},
		eventsCh: make(chan model.Event, 1000),
		done:     make(chan struct{}),
	}
}

func (s *Store) Stats(string) (model.Stats, error) {
	s.stats.ProversDeployed += rand.Uint64() % 9000
	s.stats.ProofsGenerated += rand.Uint64() % 9000
	s.stats.ProofsVerified += rand.Uint64() % 9000
	s.stats.RegisteredUsers += rand.Uint64() % 9000
	s.stats.ProversDeployedDelta = rand.Float64() * 100
	s.stats.ProofsGeneratedDelta = rand.Float64() * 100
	s.stats.ProofsVerifiedDelta = rand.Float64() * 100
	s.stats.RegisteredUsersDelta = rand.Float64() * 100
	return s.stats, nil
}

func (s *Store) Run() error {
	defer close(s.eventsCh)
	for {
		e := randomEvent()
		s.eventsMu.Lock()
		s.events = append(s.events, e)
		s.eventsMu.Unlock()
		select {
		case <-s.done:
			return nil
		case s.eventsCh <- e:
		}
		time.Sleep(1 * time.Second)
	}
}

func (s *Store) Events() <-chan model.Event {
	return s.eventsCh
}

func (s *Store) Search(filter string) ([]model.Event, error) {
	s.eventsMu.RLock()
	defer s.eventsMu.RUnlock()
	events := make([]model.Event, 0, 50)
	for i := len(s.events) - 1; i >= 0; i-- {
		e := s.events[i]
		if strings.Contains(e.ProverID, filter) || strings.Contains(e.TxID, filter) || strings.Contains(e.Tag, filter) {
			events = append(events, e)
			if len(events) == 50 {
				return events, nil
			}
		}
	}
	return events, nil
}

func (s *Store) Stop() error {
	close(s.done)
	return nil
}

func randomEvent() model.Event {
	txID := sha512.Sum512([]byte(time.Now().String()))
	proverID := sha512.Sum512([]byte(time.Now().String()))
	return model.Event{
		State:     []string{"submitted", "verifying", "proving", "complete"}[rand.Intn(4)],
		Tag:       []string{"starknet", "polygon", "", "", "", "", "", ""}[rand.Intn(8)],
		TxID:      hex.EncodeToString(txID[:]),
		ProverID:  hex.EncodeToString(proverID[:]),
		Timestamp: time.Now().Local(),
	}
}
