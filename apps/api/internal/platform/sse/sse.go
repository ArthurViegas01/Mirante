// Package sse is a single-process Server-Sent Events hub (ADR-0002). Event ids
// are durable (the monitor's events-outbox row id), so a client reconnecting
// with Last-Event-ID replays missed events from the database — surviving server
// restarts. A slow consumer is disconnected, never allowed to block the
// publisher; on reconnect it re-hydrates via REST + replay.
package sse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Event is a streamed event with a durable monotonic id.
type Event struct {
	ID   int64
	Type string
	Data []byte
}

// ReplayFunc returns events with id greater than afterID, oldest first.
type ReplayFunc func(ctx context.Context, afterID int64, limit int) ([]Event, error)

type client struct {
	ch     chan Event
	closed chan struct{}
	once   sync.Once
}

func (c *client) close() { c.once.Do(func() { close(c.closed) }) }

// Hub broadcasts events to connected SSE clients.
type Hub struct {
	replay     ReplayFunc
	maxClients int
	bufSize    int

	mu      sync.Mutex
	clients map[*client]struct{}
}

// NewHub builds a hub. replay may be nil (no DB-backed replay).
func NewHub(replay ReplayFunc) *Hub {
	return &Hub{
		replay:     replay,
		maxClients: 32,
		bufSize:    64,
		clients:    make(map[*client]struct{}),
	}
}

// Emit implements monitor.EventSink. It broadcasts to all clients, dropping
// (disconnecting) any whose buffer is full so a slow consumer never blocks the
// publisher. Evictions happen after releasing the lock (no lock upgrade).
func (h *Hub) Emit(id int64, eventType string, data []byte) {
	ev := Event{ID: id, Type: eventType, Data: data}
	var slow []*client
	h.mu.Lock()
	for c := range h.clients {
		select {
		case c.ch <- ev:
		default:
			slow = append(slow, c)
			delete(h.clients, c)
		}
	}
	h.mu.Unlock()
	for _, c := range slow {
		c.close()
	}
}

func (h *Hub) add(c *client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.clients) >= h.maxClients {
		return false
	}
	h.clients[c] = struct{}{}
	return true
}

func (h *Hub) remove(c *client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	c.close()
}

// Clients returns the current connection count (for tests/metrics).
func (h *Hub) Clients() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

// ServeHTTP streams events as text/event-stream.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	c := &client{ch: make(chan Event, h.bufSize), closed: make(chan struct{})}
	if !h.add(c) {
		http.Error(w, "too many streams", http.StatusServiceUnavailable)
		return
	}
	defer h.remove(c)

	header := w.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	// Replay events missed since the client's Last-Event-ID (durable).
	last := lastEventID(r)
	if last > 0 && h.replay != nil {
		if events, err := h.replay(r.Context(), last, 500); err == nil {
			for _, ev := range events {
				writeEvent(w, ev)
				last = ev.ID
			}
		}
	}
	_, _ = fmt.Fprint(w, "retry: 3000\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-c.closed:
			return
		case ev := <-c.ch:
			if ev.ID <= last {
				continue // already delivered via replay
			}
			writeEvent(w, ev)
			last = ev.ID
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

func writeEvent(w io.Writer, ev Event) {
	_, _ = fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", ev.ID, ev.Type, ev.Data)
}

func lastEventID(r *http.Request) int64 {
	v := r.Header.Get("Last-Event-ID")
	if v == "" {
		v = r.URL.Query().Get("lastEventId")
	}
	if v == "" {
		return 0
	}
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil || id < 0 {
		return 0
	}
	return id
}
