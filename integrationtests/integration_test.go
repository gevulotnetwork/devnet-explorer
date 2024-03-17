//go:build integration

package integrationtests_test

import (
	"io"
	"net/http"
	"testing"
	"time"

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
		testEmptyStatsJSON,
		testEmptyStatsHTML,
	} {
		t.Run(testName(test), test)
	}
}

func testEmptyStatsJSON(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8383/api/v1/stats", nil)
	require.NoError(t, err)

	r.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{}).Do(r)
	require.NoError(t, err)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	const expectedResp = `{"registered_users":0,"programs":0,"proofs_generated":0,"proofs_verified":0}`
	require.JSONEq(t, expectedResp, string(data))
}

func testEmptyStatsHTML(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8383/api/v1/stats", nil)
	require.NoError(t, err)

	r.Header.Set("Accept", "test/html")
	resp, err := (&http.Client{}).Do(r)
	require.NoError(t, err)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	const expectedResp = `<div id="stats" hx-get="/api/v1/stats" hx-trigger="every 2s" hx-swap="outerHTML"><div class="number-block"><div class="rolling-number" id="registered_users">0</div><div class="number-title">Registered<br>Users</div></div><div class="number-block"><div class="rolling-number" id="provers_deployed">0</div><div class="number-title">Provers<br>Deployed</div></div><div class="number-block"><div class="rolling-number" id="proofs_generated">0</div><div class="number-title">Proofs<br>Generated</div></div><div class="number-block"><div class="rolling-number" id="proofs_verified">0</div><div class="number-title">Proofs<br>Verified</div></div></div>`
	require.Equal(t, expectedResp, string(data))
}
