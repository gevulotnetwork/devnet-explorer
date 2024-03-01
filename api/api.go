package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/gevulotnetwork/devnet-explorer/model"
	"github.com/julienschmidt/httprouter"
)

//go:embed all:public
var public embed.FS

type Store interface {
	Stats() (model.Stats, error)
}

type API struct {
	r *httprouter.Router
	s Store
}

func New(s Store) (*API, error) {
	a := &API{
		r: httprouter.New(),
		s: s,
	}

	publicFS, err := fs.Sub(public, "public")
	if err != nil {
		return nil, fmt.Errorf("failed to create public fs: %w", err)
	}

	a.r.NotFound = http.FileServer(http.FS(publicFS))
	a.r.GET("/api/v1/stat", a.stats)

	return a, nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.r.ServeHTTP(w, r)
}

func (a *API) stats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	stats, err := a.s.Stats()
	if err != nil {
		slog.Error("failed to get stats", slog.Any("error", err))
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		slog.Error("failed encode stats", slog.Any("error", err))
		return
	}
}
