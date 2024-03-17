package api

import (
	"bytes"
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gevulotnetwork/devnet-explorer/api/templates"
	"github.com/gevulotnetwork/devnet-explorer/model"
)

//go:embed all:assets
var assets embed.FS

type Store interface {
	Stats() (model.Stats, error)
	Events() <-chan model.Event
}

type API struct {
	r *http.ServeMux
	s Store
}

func New(s Store) (*API, error) {
	a := &API{
		r: http.NewServeMux(),
		s: s,
	}

	assetsFS, err := fs.Sub(assets, "assets")
	if err != nil {
		return nil, err
	}

	a.r.Handle("GET /", templ.Handler(templates.Index()))
	a.r.HandleFunc("GET /api/v1/stats", a.stats)
	a.r.HandleFunc("GET /api/v1/events", a.events)
	a.r.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetsFS))))

	return a, nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.r.ServeHTTP(w, r)
}

func (a *API) stats(w http.ResponseWriter, r *http.Request) {
	stats, err := a.s.Stats()
	if err != nil {
		slog.Error("failed to get stats", slog.Any("error", err))
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			slog.Error("failed encode stats", slog.Any("error", err))
			return
		}
		return
	}

	if err := templates.Stats(stats).Render(r.Context(), w); err != nil {
		slog.Error("failed render stats", slog.Any("error", err))
		return
	}
}

func (a *API) events(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	slog.Info("client connected", slog.String("remote_addr", r.RemoteAddr))

	for {
		select {
		case <-r.Context().Done(): // Client disconnected
			slog.Info("client disconnected", slog.String("remote_addr", r.RemoteAddr))
			return
		case event, ok := <-a.s.Events():
			if !ok {
				return
			}

			buf := &bytes.Buffer{}
			buf.WriteString("data: ")
			if err := templates.Row(event).Render(r.Context(), buf); err != nil {
				slog.Error("failed render row", slog.Any("error", err))
				return
			}
			buf.WriteString("\n\n")
			if _, err := buf.WriteTo(w); err != nil {
				slog.Error("failed send event", slog.Any("error", err))
			}
			w.(http.Flusher).Flush()
		}
	}
}
