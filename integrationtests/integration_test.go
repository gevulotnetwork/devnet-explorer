//go:build integration

package integrationtests_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	binPath  = "../target/test-artifacts/bin/devnet-explorer"
	coverDir = "../target/test-artifacts/coverage/bin/int"
)

func TestMain(m *testing.M) {
	if err := buildApp(); err != nil {
		log.Fatal("Failed to build test binary", err)
	}

	code := m.Run()
	if code != 0 {
		os.Exit(1)
	}
}

func TestIntegration(t *testing.T) {
	runApp(t)
	time.Sleep(1 * time.Second)

	resp, err := http.Get("http://127.0.0.1:8383")
	require.NoError(t, err)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, "Hello, World!", string(data))
}

func buildApp() error {
	args := []string{"build", "-o", binPath, "-race", "-cover", "-covermode", "atomic", "../cmd/devnet-explorer"}
	env := []string{"CGO_ENABLED=1"}

	c := exec.Command("go", args...) //nolint:gosec
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(c.Environ(), env...)
	fmt.Println("Building app:", c.String())
	return c.Run()
}

func runApp(t *testing.T, runArgs ...string) {
	err := os.MkdirAll(coverDir, 0o755)
	require.NoError(t, err)

	c := exec.Command(binPath, runArgs...) //nolint:gosec
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(c.Environ(), "GOCOVERDIR="+coverDir)

	t.Log("Running app:", c.String())
	require.NoError(t, c.Start())
	t.Cleanup(func() {
		require.NoError(t, c.Process.Signal(os.Interrupt))
		require.NoError(t, c.Wait())
	})
}
