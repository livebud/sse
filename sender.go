package sse

import (
	"fmt"
	"net/http"
)

// Create a sender from a response writer
func Create(w http.ResponseWriter) (*Sender, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("sse: response writer is not a flusher")
	}
	// Set the appropriate response headers
	headers := w.Header()
	headers.Add(`Content-Type`, `text/event-stream`)
	headers.Add(`Cache-Control`, `no-cache`)
	headers.Add(`Connection`, `keep-alive`)
	headers.Add(`Access-Control-Allow-Origin`, "*")
	// Flush the headers
	flusher.Flush()
	return &Sender{w, flusher}, nil
}

type Sender struct {
	w http.ResponseWriter
	f http.Flusher
}

func (s *Sender) Send(evt *Event) (int, error) {
	n, err := s.w.Write(evt.Format().Bytes())
	if err != nil {
		return n, err
	}
	s.f.Flush()
	return n, nil
}
