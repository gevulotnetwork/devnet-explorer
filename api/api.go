package api

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gevulotnetwork/devnet-explorer/api/templates"
	"github.com/gevulotnetwork/devnet-explorer/model"
)

//go:embed all:assets
var assets embed.FS

type Store interface {
	Search(filter string) ([]model.Event, error)
	Stats() (model.Stats, error)
	Events() <-chan model.Event
}

type statsCache struct {
	ttl     time.Duration
	updated time.Time
	stats   model.Stats
}

type API struct {
	r  *http.ServeMux
	s  Store
	b  *Broadcaster
	st statsCache
}

func New(s Store, b *Broadcaster, statsTTL time.Duration) (*API, error) {
	a := &API{
		r: http.NewServeMux(),
		s: s,
		b: b,
		st: statsCache{
			ttl: statsTTL,
		},
	}

	assetsFS, err := fs.Sub(assets, "assets")
	if err != nil {
		return nil, err
	}

	a.r.Handle("GET /", templ.Handler(templates.Index()))
	a.r.HandleFunc("GET /api/v1/stream", a.stream)
	a.r.HandleFunc("GET /api/v1/stats", a.stats)
	a.r.HandleFunc("GET /api/v1/events", a.table)
	a.r.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetsFS))))

	return a, nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.r.ServeHTTP(w, r)
}

func (a *API) stats(w http.ResponseWriter, r *http.Request) {
	if time.Since(a.st.updated) > time.Second*5 {
		var err error
		if a.st.stats, err = a.s.Stats(); err != nil {
			slog.Error("failed to render stats", slog.Any("err", err))
		}
	}

	if err := templates.Stats(a.st.stats, a.st.ttl).Render(r.Context(), w); err != nil {
		slog.Error("failed to render stats", slog.Any("err", err))
		return
	}
}

func (a *API) table(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	if q == "" {
		if err := templates.Table(nil, url.Values{}).Render(r.Context(), w); err != nil {
			slog.Error("failed to render stats", slog.Any("err", err))
		}
		return
	}

	events, err := a.s.Search(q)
	if err != nil {
		// Let's not return an error to the client but instead continue with empty result set.
		slog.Error("failed to search events", slog.Any("err", err))
	}

	query := url.Values{}
	query.Set("q", q)
	if len(events) > 0 {
		query.Set("since", events[0].Timestamp.Format(time.RFC3339))
	}

	if err := templates.Table(events, query).Render(r.Context(), w); err != nil {
		slog.Error("failed to render stats", slog.Any("err", err))
		return
	}
}

func (a *API) stream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	prefill := true
	filter := NoFilter
	q := strings.ToLower(r.URL.Query().Get("q"))
	since := r.URL.Query().Get("since")
	if q != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err != nil {
			slog.Error("failed to parse 'since' time, using 0 time", slog.Any("err", err))
		}
		filter = SearchFilter(q, t)
		prefill = false
	}

	slog.Info("client connected", slog.String("remote_addr", r.RemoteAddr))
	ch, unsubscribe := a.b.Subscribe(filter, prefill)
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

func SearchFilter(f string, since time.Time) Filter {
	return func(e model.Event) bool {
		return e.Timestamp.After(since) &&
			(strings.Contains(e.ProverID, f) ||
				strings.Contains(e.TxID, f) ||
				strings.Contains(e.Tag, f))
	}
}

func NoFilter(e model.Event) bool {
	return true
}
