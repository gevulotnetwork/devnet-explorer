package api_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/api"
	"github.com/gevulotnetwork/devnet-explorer/app"
	"github.com/gevulotnetwork/devnet-explorer/model"
	"github.com/hashicorp/go-multierror"
	"github.com/r3labs/sse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerMultipleClients(t *testing.T) {
	const (
		clients     = 100
		numOfEvents = 10 * api.BufferSize
	)

	s := &MockStore{events: make(chan model.Event, numOfEvents+1)}
	b := api.NewBroadcaster(s, time.Second)
	srv, err := api.NewServer("127.0.0.1:7645", s, b)
	require.NoError(t, err)
	r := app.NewRunner(b, srv)

	eg := &multierror.Group{}
	eg.Go(r.Run)

	wg := &sync.WaitGroup{}
	done := make(chan struct{}, clients)
	for i := 0; i < clients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			consume(t, numOfEvents, done)
		}()
	}

	// Give some time for all clients to start.
	time.Sleep(time.Second * 1)

	for i := 0; i < numOfEvents; i++ {
		s.events <- model.Event{TxID: fmt.Sprint(i), State: model.StateProving}
	}

	// Wait that at least half of the clients receive all events.
	ready := 0
	for range done {
		if ready++; ready == clients {
			break
		}
	}

	r.Stop()
	wg.Wait()

	t.Logf("%d clients out of %d received all events", ready, clients)
	require.NoError(t, eg.Wait().ErrorOrNil())
}

func consume(t *testing.T, num int, done chan struct{}) {
	counter := 0
	client := sse.NewClient("http://127.0.0.1:7645/api/v1/stream")
	err := client.SubscribeRaw(func(msg *sse.Event) {
		counter++
		if counter == num {
			done <- struct{}{}
		}
	})
	assert.NoError(t, err)
}
