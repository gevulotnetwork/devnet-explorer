package api

import (
	"embed"
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
	b *Broadcaster
}

func New(s Store, b *Broadcaster) (*API, error) {
	a := &API{
		r: http.NewServeMux(),
		s: s,
		b: b,
	}

	assetsFS, err := fs.Sub(assets, "assets")
	if err != nil {
		return nil, err
	}

	a.r.Handle("GET /", templ.Handler(templates.Index()))
	a.r.HandleFunc("GET /api/v1/stream", a.stream)
	a.r.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetsFS))))

	return a, nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.r.ServeHTTP(w, r)
}

func (a *API) stream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	slog.Info("client connected", slog.String("remote_addr", r.RemoteAddr))
	ch, unsubscribe := a.b.subscribe()
	defer unsubscribe()
	for {
		select {
		case <-r.Context().Done(): // Client disconnected
			slog.Info("client disconnected", slog.String("remote_addr", r.RemoteAddr))
			return
		case data := <-ch:
			if _, err := w.Write(data); err != nil {
				slog.Error("failed to write to client, closing connection", slog.String("remote_addr", r.RemoteAddr), slog.Any("err", err))
				return
			}
			w.(http.Flusher).Flush()
		case <-a.b.done:
			slog.Info("broadcaster stopped, closing connection", slog.String("remote_addr", r.RemoteAddr))
			return
		}
	}
}
