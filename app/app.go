// Package app provides self-contained application business logic and signal handling.
package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/gevulotnetwork/devnet-explorer/api"
	"github.com/gevulotnetwork/devnet-explorer/store"
)

// Run starts the application and listens for OS signals to gracefully shutdown.
func Run() error {
	conf := ParseConfig()
	sigInt := make(chan os.Signal, 1)
	srv := &http.Server{
		Addr:    conf.ServerListenAddr,
		Handler: api.New(store.New()),
	}

	r := NewRunner()
	r.Go(func() error {
		signal.Notify(sigInt, os.Interrupt)
		<-sigInt
		slog.Info("SIGINT received, stopping application")
		return nil
	})

	r.Cleanup(func() error {
		close(sigInt)
		return nil
	})

	r.Go(func() error {
		slog.Info("server starting", slog.String("addr", "http://"+conf.ServerListenAddr))
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	r.Cleanup(func() error {
		return srv.Shutdown(context.Background())
	})

	return r.Wait()
}

type Config struct {
	ServerListenAddr string
}

func ParseConfig() Config {
	addr := os.Getenv("SERVER_LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8383"
	}

	return Config{ServerListenAddr: addr}
}
