package sse_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/livebud/sse"
	"github.com/matryer/is"
)

func TestEmptyEvent(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()
	log := slog.Default()
	handler := sse.New(log)
	server := httptest.NewServer(handler)
	defer server.Close()
	stream, err := sse.Dial(log, server.URL)
	is.NoErr(err)
	defer stream.Close()
	err = handler.Publish(ctx, &sse.Event{})
	is.NoErr(err)
	event, err := stream.Next(ctx)
	is.NoErr(err)
	is.Equal(string(event.Data), "")
	is.NoErr(stream.Close())
}

func TestHandlerClient(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()
	log := slog.Default()
	handler := sse.New(log)
	server := httptest.NewServer(handler)
	defer server.Close()
	stream, err := sse.Dial(log, server.URL)
	is.NoErr(err)
	defer stream.Close()
	err = handler.Publish(ctx, &sse.Event{
		Type: "test",
		Data: []byte("hello"),
	})
	is.NoErr(err)
	event, err := stream.Next(ctx)
	is.NoErr(err)
	is.Equal(event.Type, "test")
	is.Equal(string(event.Data), "hello")
	is.NoErr(stream.Close())
}

func TestMultipleEvents(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()
	log := slog.Default()
	handler := sse.New(log)
	server := httptest.NewServer(handler)
	defer server.Close()
	stream, err := sse.Dial(log, server.URL)
	is.NoErr(err)
	defer stream.Close()
	err = handler.Publish(ctx, &sse.Event{
		Data: []byte("1"),
	})
	is.NoErr(err)
	err = handler.Publish(ctx, &sse.Event{
		Data: []byte("2"),
	})
	is.NoErr(err)
	err = handler.Publish(ctx, &sse.Event{
		Data: []byte("3"),
	})
	is.NoErr(err)
	event, err := stream.Next(ctx)
	is.NoErr(err)
	is.Equal(string(event.Data), "1")
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(200*time.Millisecond))
	defer cancel()
	// We should expect the deadline to elapse because broadcast doesn't buffer
	// events, streams must be listening already
	event, err = stream.Next(ctx)
	is.True(err != nil)
	is.True(errors.Is(err, context.DeadlineExceeded))
	is.Equal(event, nil)
	is.NoErr(stream.Close())
}

func TestMultilineData(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()
	log := slog.Default()
	handler := sse.New(log)
	server := httptest.NewServer(handler)
	defer server.Close()
	stream, err := sse.Dial(log, server.URL)
	is.NoErr(err)
	defer stream.Close()
	err = handler.Publish(ctx, &sse.Event{
		Data: []byte("1\n2\n3"),
	})
	is.NoErr(err)
	event, err := stream.Next(ctx)
	is.NoErr(err)
	is.Equal(string(event.Data), "1\n2\n3")
	is.NoErr(stream.Close())
}

func TestNoLockup(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()
	log := slog.Default()
	handler := sse.New(log)
	server := httptest.NewServer(handler)
	defer server.Close()
	stream, err := sse.Dial(log, server.URL)
	is.NoErr(err)
	defer stream.Close()
	err = handler.Publish(ctx, &sse.Event{
		Data: []byte("1"),
	})
	is.NoErr(err)
	err = handler.Publish(ctx, &sse.Event{
		Data: []byte("2"),
	})
	is.NoErr(err)
	err = handler.Publish(ctx, &sse.Event{
		Data: []byte("3"),
	})
	is.NoErr(err)
	event, err := stream.Next(ctx)
	is.NoErr(err)
	is.Equal(string(event.Data), "1")
	is.NoErr(stream.Close())
}

func TestMultipleClients(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()
	log := slog.Default()
	handler := sse.New(log)
	server := httptest.NewServer(handler)
	defer server.Close()
	stream1, err := sse.Dial(log, server.URL)
	is.NoErr(err)
	defer stream1.Close()
	stream2, err := sse.Dial(log, server.URL)
	is.NoErr(err)
	defer stream2.Close()
	err = handler.Publish(ctx, &sse.Event{
		Data: []byte("1"),
	})
	is.NoErr(err)
	event, err := stream1.Next(ctx)
	is.NoErr(err)
	is.Equal(string(event.Data), "1")
	event, err = stream2.Next(ctx)
	is.NoErr(err)
	is.Equal(string(event.Data), "1")

	is.NoErr(stream1.Close())

	err = handler.Publish(ctx, &sse.Event{
		Data: []byte("2"),
	})
	is.NoErr(err)
	event, err = stream1.Next(ctx)
	is.True(err != nil)
	is.True(errors.Is(err, sse.ErrStreamClosed))
	is.Equal(event, nil)
	event, err = stream2.Next(ctx)
	is.NoErr(err)
	is.Equal(string(event.Data), "2")
	is.NoErr(stream2.Close())

	err = handler.Publish(ctx, &sse.Event{
		Data: []byte("3"),
	})
	is.NoErr(err)
	event, err = stream1.Next(ctx)
	is.True(err != nil)
	is.True(errors.Is(err, sse.ErrStreamClosed))
	is.Equal(event, nil)
	event, err = stream2.Next(ctx)
	is.True(err != nil)
	is.True(errors.Is(err, sse.ErrStreamClosed))
	is.Equal(event, nil)
}
