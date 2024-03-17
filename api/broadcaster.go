package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/gevulotnetwork/devnet-explorer/api/templates"
	"github.com/gevulotnetwork/devnet-explorer/model"
)

const BufferSize = 50

type Broadcaster struct {
	s         Store
	clientsMu sync.Mutex
	nextID    uint64
	clients   map[uint64]chan<- []byte
	headIndex uint8
	head      [BufferSize][]byte

	done chan struct{}
}

func NewBroadcaster(s Store) *Broadcaster {
	return &Broadcaster{
		s:       s,
		clients: make(map[uint64]chan<- []byte),
		done:    make(chan struct{}),
	}
}

func (b *Broadcaster) Subscribe() (data <-chan []byte, unsubscribe func()) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	id := b.nextID
	ch := make(chan []byte, len(b.head)+2)
	b.clients[id] = ch
	b.nextID++
	slog.Info("client subscribed", slog.Uint64("id", id))

	for i := 1; i <= len(b.head); i++ {
		idx := (b.headIndex + uint8(i)) % uint8(len(b.head))
		if b.head[idx] != nil {
			ch <- b.head[idx]
		}
	}

	return ch, func() {
		slog.Info("client unsubscribed", slog.Uint64("id", id))
		b.clientsMu.Lock()
		defer b.clientsMu.Unlock()
		delete(b.clients, id)
		close(ch)
	}
}

func (b *Broadcaster) Run() error {
	t := time.NewTicker(time.Second * 2)
	for {
		var ev EventComponent
		select {
		case event, ok := <-b.s.Events():
			if !ok {
				slog.Info("store.Events() channel closed, broadcasting stopped")
				return nil
			}
			slog.Debug("new tx event received")
			ev = TXRowEvent(event)
		case <-t.C:
			stats, err := b.s.Stats()
			if err != nil {
				return fmt.Errorf("failed to get stats: %w", err)
			}
			slog.Debug("stats updated")
			ev = StatEvent(stats)
		case <-b.done:
			return nil
		}

		buf := &bytes.Buffer{}
		if err := writeEvent(buf, ev); err != nil {
			slog.Error("failed write event into buffer", slog.Any("error", err))
			continue
		}

		b.broadcast(buf.Bytes())
	}
}

func (b *Broadcaster) broadcast(data []byte) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()
	b.head[b.headIndex] = data
	b.headIndex = (b.headIndex + 1) % uint8(len(b.head))
	for id, c := range b.clients {
		select {
		case c <- data:
			slog.Debug("data broadcasted", slog.Uint64("id", id))
		default:
			slog.Info("client blocked, broadcasting event skipped", slog.Uint64("id", id))
		}
	}
}

func (b *Broadcaster) Stop() error {
	close(b.done)
	return nil
}

type EventComponent struct {
	templ.Component
	name string
}

func (e EventComponent) Name() string {
	return e.name
}

func TXRowEvent(e model.Event) EventComponent {
	return EventComponent{
		Component: templates.Row(e),
		name:      templates.EventTXRow,
	}
}

func StatEvent(s model.Stats) EventComponent {
	return EventComponent{
		Component: templates.Stats(s),
		name:      templates.EventStats,
	}
}

func writeEvent(w io.Writer, c EventComponent) error {
	fmt.Fprintf(w, "event: %s\ndata: ", c.Name())
	if err := c.Render(context.Background(), w); err != nil {
		return fmt.Errorf("failed render html: %w", err)
	}
	fmt.Fprint(w, "\n\n")
	return nil
}
