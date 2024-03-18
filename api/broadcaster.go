package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/api/templates"
	"github.com/gevulotnetwork/devnet-explorer/model"
)

const BufferSize = 50

type Broadcaster struct {
	s            Store
	clientsMu    sync.Mutex
	nextID       uint64
	clients      map[uint64]member
	headIndex    uint8
	head         [BufferSize][]byte
	retryTimeout time.Duration
	done         chan struct{}
}

type member struct {
	ch     chan<- []byte
	filter Filter
}

type Filter func(model.Event) bool

func NewBroadcaster(s Store, retryTimeout time.Duration) *Broadcaster {
	return &Broadcaster{
		s:            s,
		clients:      make(map[uint64]member),
		retryTimeout: retryTimeout,
		done:         make(chan struct{}),
	}
}

func (b *Broadcaster) Subscribe(f Filter, prefill bool) (data <-chan []byte, unsubscribe func()) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	id := b.nextID
	ch := make(chan []byte, len(b.head)+2)
	b.clients[id] = member{ch: ch, filter: f}
	b.nextID++
	slog.Info("client subscribed", slog.Uint64("id", id))

	if prefill {
		for i := 1; i <= len(b.head); i++ {
			idx := (b.headIndex + uint8(i)) % uint8(len(b.head))
			if b.head[idx] != nil {
				ch <- b.head[idx]
			}
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
	for {
		select {
		case e, ok := <-b.s.Events():
			if !ok {
				slog.Info("store.Events() channel closed, broadcasting stopped")
				return nil
			}
			b.broadcast(e)
		case <-b.done:
			return nil
		}
	}
}

func (b *Broadcaster) broadcast(e model.Event) {
	slog.Debug("new tx event received")
	buf := &bytes.Buffer{}
	if err := writeEvent(buf, e); err != nil {
		slog.Error("failed write event into buffer", slog.Any("error", err))
		return
	}
	data := buf.Bytes()

	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()
	b.head[b.headIndex] = data
	b.headIndex = (b.headIndex + 1) % uint8(len(b.head))
	blocked := make([]uint64, 0, len(b.clients))

	for id, c := range b.clients {
		if c.filter(e) {
			select {
			case c.ch <- data:
				slog.Debug("data broadcasted", slog.Uint64("id", id))
			default:
				slog.Info("client blocked, adding to retry block", slog.Uint64("id", id))
				blocked = append(blocked, id)
			}
		}
	}

	for _, id := range blocked {
		select {
		case b.clients[id].ch <- data:
			slog.Debug("data broadcasted", slog.Uint64("id", id))
		case <-time.After(b.retryTimeout):
			slog.Info("client blocked after retry, skipping", slog.Uint64("id", id))
			blocked = append(blocked, id)
		}
	}
}

func (b *Broadcaster) Stop() error {
	close(b.done)
	return nil
}

func writeEvent(w io.Writer, e model.Event) error {
	fmt.Fprintf(w, "event: %s\ndata: ", templates.EventTXRow)
	if err := templates.Row(e).Render(context.Background(), w); err != nil {
		return fmt.Errorf("failed render html: %w", err)
	}
	fmt.Fprint(w, "\n\n")
	return nil
}
