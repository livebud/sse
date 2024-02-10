package sse

import (
	"context"
	"fmt"
	"sync"
)

type client struct {
	publisher Publisher
	eventCh   chan *Event
}

func newPublishers(log logger) *publishers {
	return &publishers{
		log:     log,
		clients: map[string]*client{},
	}
}

type publishers struct {
	log     logger
	mu      sync.RWMutex
	clients map[string]*client
}

var _ Publisher = (*publishers)(nil)

func (b *publishers) Set(id string, publisher Publisher) <-chan *Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	eventCh := make(chan *Event)
	b.clients[id] = &client{
		publisher: publisher,
		eventCh:   eventCh,
	}
	return eventCh
}

func (b *publishers) Remove(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, id)
}

// Publish an event to all clients. If a client is slow to receive events,
// events will be dropped.
func (b *publishers) Publish(ctx context.Context, event *Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for id, client := range b.clients {
		select {
		case client.eventCh <- event:
			b.log.Debug(fmt.Sprintf("sse: sent event to %s", id))
			continue
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}
