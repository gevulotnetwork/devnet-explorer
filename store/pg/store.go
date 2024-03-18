// Package pg abstracts the storage layer and provides a simple interface to work with.
package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/model"
	"github.com/go-gorp/gorp/v3"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	db     *gorp.DbMap
	events chan model.Event
	ctx    context.Context
	cancel context.CancelFunc
}

func New(dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Store{
		db:     &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}},
		events: make(chan model.Event, 1000),
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (s *Store) Run() error {
	defer close(s.events)

	conn, err := s.db.Db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get connection for listen/notify: %w", err)
	}

	return conn.Raw(func(driverConn any) error {
		conn := driverConn.(*stdlib.Conn).Conn()
		_, err := conn.Exec(context.Background(), "listen tx_events")
		if err != nil {
			return err
		}

		for {
			n, err := conn.WaitForNotification(s.ctx)
			if errors.Is(err, context.Canceled) {
				slog.Info("pg notify listener stopped by context")
				return nil
			}

			if err != nil {
				return fmt.Errorf("error occurred while waiting for notification: %w", err)
			}

			e := model.Event{}
			if err = json.Unmarshal([]byte(n.Payload), &e); err != nil {
				return fmt.Errorf("notification payload '%s': %w", n.Payload, err)
			}

			select {
			case s.events <- e:
			case <-time.After(time.Minute):
				return errors.New("timeout waiting for event to be sent")
			}
		}
	})
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

func (s *Store) Search(filter string) ([]model.Event, error) {
	const query = `` // TODO: implement search query
	var events []model.Event
	if _, err := s.db.Select(&events, query); err != nil {
		return nil, err
	}

	return events, nil
}

func (s *Store) Events() <-chan model.Event {
	return s.events
}

func (s *Store) Stop() error {
	s.cancel()
	s.db.Db.Close()
	return nil
}
