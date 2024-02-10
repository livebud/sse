package sse

import (
	"bytes"
	"strconv"
)

// Minimal logger that allows you to pass in a *slog.Logger
type logger interface {
	Debug(string, ...any)
}

// https://html.spec.whatwg.org/multipage/server-sent-events.html#event-stream-interpretation
type Event struct {
	ID    string // id (optional)
	Type  string // event type (optional)
	Data  []byte // data
	Retry int    // retry (optional)
}

func (e *Event) Format() *bytes.Buffer {
	b := new(bytes.Buffer)
	if e.ID != "" {
		b.WriteString("id: " + e.ID + "\n")
	}
	if e.Type != "" {
		b.WriteString("event: " + e.Type + "\n")
	}
	if len(e.Data) > 0 {
		b.WriteString("data: ")
		b.Write(e.Data)
		b.WriteByte('\n')
	}
	if e.Retry > 0 {
		b.WriteString("retry: " + strconv.Itoa(e.Retry) + "\n")
	}
	b.WriteByte('\n')
	return b
}

func (e *Event) String() string {
	return e.Format().String()
}
