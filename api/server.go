package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type Server struct {
	srv *http.Server
}

func NewServer(addr string, s Store, b *Broadcaster) (*Server, error) {
	a, err := New(s, b)
	if err != nil {
		return nil, fmt.Errorf("failed to create api: %w", err)
	}

	return &Server{
		srv: &http.Server{
			Addr:    addr,
			Handler: a,
		},
	}, nil
}

func (s *Server) Run() error {
	slog.Info("starting server", slog.String("addr", "http://"+s.srv.Addr))
	err := s.srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	slog.Info("stopping server")
	return s.srv.Shutdown(context.Background())
}
