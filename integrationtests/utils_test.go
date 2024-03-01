//go:build integration

package integrationtests_test

import (
	"context"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
)

func testName(test func(*testing.T)) string {
	return strings.Split(runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name(), ".")[1]
}

func buildApp(t *testing.T) {
	args := []string{"build", "-o", binPath, "-race", "-cover", "-covermode", "atomic", "../cmd/devnet-explorer"}
	env := []string{"CGO_ENABLED=1"}

	c := exec.Command("go", args...) //nolint:gosec
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(c.Environ(), env...)
	t.Log("Building app:", c.String())
	require.NoError(t, c.Run())
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

func startDB(t *testing.T) {
	compose, err := tc.NewDockerCompose("../docker-compose.yml")
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), tc.RemoveOrphans(true), tc.RemoveImagesLocal))
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	require.NoError(t, compose.Up(ctx, tc.Wait(true)))
}

func initTables(t *testing.T) {
	sql, err := os.ReadFile("../testdata/tables.sql")
	require.NoError(t, err)

	conn, err := pgx.Connect(context.Background(), "postgres://gevulot:gevulot@localhost:5432/gevulot")
	require.NoError(t, err)

	_, err = conn.Exec(context.Background(), string(sql))
	require.NoError(t, err)
	require.NoError(t, conn.Close(context.Background()))
}
