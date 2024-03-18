package main

import (
	"log/slog"
	"os"
	_ "time/tzdata"

	_ "github.com/KimMachineGun/automemlimit"
	"github.com/kelseyhightower/envconfig"
	_ "go.uber.org/automaxprocs"

	"github.com/gevulotnetwork/devnet-explorer/app"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

func main() {
	if len(os.Args) > 1 {
		envconfig.Usagef("", &app.Config{}, os.Stdout, envconfig.DefaultListFormat) // nolint: errcheck
		return
	}

	if err := app.Run(os.Args...); err != nil {
		slog.Error("running application failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("application stopped successfully")
}
