// Package pg abstracts the storage layer and provides a simple interface to work with.
package pg

import (
	"database/sql"

	"github.com/gevulotnetwork/devnet-explorer/model"
	"github.com/go-gorp/gorp/v3"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	db     *gorp.DbMap
	events chan model.Event
	done   chan struct{}
}

func New(dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &Store{
		db:     &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}},
		events: make(chan model.Event, 1000),
		done:   make(chan struct{}),
	}, nil
}

func (s *Store) Run() error {
	defer close(s.events)
	eventSource := make(chan model.Event)
	for {
		select {
		case <-s.done:
			return nil
		case e := <-eventSource:
			select {
			case <-s.done:
				return nil
			case s.events <- e:
			}
		}
	}
}

func (s *Store) Stats() (model.Stats, error) {
	const query = `
	SELECT
		(SELECT COUNT(*) FROM acl_whitelist) as RegisteredUsers,
		(SELECT COUNT(*)/2 FROM program) as ProofsGenerated,
		(SELECT COUNT(*) FROM transaction WHERE kind = 'proof' AND executed IS TRUE) as ProofsGenerated,
		(SELECT COUNT(*) FROM transaction WHERE kind = 'proof' AND executed IS TRUE) as ProofsVerified;`

	var stats model.Stats
	if err := s.db.SelectOne(&stats, query); err != nil {
		return stats, err
	}

	return stats, nil
}

func (s *Store) Events() <-chan model.Event {
	return s.events
}

func (s *Store) Stop() error {
	close(s.done)
	return nil
}
