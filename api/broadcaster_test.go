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

	b := api.NewBroadcaster(s)

	eg := &multierror.Group{}
	eg.Go(b.Run)

	ch, unsubscribe := b.Subscribe()
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

	b := api.NewBroadcaster(s)
	eg := &multierror.Group{}
	eg.Go(b.Run)

	const numOfEvents = api.BufferSize + 1
	for i := 0; i < numOfEvents; i++ {
		s.events <- model.Event{}
	}

	time.Sleep(time.Second)

	ch, unsubscribe := b.Subscribe()
	for i := 0; i < numOfEvents; i++ {
		select {
		case <-ch:
		case <-time.After(time.Second):
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

type MockStore struct {
	stats  model.Stats
	err    error
	events chan model.Event
}

func (m *MockStore) Stats() (model.Stats, error) { return m.stats, m.err }

func (m *MockStore) Events() <-chan model.Event { return m.events }
