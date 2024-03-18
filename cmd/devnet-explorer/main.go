package main

import (
	"log/slog"
	"os"
	_ "time/tzdata"

	_ "github.com/KimMachineGun/automemlimit"
	_ "go.uber.org/automaxprocs"

	"github.com/gevulotnetwork/devnet-explorer/app"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

func main() {
	slog.Info("starting application")
	if err := app.Run(); err != nil {
		slog.Error("running application failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("application stopped successfully")
}
