package cache

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

var errUnexpectedSubscribeType = errors.New("unexpected subscribe confirmation type")

// SubscriptionOption configures a Subscription at creation time.
type SubscriptionOption func(*subscriptionOptions)

type subscriptionOptions struct {
	messageBufferSize int
}

func defaultSubscriptionOptions() subscriptionOptions {
	return subscriptionOptions{messageBufferSize: pubSubMessageBufferSize}
}

// WithMessageBuffer sets the capacity of the Messages channel.
// Values less than 1 are ignored and the default (100) is used.
func WithMessageBuffer(n int) SubscriptionOption {
	return func(o *subscriptionOptions) {
		if n >= 1 {
			o.messageBufferSize = n
		}
	}
}

const (
	// pubSubMessageBufferSize is the default number of messages to buffer before blocking
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

// Subscription represents an active Redis pub/sub subscription.
// Messages are delivered on the Messages channel; call Close() to unsubscribe and release resources.
type Subscription struct {
	Messages <-chan Message // Buffered incoming messages; receive until closed
	Errors   <-chan error   // Buffered reconnection errors; non-blocking — excess errors are dropped

	client    *Client
	conn      redis.Conn
	psc       redis.PubSubConn
	channels  []string
	patterns  []string
	msgCh     chan Message
	done      chan struct{}
	closeOnce sync.Once
	connOnce  sync.Once      // guards conn.Close(); separate from closeOnce to avoid deadlock
	wg        sync.WaitGroup // tracks the readLoop goroutine; Wait() before conn.Close()
	errCh     chan error     // internal; receives reconnection errors for visibility
}

// Publish sends a message to the given channel.
// Returns the number of subscribers that received the message.
// Creates a new connection and closes connection at end of function call.
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

// PublishRaw sends a message to the given channel.
// Returns the number of subscribers that received the message.
// Uses existing connection (does not close connection).
//
// Spec: https://redis.io/commands/publish
func PublishRaw(conn redis.Conn, channel string, message interface{}) (int64, error) {
	return redis.Int64(conn.Do(PublishCommand, channel, message))
}

// Subscribe subscribes to one or more Redis channels and returns a Subscription.
// The Subscription's Messages channel delivers incoming messages until Close() is called
// or the context is canceled. The subscription reconnects automatically on connection failure.
// Creates a dedicated connection (not from the pool command-cycle).
//
// Spec: https://redis.io/commands/subscribe
func Subscribe(ctx context.Context, client *Client, channels []string, opts ...SubscriptionOption) (*Subscription, error) {
	if len(channels) == 0 {
		return nil, redis.ErrNil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
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

	// Read one subscription confirmation per channel to guarantee the subscription
	// is registered with Redis before returning. Without this, callers that publish
	// immediately after Subscribe may lose messages.
	for range channels {
		v := psc.Receive()
		if _, ok := v.(redis.Subscription); !ok {
			_ = conn.Close()
			if err, ok2 := v.(error); ok2 {
				return nil, err
			}
			return nil, fmt.Errorf("%w: %T", errUnexpectedSubscribeType, v)
		}
	}

	o := defaultSubscriptionOptions()
	for _, opt := range opts {
		opt(&o)
	}
	sub := newSubscription(client, conn, psc, channels, nil, o.messageBufferSize)
	sub.start(ctx)
	return sub, nil
}

// PSubscribe subscribes to one or more Redis patterns and returns a Subscription.
// The Subscription's Messages channel delivers incoming messages until Close() is called
// or the context is canceled. The subscription reconnects automatically on connection failure.
// Creates a dedicated connection (not from the pool command-cycle).
//
// Spec: https://redis.io/commands/psubscribe
func PSubscribe(ctx context.Context, client *Client, patterns []string, opts ...SubscriptionOption) (*Subscription, error) {
	if len(patterns) == 0 {
		return nil, redis.ErrNil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
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

	// Read one subscription confirmation per pattern to guarantee the subscription
	// is registered with Redis before returning.
	for range patterns {
		v := psc.Receive()
		if _, ok := v.(redis.Subscription); !ok {
			_ = conn.Close()
			if err, ok2 := v.(error); ok2 {
				return nil, err
			}
			return nil, fmt.Errorf("%w: %T", errUnexpectedSubscribeType, v)
		}
	}

	o := defaultSubscriptionOptions()
	for _, opt := range opts {
		opt(&o)
	}
	sub := newSubscription(client, conn, psc, nil, patterns, o.messageBufferSize)
	sub.start(ctx)
	return sub, nil
}

// Close unsubscribes and releases all resources held by the Subscription.
// It is safe to call Close multiple times; subsequent calls are no-ops.
//
// Shutdown follows the canonical redigo pub/sub pattern:
//  1. Close the done channel so the outer reconnect loop can exit on the
//     next iteration.
//  2. Send UNSUBSCRIBE / PUNSUBSCRIBE on the active PubSubConn. Redigo
//     supports one concurrent reader and one concurrent writer on the
//     same conn, so this Send is safe even while readLoop is blocked in
//     Receive(). The server replies with Subscription{Count: 0}, which
//     readLoop detects as its clean exit signal.
//  3. Wait for the readLoop goroutine to exit before closing the pooled
//     connection — redigo's conn.Close drains pending replies via
//     Receive() and is not safe to call concurrently with another
//     reader.
//  4. Close the pooled connection to release the pool slot. Any cleanup
//     error is intentionally ignored: the pool slot is released either
//     way, and pub/sub conns can be left in a broken state after a
//     reconnect-backoff window or a server-side disconnect, producing
//     spurious "use of closed network connection" errors during the
//     pool's UNSUBSCRIBE cleanup.
func (s *Subscription) Close() error {
	s.closeOnce.Do(func() {
		close(s.done)
		// Signal the readLoop to exit cleanly. Errors are ignored: if the
		// conn is already broken, the readLoop's Receive() will return an
		// error and the goroutine will exit on its own.
		if len(s.channels) > 0 {
			_ = s.psc.Unsubscribe()
		}
		if len(s.patterns) > 0 {
			_ = s.psc.PUnsubscribe()
		}
	})
	// Wait for readLoop to finish before closing the underlying conn.
	s.wg.Wait()
	// Release the pool slot. Pool cleanup errors for pub/sub conns are not
	// actionable — the slot is released regardless.
	s.connOnce.Do(func() {
		_ = s.conn.Close()
	})
	return nil
}

// newSubscription creates a Subscription struct (does not start the goroutine).
func newSubscription(client *Client, conn redis.Conn, psc redis.PubSubConn, channels, patterns []string, bufSize int) *Subscription {
	msgCh := make(chan Message, bufSize)
	errCh := make(chan error, 16)
	return &Subscription{
		Messages: msgCh,
		Errors:   errCh,
		client:   client,
		conn:     conn,
		psc:      psc,
		channels: channels,
		patterns: patterns,
		msgCh:    msgCh,
		done:     make(chan struct{}),
		errCh:    errCh,
	}
}

// start launches the background goroutine that reads messages and handles reconnects.
// The goroutine exits when s.done is closed or when ctx is canceled.
func (s *Subscription) start(ctx context.Context) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
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

			// Check whether we should stop.
			select {
			case <-s.done:
				return
			default:
			}

			// Connection dropped — attempt reconnect with exponential backoff.
			select {
			case <-s.done:
				return
			case <-time.After(backoff):
			}

			if err := s.reconnect(ctx); err != nil {
				// Could not reconnect; report the error and advance backoff.
				select {
				case s.errCh <- err:
				default:
				}
				backoff = nextBackoff(backoff)
				continue
			}

			// Reconnected — reset backoff.
			backoff = pubSubReconnectMin
		}
	}()
}

