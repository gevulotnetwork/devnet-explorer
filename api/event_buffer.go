package api

import (
	"bytes"
	"sync"

	"github.com/gevulotnetwork/devnet-explorer/api/templates"
	"github.com/gevulotnetwork/devnet-explorer/model"
)

type eventBuffer struct {
	headMu    sync.Mutex
	headIndex int
	head      []eventData
	headMap   map[string]header
}

type eventData struct {
	txID string
	data []byte
}

type header struct {
	state model.State
	index int
}

func newEventBuffer(size uint) *eventBuffer {
	return &eventBuffer{
		head:    make([]eventData, size),
		headMap: make(map[string]header, size),
	}
}

func (b *eventBuffer) add(e model.Event, data []byte) {
	b.headMu.Lock()
	defer b.headMu.Unlock()

	data = bytes.Replace(data, []byte("event: "+e.TxID), []byte("event: "+templates.EventTXRow), 1)
	old, ok := b.headMap[e.TxID]
	if !ok {
		delete(b.headMap, b.head[b.headIndex].txID)
		b.head[b.headIndex] = eventData{txID: e.TxID, data: data}
		b.headMap[e.TxID] = header{state: e.State, index: b.headIndex}
		b.headIndex = (b.headIndex + 1) % len(b.head)
		return
	}

	if e.State.LessThan(old.state) {
		return
	}

	b.head[old.index] = eventData{txID: e.TxID, data: data}
	b.headMap[e.TxID] = header{state: e.State, index: old.index}
}

func (b *eventBuffer) writeAllToCh(ch chan<- []byte) {
	b.headMu.Lock()
	index := b.headIndex
	headCopy := make([]eventData, len(b.head))
	copy(headCopy, b.head)
	b.headMu.Unlock()

	for i := 1; i <= len(headCopy); i++ {
		idx := (index + i) % len(headCopy)
		if headCopy[idx].data != nil {
			ch <- headCopy[idx].data
		}
	}
}
