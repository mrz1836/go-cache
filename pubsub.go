package cache

import (
	"context"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	// pubSubMessageBufferSize is the number of messages to buffer before blocking
	pubSubMessageBufferSize = 100

	// pubSubReconnectMin is the initial backoff duration
	pubSubReconnectMin = 1 * time.Second

	// pubSubReconnectMax is the maximum backoff duration
	pubSubReconnectMax = 30 * time.Second
)

// Message represents a pub/sub message received from a Redis channel
type Message struct {
	Channel string // Channel the message was published to
	Pattern string // Pattern that matched (only set for PSubscribe messages)
	Data    []byte // Payload
}

// Subscription represents an active Redis pub/sub subscription
// Messages are delivered on the Messages channel; call Close() to unsubscribe and release resources.
type Subscription struct {
	Messages <-chan Message // Buffered (100) incoming messages; receive until closed

	client    *Client
	conn      redis.Conn
	psc       redis.PubSubConn
	channels  []string
	patterns  []string
	msgCh     chan Message
	done      chan struct{}
	closeOnce sync.Once
	errCh     chan error // internal; receives reconnection errors for visibility
}

// Publish sends a message to the given channel
// Returns the number of subscribers that received the message
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: PublishRaw()
func Publish(ctx context.Context, client *Client, channel string, message interface{}) (int64, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return 0, err
	}
	defer client.CloseConnection(conn)
	return PublishRaw(conn, channel, message)
}

// PublishRaw sends a message to the given channel
// Returns the number of subscribers that received the message
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/publish
func PublishRaw(conn redis.Conn, channel string, message interface{}) (int64, error) {
	return redis.Int64(conn.Do(PublishCommand, channel, message))
}

// Subscribe subscribes to one or more Redis channels and returns a Subscription
// The Subscription's Messages channel delivers incoming messages until Close() is called
// or the context is canceled. The subscription reconnects automatically on connection failure.
// Creates a dedicated connection (not from the pool command-cycle).
//
// Spec: https://redis.io/commands/subscribe
func Subscribe(ctx context.Context, client *Client, channels ...string) (*Subscription, error) {
	if len(channels) == 0 {
		return nil, redis.ErrNil
	}
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, err
	}

	psc := redis.PubSubConn{Conn: conn}
	if err = psc.Subscribe(toInterfaces(channels)...); err != nil {
		_ = conn.Close()
		return nil, err
	}

	sub := newSubscription(client, conn, psc, channels, nil)
	sub.start(ctx)
	return sub, nil
}

// PSubscribe subscribes to one or more Redis patterns and returns a Subscription
// The Subscription's Messages channel delivers incoming messages until Close() is called
// or the context is canceled. The subscription reconnects automatically on connection failure.
// Creates a dedicated connection (not from the pool command-cycle).
//
// Spec: https://redis.io/commands/psubscribe
func PSubscribe(ctx context.Context, client *Client, patterns ...string) (*Subscription, error) {
	if len(patterns) == 0 {
		return nil, redis.ErrNil
	}
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, err
	}

	psc := redis.PubSubConn{Conn: conn}
	if err = psc.PSubscribe(toInterfaces(patterns)...); err != nil {
		_ = conn.Close()
		return nil, err
	}

	sub := newSubscription(client, conn, psc, nil, patterns)
	sub.start(ctx)
	return sub, nil
}

// Close unsubscribes and releases all resources held by the Subscription.
// It is safe to call Close multiple times; subsequent calls are no-ops.
func (s *Subscription) Close() error {
	var closeErr error
	s.closeOnce.Do(func() {
		close(s.done)
		// Best-effort unsubscribe before closing the connection; ignore errors
		// because the connection may already be broken.
		if len(s.channels) > 0 {
			_ = s.psc.Unsubscribe(toInterfaces(s.channels)...)
		}
		if len(s.patterns) > 0 {
			_ = s.psc.PUnsubscribe(toInterfaces(s.patterns)...)
		}
		closeErr = s.conn.Close()
	})
	return closeErr
}

// newSubscription creates a Subscription struct (does not start the goroutine)
func newSubscription(client *Client, conn redis.Conn, psc redis.PubSubConn, channels, patterns []string) *Subscription {
	msgCh := make(chan Message, pubSubMessageBufferSize)
	return &Subscription{
		Messages: msgCh,
		client:   client,
		conn:     conn,
		psc:      psc,
		channels: channels,
		patterns: patterns,
		msgCh:    msgCh,
		done:     make(chan struct{}),
		errCh:    make(chan error, 16),
	}
}

// start launches the background goroutine that reads messages and handles reconnects.
// The goroutine exits when s.done is closed or when ctx is canceled.
func (s *Subscription) start(ctx context.Context) {
	go func() {
		defer close(s.msgCh)

		// Context cancellation triggers Close, which signals done.
		go func() {
			select {
			case <-ctx.Done():
				_ = s.Close()
			case <-s.done:
			}
		}()

		backoff := pubSubReconnectMin
		for {
			s.readLoop() // blocks until error or done

			// Check whether we should stop
			select {
			case <-s.done:
				return
			default:
			}

			// Connection dropped — attempt reconnect with exponential backoff
			select {
			case <-s.done:
				return
			case <-time.After(backoff):
			}

			if err := s.reconnect(ctx); err != nil {
				// Could not reconnect; report the error and advance backoff
				select {
				case s.errCh <- err:
				default:
				}
				backoff = nextBackoff(backoff)
				continue
			}

			// Reconnected — reset backoff
			backoff = pubSubReconnectMin
		}
	}()
}

// readLoop reads from the current PubSubConn until it returns an error or s.done is closed.
func (s *Subscription) readLoop() {
	for {
		select {
		case <-s.done:
			return
		default:
		}

		switch v := s.psc.Receive().(type) {
		case redis.Message:
			// Both regular (SUBSCRIBE) and pattern (PSUBSCRIBE) messages arrive as
			// redis.Message; Pattern is non-empty only for pattern-matched messages.
			msg := Message{Channel: v.Channel, Pattern: v.Pattern, Data: v.Data}
			select {
			case s.msgCh <- msg:
			case <-s.done:
				return
			}
		case redis.Subscription:
			// Subscription confirmation — no user-facing action needed
		case error:
			// Connection dropped or closed — exit readLoop so the caller can reconnect
			return
		}
	}
}

// reconnect obtains a new connection from the pool and re-subscribes to all channels/patterns.
func (s *Subscription) reconnect(ctx context.Context) error {
	conn, err := s.client.GetConnectionWithContext(ctx)
	if err != nil {
		return err
	}

	psc := redis.PubSubConn{Conn: conn}

	if len(s.channels) > 0 {
		if err = psc.Subscribe(toInterfaces(s.channels)...); err != nil {
			_ = conn.Close()
			return err
		}
	}
	if len(s.patterns) > 0 {
		if err = psc.PSubscribe(toInterfaces(s.patterns)...); err != nil {
			_ = conn.Close()
			return err
		}
	}

	// Swap in the new connection (old one is already dead)
	s.conn = conn
	s.psc = psc
	return nil
}

// toInterfaces converts a []string to []interface{} for redigo variadic calls
func toInterfaces(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

// nextBackoff doubles the backoff duration, capped at pubSubReconnectMax
func nextBackoff(d time.Duration) time.Duration {
	d *= 2
	if d > pubSubReconnectMax {
		return pubSubReconnectMax
	}
	return d
}
