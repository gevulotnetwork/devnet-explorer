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
		testEmptyStats,
	} {
		t.Run(testName(test), test)
	}
}

func testEmptyStats(t *testing.T) {
	resp, err := http.Get("http://127.0.0.1:8383/api/v1/stats")
	require.NoError(t, err)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	const expectedResp = `{"RegisteredUsers":0,"Programs":0,"ProofsGenerated":0,"ProofsVerified":0}`
	require.JSONEq(t, expectedResp, string(data))
}
