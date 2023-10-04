package sse

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/livebud/log"
)

func defaultPermitter(w http.ResponseWriter, r *http.Request) bool {
	return r.Header.Get("Accept") == "text/event-stream"
}

// New server-sent event (SSE) handler
func New(log log.Log) *Handler {
	var id atomic.Int64
	return &Handler{
		Permit: defaultPermitter,
		Identity: func(r *http.Request) string {
			return fmt.Sprintf("%d", id.Add(1))
		},
		pub: newPublisher(log),
		log: log,
	}
}

type Handler struct {
	Permit   func(w http.ResponseWriter, r *http.Request) bool
	Identity func(r *http.Request) string
	pub      *publisher
	log      log.Log
}

func (h *Handler) Broadcast(ctx context.Context, event *Event) error {
	return h.pub.Broadcast(ctx, event)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.Permit(w, r) {
		h.log.Errorf("sse: request not permitted")
		return
	}
	sender, err := Create(w)
	if err != nil {
		h.log.Errorf("sse: unable to create sender: %w", err)
		http.Error(w, err.Error(), 500)
		return
	}
	// Add the client to the publisher
	clientID := h.Identity(r)
	eventCh := h.pub.Set(clientID, sender)
	defer h.pub.Remove(clientID)
	// Wait for the client to disconnect
	ctx := r.Context()
	for {
		select {
		// Send events to the client
		case event := <-eventCh:
			sender.Send(event)
		// Client disconnected
		case <-ctx.Done():
			return
		}
	}
}

type client struct {
	sender  *Sender
	eventCh chan *Event
}

func newPublisher(log log.Log) *publisher {
	return &publisher{
		log:     log,
		clients: map[string]*client{},
	}
}

type publisher struct {
	log     log.Log
	mu      sync.RWMutex
	clients map[string]*client
}

func (b *publisher) Set(id string, sender *Sender) <-chan *Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	eventCh := make(chan *Event)
	b.clients[id] = &client{
		sender:  sender,
		eventCh: eventCh,
	}
	return eventCh
}

func (b *publisher) Remove(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, id)
}

// Broadcast an event to all clients. If a client is slow to receive events,
// events will be dropped.
func (b *publisher) Broadcast(ctx context.Context, event *Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for id, client := range b.clients {
		select {
		case client.eventCh <- event:
			b.log.Debugf("sse: sent event to %s", id)
			continue
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}
