package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gevulotnetwork/devnet-explorer/model"
	"github.com/julienschmidt/httprouter"
)

type Store interface {
	Stats() (model.Stats, error)
}

type API struct {
	r *httprouter.Router
	s Store
}

func New(s Store) *API {
	a := &API{
		r: httprouter.New(),
		s: s,
	}
	a.bind()
	return a
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.r.ServeHTTP(w, r)
}

func (a *API) bind() {
	a.r.GET("/api/v1/stat", a.stats)
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
