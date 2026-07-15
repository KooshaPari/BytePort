// Package notify provides a Postgres LISTEN/NOTIFY pub-sub bus.
//
// Usage
//
//	bus, _ := notify.New(ctx, connString)
//	sub := bus.Subscribe("events")
//	go func() {
//		for msg := range sub {
//			log.Printf("got: %s", msg.Payload)
//		}
//	}()
//	bus.Publish(ctx, "events", `{"hello":"world"}`)
//
// Channels are created on first use. Subscribers are local to this process;
// cross-process delivery is handled by Postgres NOTIFY delivery semantics.
package notify

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// Message represents a single notification received via Postgres NOTIFY.
type Message struct {
	Channel string `json:"channel"`
	Payload string `json:"payload"`
	PID     int32  `json:"pid"` // server PID that sent the notification
}

// Bus is a pub-sub bus backed by Postgres LISTEN/NOTIFY.
type Bus struct {
	pool   *pgxpool.Pool
	subs   map[string][]chan Message
	mu     sync.RWMutex
	log    *slog.Logger
	closed atomic.Bool
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

// New creates a Bus backed by the given connection pool.
// It does not start listening until Subscribe is called.
func New(pool *pgxpool.Pool) *Bus {
	return &Bus{
		pool: pool,
		subs: make(map[string][]chan Message),
		log:  slog.With("bus", "pg_notify"),
	}
}

// ---------------------------------------------------------------------------
// Publish
// ---------------------------------------------------------------------------

// Publish sends a NOTIFY on the given channel. The channel is created if it
// does not exist. ctx is used for the underlying query timeout.
func (b *Bus) Publish(ctx context.Context, channel, payload string) error {
	if b.closed.Load() {
		return fmt.Errorf("bus is closed")
	}
	q := fmt.Sprintf("NOTIFY %s, '%s'", pgEscape(channel), pgEscape(payload))
	_, err := b.pool.Exec(ctx, q)
	if err != nil {
		return fmt.Errorf("NOTIFY %q: %w", channel, err)
	}
	b.log.DebugContext(ctx, "published", "channel", channel, "payload_len", len(payload))
	return nil
}

// ---------------------------------------------------------------------------
// Subscribe / Unsubscribe
// ---------------------------------------------------------------------------

// Subscribe returns a receive-only channel that delivers messages on the
// given Postgres channel. The channel is auto-created with LISTEN on first
// subscription. Callers MUST read from the channel or risk blocking delivery.
func (b *Bus) Subscribe(ctx context.Context, channel string) <-chan Message {
	ch := make(chan Message, 64)

	b.mu.Lock()
	b.subs[channel] = append(b.subs[channel], ch)
	// If this is the first subscriber, start the listener.
	if len(b.subs[channel]) == 1 {
		go b.listen(ctx, channel)
	}
	b.mu.Unlock()

	b.log.InfoContext(ctx, "subscribed", "channel", channel)
	return ch
}

// Unsubscribe removes a channel subscription and closes the Go channel.
func (b *Bus) Unsubscribe(ctx context.Context, channel string, sub <-chan Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs := b.subs[channel]
	for i, s := range subs {
		if s == sub {
			b.subs[channel] = append(subs[:i], subs[i+1:]...)
			close(s)
			b.log.InfoContext(ctx, "unsubscribed", "channel", channel)
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Listener goroutine
// ---------------------------------------------------------------------------

// listen runs a dedicated connection that issues LISTEN and relays incoming
// notifications to all local subscribers for the given channel.
func (b *Bus) listen(ctx context.Context, channel string) {
	conn, err := b.pool.Acquire(ctx)
	if err != nil {
		b.log.ErrorContext(ctx, "acquire conn for listen", "channel", channel, "error", err)
		return
	}
	defer conn.Release()

	q := fmt.Sprintf("LISTEN %s", pgEscape(channel))
	if _, err := conn.Exec(ctx, q); err != nil {
		b.log.ErrorContext(ctx, "LISTEN", "channel", channel, "error", err)
		return
	}
	b.log.InfoContext(ctx, "listening", "channel", channel)

	for {
		if b.closed.Load() {
			return
		}
		// WaitForNotification blocks until a notification arrives.
		n, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			b.log.WarnContext(ctx, "wait for notification", "channel", channel, "error", err)
			return
		}
		msg := Message{
			Channel: n.Channel,
			Payload: n.Payload,
			PID:     int32(n.PID),
		}
		b.deliver(channel, msg)
	}
}

// deliver fans out a message to all local subscribers.
func (b *Bus) deliver(channel string, msg Message) {
	b.mu.RLock()
	subs := b.subs[channel]
	b.mu.RUnlock()

	for _, sub := range subs {
		select {
		case sub <- msg:
		default:
			b.log.Warn("dropping message on slow subscriber", "channel", channel)
		}
	}
}

// ---------------------------------------------------------------------------
// Admin
// ---------------------------------------------------------------------------

// Close shuts down the bus, unsubscribes all listeners, and closes channels.
func (b *Bus) Close() {
	b.closed.Store(true)
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch, subs := range b.subs {
		for _, sub := range subs {
			close(sub)
		}
		delete(b.subs, ch)
	}
	b.log.Info("bus closed")
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// pgEscape wraps a string in single quotes and doubles any embedded quotes
// to prevent SQL injection through channel names or payloads.
func pgEscape(s string) string {
	return "'" + escapeSingleQuote(s) + "'"
}

func escapeSingleQuote(s string) string {
	out := make([]byte, 0, len(s)+8)
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			out = append(out, '\'', '\'')
		} else {
			out = append(out, s[i])
		}
	}
	return string(out)
}
