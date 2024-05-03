// Package pg abstracts the storage layer and provides a simple interface to work with.
package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/model"
	"github.com/go-gorp/gorp/v3"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type gevulotTransaction struct {
	author     string
	hash       string
	kind       string
	nonce      int
	signature  string
	propagated bool
	executed   bool
	created_at time.Time
}

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
		_, err := conn.Exec(context.Background(), "listen dashboard_data_stream")
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

			slog.Debug("received notification", slog.String("payload", n.Payload))
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

// Stats returns stats for the given time range.
func (s *Store) Stats(r model.StatsRange) (model.Stats, error) {
	// TODO: Get percentages for stats.
	// TODO: Get stats only for the given range.

	const query = `
	SELECT
		(SELECT COUNT(*) FROM acl_whitelist) as RegisteredUsers,
		(SELECT COUNT(DISTINCT(prover)) FROM deploy) as ProversDeployed,
		(SELECT COUNT(*) FROM transaction WHERE kind = 'proof') as ProofsGenerated,
		(SELECT COUNT(*) FROM transaction WHERE kind = 'verification') as ProofsVerified;`

	var stats model.Stats
	if err := s.db.SelectOne(&stats, query); err != nil {
		return stats, err
	}

	return stats, nil
}

func (s *Store) Search(filter string) ([]model.Event, error) {
	filter = strings.TrimSpace(filter)

	// filter string: free text search input straight from the user, handle as such.
	// This query should return 50 most recent matching events sorted by timestamp in newest first order.
	const query = `
		 WITH t2 AS ((
			(SELECT created_at, t.hash, t.kind FROM transaction AS t WHERE t.hash = $1)
			UNION ALL
			(SELECT created_at, t.hash, t.kind FROM transaction AS t JOIN workflow_step AS ws ON ws.tx=t.hash WHERE ws.sequence=1 AND ws.program = $1)
			UNION ALL
			(SELECT created_at, t.hash, t.kind FROM transaction AS t JOIN proof AS p ON t.hash = p.tx WHERE p.prover = $1)
			UNION ALL
			(SELECT created_at, t.hash, t.kind FROM transaction AS t JOIN verification AS v ON t.hash = v.tx JOIN proof AS p ON v.parent = p.tx WHERE p.prover = $1)
		) ORDER BY created_at DESC LIMIT 50)
		SELECT
			CASE WHEN t2.kind = 'run' THEN
					CASE WHEN (SELECT COUNT(*) FROM proof WHERE parent = t2.hash) = 0 THEN 'Submitted'
						WHEN ((SELECT COUNT(*) FROM proof WHERE parent = t2.hash) >= 1 AND (SELECT COUNT(*) FROM verification AS v JOIN proof AS p ON v.parent=p.tx WHERE p.parent = t2.hash) = 0) IS TRUE THEN 'Proving'
						WHEN (SELECT COUNT(*) FROM verification AS v JOIN proof AS p ON v.parent=p.tx WHERE p.parent = t2.hash) = 1 THEN 'Verifying'
						WHEN (SELECT COUNT(*) FROM verification AS v JOIN proof AS p ON v.parent=p.tx WHERE p.parent = t2.hash) = 2 THEN 'Verifying'
						WHEN (SELECT COUNT(*) FROM verification AS v JOIN proof AS p ON v.parent=p.tx WHERE p.parent = t2.hash) > 2 THEN 'Complete'
					END
				WHEN t2.kind = 'proof' THEN
					CASE WHEN (SELECT COUNT(*) FROM proof WHERE parent = (SELECT parent FROM proof WHERE tx = t2.hash)) = 1 THEN 'Proving'
						WHEN (SELECT COUNT(*) FROM verification AS v JOIN proof AS p ON v.parent=p.tx WHERE p.parent = (SELECT parent FROM proof WHERE tx = t2.hash)) = 1 THEN 'Verifying'
						WHEN (SELECT COUNT(*) FROM verification AS v JOIN proof AS p ON v.parent=p.tx WHERE p.parent = (SELECT parent FROM proof WHERE tx = t2.hash)) = 2 THEN 'Verifying'
						WHEN (SELECT COUNT(*) FROM verification AS v JOIN proof AS p ON v.parent=p.tx WHERE p.parent = (SELECT parent FROM proof WHERE tx = t2.hash)) > 2 THEN 'Complete'
					END
				WHEN t2.kind = 'verification' THEN
					CASE WHEN (SELECT COUNT(*) FROM verification AS v JOIN proof AS p ON v.parent=p.tx WHERE p.parent = (SELECT p2.parent FROM proof AS p2 JOIN verification AS v2 ON p2.tx=v2.parent WHERE v2.tx = t2.hash)) = 1 THEN 'Verifying'
						WHEN (SELECT COUNT(*) FROM verification AS v JOIN proof AS p ON v.parent=p.tx WHERE p.parent = (SELECT p2.parent FROM proof AS p2 JOIN verification AS v2 ON p2.tx=v2.parent WHERE v2.tx = t2.hash)) = 2 THEN 'Verifying'
						WHEN (SELECT COUNT(*) FROM verification AS v JOIN proof AS p ON v.parent=p.tx WHERE p.parent = (SELECT p2.parent FROM proof AS p2 JOIN verification AS v2 ON p2.tx=v2.parent WHERE v2.tx = t2.hash)) > 2 THEN 'Complete'
					END
			END AS state,
			t.hash AS tx_id,
			(
				(SELECT name AS tag FROM program AS p JOIN workflow_step AS ws ON p.hash = ws.program JOIN t2 ON ws.tx = t2.hash WHERE ws.sequence = 1)
			UNION
				(SELECT name AS tag FROM program WHERE hash = $1)
			),
			(
				(SELECT program AS prover_id FROM workflow_step AS ws JOIN t2 ON ws.tx = t2.hash WHERE ws.sequence = 1)
			UNION
				(SELECT hash AS prover_id FROM program WHERE hash = $1)
			),
			t.created_at AS timestamp FROM transaction AS t JOIN t2 ON t.hash = t2.hash
		`

	var events []model.Event

	if _, err := s.db.Select(&events, query, filter); err != nil {
		return nil, err
	}

	return events, nil
}

