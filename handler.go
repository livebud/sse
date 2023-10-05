package sse

import (
	"context"
	"fmt"
	"net/http"
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
		pub: newPublishers(log),
		log: log,
	}
}

type Handler struct {
	Permit   func(w http.ResponseWriter, r *http.Request) bool
	Identity func(r *http.Request) string
	pub      *publishers
	log      log.Log
}

var _ http.Handler = (*Handler)(nil)
var _ Publisher = (*Handler)(nil)

func (h *Handler) Publish(ctx context.Context, event *Event) error {
	return h.pub.Publish(ctx, event)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.Permit(w, r) {
		h.log.Errorf("sse: request not permitted")
		return
	}
	publisher, err := Create(w)
	if err != nil {
		h.log.Errorf("sse: unable to create publisher: %w", err)
		http.Error(w, err.Error(), 500)
		return
	}
	// Add the client to the publisher
	clientID := h.Identity(r)
	eventCh := h.pub.Set(clientID, publisher)
	defer h.pub.Remove(clientID)
	// Wait for the client to disconnect
	ctx := r.Context()
	for {
		select {
		// Send events to the client
		case event := <-eventCh:
			publisher.Publish(ctx, event)
		// Client disconnected
		case <-ctx.Done():
			return
		}
	}
}
