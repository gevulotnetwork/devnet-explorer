package main

import (
	"log/slog"
	"os"

	"github.com/gevulotnetwork/devnet-explorer/app"
)

func main() {
	slog.Info("starting application")
	if err := app.Run(); err != nil {
		slog.Error("running application failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("application stopped successfully")
}
