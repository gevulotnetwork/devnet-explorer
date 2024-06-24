// Package mock abstracts the storage layer and provides a simple mock storage.
package mock

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	Parallelism = 20
)

type Store struct {
	eventsMu   sync.RWMutex
	events     []model.Event
	eventQueue [Parallelism]model.Event
	eventMap   map[string]model.TxInfo
	stats      model.CombinedStats
	eventsCh   chan model.Event
	done       chan struct{}
}

func New() *Store {
	return &Store{
		eventsCh: make(chan model.Event, 1000),
		eventMap: make(map[string]model.TxInfo),
		done:     make(chan struct{}),
	}
}

func (s *Store) Stats(model.StatsRange) (model.CombinedStats, error) {
	s.stats.Stats.ProversDeployed += rand.Uint64() % 9000
	s.stats.Stats.ProofsGenerated += rand.Uint64() % 9000
	s.stats.Stats.ProofsVerified += rand.Uint64() % 9000
	s.stats.Stats.RegisteredUsers += rand.Uint64() % 9000
	s.stats.DeltaStats.ProversDeployed = rand.Float64() * 100
	s.stats.DeltaStats.ProofsGenerated = rand.Float64() * 100
	s.stats.DeltaStats.ProofsVerified = rand.Float64() * 100
	s.stats.DeltaStats.RegisteredUsers = rand.Float64() * 100
	return s.stats, nil
}

func (s *Store) Run() error {
	defer close(s.eventsCh)
	for {
		e := s.nextEvent()
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

func (s *Store) TxInfo(id string) (model.TxInfo, error) {
	info, ok := s.eventMap[id]
	if !ok {
		return model.TxInfo{}, fmt.Errorf("tx %s not found", id)
	}
	return info, nil
}

func (s *Store) Stop() error {
	close(s.done)
	return nil
}

func (s *Store) nextEvent() model.Event {
	i := rand.Intn(Parallelism)
	switch s.eventQueue[i].State {
	case model.StateSubmitted:
		s.eventQueue[i].State = model.StateProving
		s.eventQueue[i].Timestamp = time.Now().Local()

		info := s.eventMap[s.eventQueue[i].TxID]
		info.Log = append(s.eventMap[s.eventQueue[i].TxID].Log, model.TxLogEvent{
			State:     model.StateProving,
			IDType:    "node id",
			ID:        s.eventQueue[i].ProverID,
			Timestamp: s.eventQueue[i].Timestamp,
		})
		s.eventMap[s.eventQueue[i].TxID] = info

	case model.StateProving:
		s.eventQueue[i].State = model.StateVerifying
		s.eventQueue[i].Timestamp = time.Now().Local()

		verifierID := sha512.Sum512([]byte(time.Now().String()))
		info := s.eventMap[s.eventQueue[i].TxID]
		info.Log = append(s.eventMap[s.eventQueue[i].TxID].Log, model.TxLogEvent{
			State:     model.StateVerifying,
			IDType:    "node id",
			ID:        hex.EncodeToString(verifierID[:]),
			Timestamp: s.eventQueue[i].Timestamp,
		})
		s.eventMap[s.eventQueue[i].TxID] = info

	case model.StateVerifying:
		s.eventQueue[i].State = []model.State{model.StateVerifying, model.StateVerifying, model.StateVerifying, model.StateComplete}[rand.Intn(4)]
		s.eventQueue[i].Timestamp = time.Now().Local()

		info := s.eventMap[s.eventQueue[i].TxID]
		id := sha512.Sum512([]byte(time.Now().String()))
		log := model.TxLogEvent{
			IDType:    "node id",
			ID:        hex.EncodeToString(id[:]),
			Timestamp: s.eventQueue[i].Timestamp,
		}

		if s.eventQueue[i].State == model.StateComplete {
			info.Duration = s.eventQueue[i].Timestamp.Sub(info.Log[0].Timestamp)
			info.State = model.StateComplete
			log.State = model.StateComplete
		} else {
			log.State = model.StateVerifying
		}

		info.Log = append(s.eventMap[s.eventQueue[i].TxID].Log, log)
		s.eventMap[s.eventQueue[i].TxID] = info

	case model.StateUnknown, model.StateComplete:
		s.eventQueue[i] = randomEvent()

		userID := sha512.Sum512([]byte(time.Now().String()))
		s.eventMap[s.eventQueue[i].TxID] = model.TxInfo{
			State:    model.StateProving,
			TxID:     s.eventQueue[i].TxID,
			UserID:   hex.EncodeToString(userID[:]),
			ProverID: s.eventQueue[i].ProverID,
			Log: []model.TxLogEvent{
				{
					State:     model.StateSubmitted,
					IDType:    "user id",
					ID:        hex.EncodeToString(userID[:]),
					Timestamp: s.eventQueue[i].Timestamp,
				},
			},
		}
	}

	return s.eventQueue[i]
}

func randomEvent() model.Event {
	txID := sha512.Sum512([]byte(time.Now().String()))
	proverID := sha512.Sum512([]byte(time.Now().String()))
	return model.Event{
		State:     model.StateSubmitted,
		Tag:       []string{"starknet", "polygon", "", "", "", "", "", ""}[rand.Intn(8)],
		TxID:      hex.EncodeToString(txID[:]),
		ProverID:  hex.EncodeToString(proverID[:]),
		Timestamp: time.Now().Local(),
	}
}
