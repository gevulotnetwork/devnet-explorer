package api

import (
	"bytes"

	"github.com/gevulotnetwork/devnet-explorer/api/templates"
	"github.com/gevulotnetwork/devnet-explorer/model"
)

type eventBuffer struct {
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
	data = bytes.Replace(data, []byte("event: "+e.TxID), []byte("event: "+templates.EventTXRow), 1)
	old, ok := b.headMap[e.TxID]
	if !ok {
		delete(b.headMap, b.head[b.headIndex].txID)
		b.head[b.headIndex] = eventData{txID: e.TxID, data: data}
		b.headMap[e.TxID] = header{state: e.State, index: b.headIndex}
		b.headIndex = (b.headIndex + 1) % len(b.head)
		return
	}

	if e.State < old.state {
		return
	}

	b.head[old.index] = eventData{txID: e.TxID, data: data}
	b.headMap[e.TxID] = header{state: e.State, index: old.index}
}

func (b *eventBuffer) writeAllToCh(ch chan<- []byte) {
	for i := 1; i <= len(b.head); i++ {
		data := b.head[(b.headIndex+i)%len(b.head)].data
		if data != nil {
			ch <- data
		}
	}
}
