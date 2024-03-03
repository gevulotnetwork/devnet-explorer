// Package mock abstracts the storage layer and provides a simple mock storage.
package mock

import (
	"math/rand"

	"github.com/gevulotnetwork/devnet-explorer/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	stats model.Stats
}

func New() *Store {
	return &Store{
		stats: model.Stats{},
	}
}

func (s *Store) Stats() (model.Stats, error) {
	s.stats.Programs += rand.Int63n(10)
	s.stats.ProofsGenerated += rand.Int63n(10)
	s.stats.ProofsVerified += rand.Int63n(10)
	s.stats.RegisteredUsers += rand.Int63n(10)
	return s.stats, nil
}
