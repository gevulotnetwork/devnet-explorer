//go:build integration

package integrationtests_test

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/r3labs/sse"
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

func emptyStats(t *testing.T) {
	events := make(chan *sse.Event)
	client := sse.NewClient("http://server/events")
	client.SubscribeChan("stats", events)
	select {
	case e := <-events:
		expected := `<div id="stats" sse-swap="stats" hx-swap="outerHTML">`
		require.Contains(t, string(e.Data), expected)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}
