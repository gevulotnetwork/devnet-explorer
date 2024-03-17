//go:build integration

package integrationtests_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/r3labs/sse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	binPath  = "../target/test-artifacts/bin/devnet-explorer"
	coverDir = "../target/test-artifacts/coverage/bin/int"
)

func TestIntegration(t *testing.T) {
	buildApp(t)
	startDB(t)
	initTables(t)
	runApp(t)
	time.Sleep(1 * time.Second)

	for _, test := range []func(*testing.T){
		index,
		receiveStats,
		receiveFirstEvent,
		receiveEventsFromBuffer,
	} {
		t.Run(testName(test), test)
	}
}

func index(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8383/", nil)
	require.NoError(t, err)

	resp, err := (&http.Client{}).Do(r)
	require.NoError(t, err)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	const expectedResp = `<div id="container" hx-ext="sse" sse-connect="/api/v1/stream">`
	require.Contains(t, string(data), expectedResp)
}

func receiveStats(t *testing.T) {
	events := sseClient(t, "stats")
	select {
	case e := <-events:
		expected := `<div id="stats" sse-swap="stats" hx-swap="outerHTML">`
		require.Contains(t, string(e.Data), expected)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func receiveFirstEvent(t *testing.T) {
	events := sseClient(t, "tx-row")
	notify(t, `{"state": "submitted","tx_id": "1234","prover_id": "5678","timestamp": "2006-01-02T15:04:05Z"}`)

	select {
	case e := <-events:
		expected := `<div class="tr"><div class="left"><div class="td">`
		require.Contains(t, string(e.Data), expected)
	case <-time.After(time.Second * 5):
		t.Fatal("timeout")
	}
}

func receiveEventsFromBuffer(t *testing.T) {
	txs := []string{
		`{"state": "submitted","tx_id": "1234","prover_id": "5678","timestamp": "2006-01-02T15:04:05Z"}`,
		`{"state": "submitted","tx_id": "1234","prover_id": "5678","timestamp": "2006-01-02T15:04:05Z"}`,
		`{"state": "submitted","tx_id": "1234","prover_id": "5678","timestamp": "2006-01-02T15:04:05Z"}`,
		`{"state": "submitted","tx_id": "1234","prover_id": "5678","timestamp": "2006-01-02T15:04:05Z"}`,
		`{"state": "submitted","tx_id": "1234","prover_id": "5678","timestamp": "2006-01-02T15:04:05Z"}`,
		`{"state": "submitted","tx_id": "1234","prover_id": "5678","timestamp": "2006-01-02T15:04:05Z"}`,
		`{"state": "submitted","tx_id": "1234","prover_id": "5678","timestamp": "2006-01-02T15:04:05Z"}`,
		`{"state": "submitted","tx_id": "1234","prover_id": "5678","timestamp": "2006-01-02T15:04:05Z"}`,
	}

	notify(t, txs...)

	// Giver server some time to buffer events before starting sse client
	time.Sleep(time.Second)
	events := sseClient(t, "tx-row")

	expectedEvents := len(txs) + 1 // +1 for event added by receiveFirstEvent
	for i := 0; i < expectedEvents; i++ {
		select {
		case e := <-events:
			expected := `<div class="tr"><div class="left"><div class="td">`
			assert.Contains(t, string(e.Data), expected)
		case <-time.After(time.Second * 5):
			t.Fatal("timeout")
		}
	}
}

func sseClient(t *testing.T, event string) chan *sse.Event {
	events := make(chan *sse.Event, 100)
	client := sse.NewClient("http://127.0.0.1:8383/api/v1/stream")
	go func() {
		err := client.SubscribeRaw(func(msg *sse.Event) {
			if string(msg.Event) == event {
				select {
				case events <- msg:
				default:
				}
			}
		})
		assert.NoError(t, err)
	}()
	t.Cleanup(func() { client.Unsubscribe(events) })
	return events
}

func notify(t *testing.T, events ...string) {
	conn, err := pgx.Connect(context.Background(), "postgres://gevulot:gevulot@localhost:5432/gevulot")
	require.NoError(t, err)
	for _, e := range events {
		_, err = conn.Exec(context.Background(), fmt.Sprintf("NOTIFY tx_events, '%s';", e))
		require.NoError(t, err)
	}
}
