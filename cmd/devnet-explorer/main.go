package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	addr := os.Getenv("SERVER_LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8383"
	}

	srv := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Hello, World!")
		}),
	}

	go func() {
		<-stop
		slog.Info("stopping server")
		err := srv.Shutdown(context.Background())
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("graceful server shutdown failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	slog.Info("server starting", slog.String("addr", "http://"+addr))
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server run failed", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("server stopped successfully")
}