func (s *Store) TxInfo(id string) (model.TxInfo, error) {
	var tx gevulotTransaction
	const fetchTxQuery = `SELECT * FROM transaction WHERE hash = $1`
	if _, err := s.db.Select(&tx, fetchTxQuery, id); err != nil {
		return model.TxInfo{}, err
	}

	const fetchTxsQuery = `
		SELECT t.* FROM transaction AS t WHERE t.hash = $1
		UNION
		SELECT t.* FROM transaction AS t JOIN proof AS p ON p.tx=t.hash WHERE p.parent = $1
		UNION
		SELECT t.* FROM transaction AS t JOIN verification AS v on v.tx=t.hash JOIN proof AS p ON v.parent = p.tx WHERE p.parent = $1
	`
	var txHash string
	switch tx.kind {
	case "run":
		txHash = tx.hash
	case "proof":
		{
			const fetchProofParent = `SELECT parent FROM proof WHERE tx = $1`
			if _, err := s.db.Select(&txHash, fetchProofParent, tx.hash); err != nil {
				return model.TxInfo{}, err
			}
		}
	case "verification":
		{
			const fetchVerificationParent = `SELECT parent FROM proof JOIN verification ON proof.tx = verification.parent WHERE verification.tx = $1`
			if _, err := s.db.Select(&txHash, fetchVerificationParent, tx.hash); err != nil {
				return model.TxInfo{}, err
			}
		}
	default:
		return model.TxInfo{}, fmt.Errorf("invalid transaction kind: %q", tx.kind)
	}

	var txs []gevulotTransaction
	if _, err := s.db.Select(txs, fetchTxsQuery, txHash); err != nil {
		return model.TxInfo{}, err
	}

	findProverProgramHashQuery := `SELECT ws.program FROM workflow_step AS ws WHERE ws.tx = $1`
	var proverHash string
	if _, err := s.db.Select(&proverHash, findProverProgramHashQuery, txHash); err != nil {
		return model.TxInfo{}, err
	}

	var author string
	for _, tx := range txs {
		if tx.kind == "run" {
			author = tx.author
			break
		}
	}

	info := model.TxInfo{
		State:    getState(txs),
		Duration: getJobDuration(txs),
		TxID:     txHash,
		UserID:   author,
		ProverID: proverHash,
		Log:      txLogEventsFromTxs(txs),
	}

	return info, nil
}

func (s *Store) Events() <-chan model.Event {
	return s.events
}

func (s *Store) Stop() error {
	s.cancel()
	s.db.Db.Close()
	return nil
}

func getJobDuration(txs []gevulotTransaction) time.Duration {
	var begin time.Time
	var end time.Time

	for _, tx := range txs {
		if begin.IsZero() || tx.created_at.Before(begin) {
			begin = tx.created_at
		}

		if end.IsZero() || tx.created_at.After(end) {
			end = tx.created_at
		}
	}

	return end.Sub(begin)
}

func getState(txs []gevulotTransaction) string {
	// Submitted is the default state.
	state := "Submitted"

	proofs := 0
	verifications := 0

	for _, tx := range txs {
		if tx.kind == "proof" {
			proofs++
		} else if tx.kind == "verification" {
			verifications++
		}
	}

	if proofs > 0 && verifications == 0 {
		state = "Proving"
	} else if verifications > 0 && verifications < 3 {
		state = "Verifying"
	} else if verifications > 2 {
		state = "Complete"
	}

	return state
}

func txLogEventsFromTxs(txs []gevulotTransaction) []model.TxLogEvent {
	stateFromKind := func(k string) string {
		switch k {
		case "run":
			return "Submitted"
		case "proof":
			return "Proving"
		case "verification":
			return "Verifying"
		}
		return ""
	}

	var events []model.TxLogEvent
	for _, tx := range txs {
		e := model.TxLogEvent{
			State:     stateFromKind(tx.kind),
			IDType:    "node id",
			ID:        tx.author,
			Timestamp: tx.created_at,
		}

		// Special handling for the Run tx.
		if e.State == "Submitted" {
			e.IDType = "user id"
		}
		events = append(events, e)
	}

	// Sort events.
	slices.SortFunc(events, func(a, b model.TxLogEvent) int {
		return a.Timestamp.Compare(b.Timestamp)
	})

	// Finalize run job as complete after two verifications.
	verifications := 0
	for _, e := range events {
		if e.State == "Verifying" {
			verifications++
			if verifications > 2 {
				e.State = "Complete"
			}
		}
	}

	return events
}