// readLoop blocks on PubSubConn.Receive() and delivers incoming messages to
// the Subscription's Messages channel until one of the following:
//   - Close() unsubscribes and the server replies with a zero-count
//     Subscription message (clean shutdown signal).
//   - The connection errors out (the outer loop will attempt a reconnect
//     unless s.done is closed).
//   - A message send to s.msgCh loses the race with s.done (late shutdown
//     while delivery is in flight).
//
// Using blocking Receive() avoids redigo's ReceiveWithTimeout, which treats
// read-deadline expiry as fatal and closes the underlying net.Conn — the
// root cause of the "use of closed network connection" error during Close().
func (s *Subscription) readLoop() {
	for {
		switch msg := s.psc.Receive().(type) {
		case redis.Message:
			// Both regular (SUBSCRIBE) and pattern (PSUBSCRIBE) messages arrive as
			// redis.Message; Pattern is non-empty only for pattern-matched messages.
			out := Message{Channel: msg.Channel, Pattern: msg.Pattern, Data: msg.Data}
			select {
			case s.msgCh <- out:
			case <-s.done:
				return
			}
		case redis.Subscription:
			// Count == 0 means every channel/pattern has been unsubscribed —
			// this is the clean-exit signal sent by Close() via Unsubscribe /
			// PUnsubscribe. Any other Count is just a subscribe-state update
			// and is ignored.
			if msg.Count == 0 {
				return
			}
		case error:
			return
		}
	}
}

// isNetTimeout reports whether err is a network timeout error.
func isNetTimeout(err error) bool {
	if err == nil {
		return false
	}
	// redigo wraps read timeouts with this prefix.
	if err.Error() == "redigo: connection read timeout" {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
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

	// Swap in the new connection (old one is already dead).
	s.conn = conn
	s.psc = psc
	return nil
}

// toInterfaces converts a []string to []interface{} for redigo variadic calls.
func toInterfaces(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

// nextBackoff doubles the backoff duration, capped at pubSubReconnectMax.
func nextBackoff(d time.Duration) time.Duration {
	d *= 2
	if d > pubSubReconnectMax {
		return pubSubReconnectMax
	}
	return d
}
