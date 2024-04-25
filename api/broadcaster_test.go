package api_test

import (
	"testing"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/api"
	"github.com/gevulotnetwork/devnet-explorer/model"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcasterOneClient(t *testing.T) {
	s := &MockStore{
		events: make(chan model.Event, 1000),
	}

	b := api.NewBroadcaster(s, time.Millisecond*10)

	eg := &multierror.Group{}
	eg.Go(b.Run)

	ch, unsubscribe := b.Subscribe(api.NoFilter, true)
	s.events <- model.Event{}
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Error("did not receive event")
	}

	unsubscribe()
	select {
	case <-ch:
	default:
		t.Error("ch not closed after unsubscribe")
	}

	assert.NoError(t, b.Stop())
	require.NoError(t, eg.Wait().ErrorOrNil())
}

func TestBroadcasterBuffer(t *testing.T) {
	s := &MockStore{
		events: make(chan model.Event, 1000),
	}

	b := api.NewBroadcaster(s, time.Millisecond*10)
	eg := &multierror.Group{}
	eg.Go(b.Run)

	const numOfEvents = api.BufferSize + 2
	for i := 0; i < numOfEvents; i++ {
		s.events <- model.Event{}
	}

	// Give server some time to buffer events.
	time.Sleep(time.Second)

	ch, unsubscribe := b.Subscribe(api.NoFilter, true)
	for i := 0; i < api.BufferSize; i++ {
		select {
		case <-ch:
		case <-time.After(time.Second * 5):
			t.Error("did not receive event")
		}
	}

	unsubscribe()
	select {
	case <-ch:
	default:
		t.Error("ch not closed after unsubscribe")
	}

	assert.NoError(t, b.Stop())
	require.NoError(t, eg.Wait().ErrorOrNil())
}

func TestBroadcasterStuckClient(t *testing.T) {
	s := &MockStore{
		events: make(chan model.Event, 1000),
	}

	b := api.NewBroadcaster(s, time.Millisecond*10)

	eg := &multierror.Group{}
	eg.Go(b.Run)

	const numOfEvents = 2 * api.BufferSize
	done := make(chan struct{})

	// Simulate stuck client by not reading from the channel.
	_, unsubscribe := b.Subscribe(api.NoFilter, true)
	defer unsubscribe()

	ready := make(chan struct{})

	// Receive all events regardless of the stuck client.
	go func() {
		defer close(done)
		counter := 0
		ch, unsubscribe := b.Subscribe(api.NoFilter, true)
		defer unsubscribe()
		close(ready)
		for {
			select {
			case <-ch:
				counter++
				if counter == numOfEvents {
					return
				}
			case <-time.After(time.Second * 5):
				t.Error("did not receive event")
				return
			}
		}
	}()

	<-ready
	for i := 0; i < numOfEvents; i++ {
		s.events <- model.Event{}
	}

	<-done

	assert.NoError(t, b.Stop())
	require.NoError(t, eg.Wait().ErrorOrNil())
}

func TestBroadcasterRetry(t *testing.T) {
	s := &MockStore{
		events: make(chan model.Event, 1000),
	}

	b := api.NewBroadcaster(s, time.Second)

	eg := &multierror.Group{}
	eg.Go(b.Run)

	const numOfEvents = 2 * api.BufferSize
	done := make(chan struct{})

	ready := make(chan struct{})
	// Receive all events with bit of sleep to trigger retry.
	go func() {
		defer close(done)
		counter := 0
		ch, unsubscribe := b.Subscribe(api.NoFilter, true)
		defer unsubscribe()
		close(ready)
		for {
			select {
			case <-ch:
				counter++
				if counter == numOfEvents {
					return
				}
				// Sleep to trigger retry.
				time.Sleep(5 * time.Millisecond)
			case <-time.After(time.Second * 5):
				t.Error("did not receive event")
				return
			}
		}
	}()

	<-ready
	for i := 0; i < numOfEvents; i++ {
		s.events <- model.Event{}
	}

	<-done

	assert.NoError(t, b.Stop())
	require.NoError(t, eg.Wait().ErrorOrNil())
}

type MockStore struct {
	stats        model.Stats
	searchResult []model.Event
	searchErr    error
	events       chan model.Event
}

func (m *MockStore) CachedStats(model.StatsRange) model.Stats { return m.stats }
func (m *MockStore) Events() <-chan model.Event               { return m.events }
func (m *MockStore) Search(string) ([]model.Event, error)     { return m.searchResult, m.searchErr }
