package cache

import (
	"context"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPublishRaw tests the PublishRaw() function using a mock connection
func TestPublishRaw(t *testing.T) {
	t.Run("publish command using mocked redis", func(t *testing.T) {
		t.Parallel()

		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase string
			channel  string
			message  interface{}
			expected int64
		}{
			{"basic publish string", "test-channel", "hello", 1},
			{"publish to empty channel", "", "hello", 0},
			{"publish empty message", "test-channel", "", 0},
			{"publish with zero subscribers", "no-subs-channel", "msg", 0},
			{"publish with multiple subscribers", "busy-channel", "broadcast", 5},
			{"publish byte slice message", "test-channel", []byte("raw bytes"), 1},
			{"publish integer message", "test-channel", 42, 1},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				cmd := conn.Command(PublishCommand, test.channel, test.message).Expect(test.expected)

				count, err := PublishRaw(conn, test.channel, test.message)
				require.NoError(t, err)
				assert.Equal(t, test.expected, count)
				assert.True(t, cmd.Called)
			})
		}
	})

	t.Run("publish returns error on failed connection", func(t *testing.T) {
		t.Parallel()

		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		conn.Clear()
		conn.Command(PublishCommand, "test-channel", "hello").ExpectError(redis.ErrNil)

		count, err := PublishRaw(conn, "test-channel", "hello")
		require.Error(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// TestPublish tests the Publish() managed function using a mock connection
func TestPublish(t *testing.T) {
	t.Run("publish using mock client pool", func(t *testing.T) {
		t.Parallel()

		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		conn.Command(PublishCommand, "test-channel", "hello").Expect(int64(2))

		count, err := Publish(context.Background(), client, "test-channel", "hello")
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("publish using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Publishing to a channel with no subscribers returns 0 — that's valid
		count, err := Publish(context.Background(), client, "test-pubsub-channel", "hello")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))
	})
}

// TestSubscribe tests the Subscribe() function using real Redis
func TestSubscribe(t *testing.T) {
	t.Run("subscribe with no channels returns error", func(t *testing.T) {
		t.Parallel()

		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		sub, err := Subscribe(context.Background(), client)
		require.Error(t, err)
		assert.Nil(t, sub)
	})

	t.Run("subscribe and receive message using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		ctx := context.Background()
		subClient, subConn, err := loadRealRedis()
		require.NoError(t, err)
		defer subClient.CloseAll(subConn)

		pubClient, pubConn, err := loadRealRedis()
		require.NoError(t, err)
		defer pubClient.CloseAll(pubConn)

		channel := "test-subscribe-recv"
		payload := "hello-from-publisher"

		// Subscribe using the subscriber client
		sub, err := Subscribe(ctx, subClient, channel)
		require.NoError(t, err)
		require.NotNil(t, sub)
		defer func() { _ = sub.Close() }()

		// Give the subscription goroutine a moment to start receiving
		time.Sleep(50 * time.Millisecond)

		// Publish from the separate publisher connection
		_, err = PublishRaw(pubConn, channel, payload)
		require.NoError(t, err)

		// Wait for the message with a 5s timeout guard
		select {
		case msg, ok := <-sub.Messages:
			require.True(t, ok, "Messages channel closed prematurely")
			assert.Equal(t, channel, msg.Channel)
			assert.Equal(t, payload, string(msg.Data))
			assert.Empty(t, msg.Pattern, "Pattern should be empty for regular subscribe")
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for pub/sub message")
		}
	})

	t.Run("subscribe to multiple channels using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		ctx := context.Background()
		subClient, subConn, err := loadRealRedis()
		require.NoError(t, err)
		defer subClient.CloseAll(subConn)

		pubClient, pubConn, err := loadRealRedis()
		require.NoError(t, err)
		defer pubClient.CloseAll(pubConn)

		ch1 := "test-multi-ch1"
		ch2 := "test-multi-ch2"

		sub, err := Subscribe(ctx, subClient, ch1, ch2)
		require.NoError(t, err)
		require.NotNil(t, sub)
		defer func() { _ = sub.Close() }()

		time.Sleep(50 * time.Millisecond)

		_, err = PublishRaw(pubConn, ch1, "msg-ch1")
		require.NoError(t, err)
		_, err = PublishRaw(pubConn, ch2, "msg-ch2")
		require.NoError(t, err)

		received := make(map[string]string)
		deadline := time.After(5 * time.Second)
		for len(received) < 2 {
			select {
			case msg, ok := <-sub.Messages:
				require.True(t, ok)
				received[msg.Channel] = string(msg.Data)
			case <-deadline:
				t.Fatalf("timed out waiting for messages; got %d/2", len(received))
			}
		}
		assert.Equal(t, "msg-ch1", received[ch1])
		assert.Equal(t, "msg-ch2", received[ch2])
	})
}

// TestPSubscribe tests the PSubscribe() function using real Redis
func TestPSubscribe(t *testing.T) {
	t.Run("psubscribe with no patterns returns error", func(t *testing.T) {
		t.Parallel()

		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		sub, err := PSubscribe(context.Background(), client)
		require.Error(t, err)
		assert.Nil(t, sub)
	})

	t.Run("psubscribe pattern matching using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		ctx := context.Background()
		subClient, subConn, err := loadRealRedis()
		require.NoError(t, err)
		defer subClient.CloseAll(subConn)

		pubClient, pubConn, err := loadRealRedis()
		require.NoError(t, err)
		defer pubClient.CloseAll(pubConn)

		pattern := "test-psubscribe-*"
		targetChannel := "test-psubscribe-events"
		payload := "pattern-matched-message"

		sub, err := PSubscribe(ctx, subClient, pattern)
		require.NoError(t, err)
		require.NotNil(t, sub)
		defer func() { _ = sub.Close() }()

		time.Sleep(50 * time.Millisecond)

		_, err = PublishRaw(pubConn, targetChannel, payload)
		require.NoError(t, err)

		select {
		case msg, ok := <-sub.Messages:
			require.True(t, ok, "Messages channel closed prematurely")
			assert.Equal(t, targetChannel, msg.Channel)
			assert.Equal(t, pattern, msg.Pattern, "Pattern field should be set for psubscribe messages")
			assert.Equal(t, payload, string(msg.Data))
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for pattern-subscribed message")
		}
	})

	t.Run("psubscribe wildcard matches multiple channels using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		ctx := context.Background()
		subClient, subConn, err := loadRealRedis()
		require.NoError(t, err)
		defer subClient.CloseAll(subConn)

		pubClient, pubConn, err := loadRealRedis()
		require.NoError(t, err)
		defer pubClient.CloseAll(pubConn)

		pattern := "test-wild-*"

		sub, err := PSubscribe(ctx, subClient, pattern)
		require.NoError(t, err)
		require.NotNil(t, sub)
		defer func() { _ = sub.Close() }()

		time.Sleep(50 * time.Millisecond)

		channels := []string{"test-wild-alpha", "test-wild-beta", "test-wild-gamma"}
		for _, ch := range channels {
			_, err = PublishRaw(pubConn, ch, "data-"+ch)
			require.NoError(t, err)
		}

		received := make(map[string]string)
		deadline := time.After(7 * time.Second)
		for len(received) < len(channels) {
			select {
			case msg, ok := <-sub.Messages:
				require.True(t, ok)
				assert.Equal(t, pattern, msg.Pattern)
				received[msg.Channel] = string(msg.Data)
			case <-deadline:
				t.Fatalf("timed out; got %d/%d messages", len(received), len(channels))
			}
		}
		for _, ch := range channels {
			assert.Equal(t, "data-"+ch, received[ch])
		}
	})
}

// TestSubscriptionClose tests the Close() method idempotency
func TestSubscriptionClose(t *testing.T) {
	t.Run("close is idempotent using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		ctx := context.Background()
		client, conn, err := loadRealRedis()
		require.NoError(t, err)
		defer client.CloseAll(conn)

		sub, err := Subscribe(ctx, client, "test-close-idempotent")
		require.NoError(t, err)
		require.NotNil(t, sub)

		// First close — should succeed without error
		err = sub.Close()
		require.NoError(t, err)

		// Second close — must not panic or deadlock
		// Use a goroutine with a timeout to detect deadlocks
		done := make(chan struct{})
		go func() {
			defer close(done)
			_ = sub.Close() // error return value is intentionally ignored for idempotency
		}()

		select {
		case <-done:
			// Passed — no deadlock
		case <-time.After(3 * time.Second):
			t.Fatal("deadlock detected on second Close() call")
		}
	})

	t.Run("messages channel is closed after Close()", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		ctx := context.Background()
		client, conn, err := loadRealRedis()
		require.NoError(t, err)
		defer client.CloseAll(conn)

		sub, err := Subscribe(ctx, client, "test-close-drains")
		require.NoError(t, err)
		require.NotNil(t, sub)

		_ = sub.Close()

		// After Close, Messages must be closed (drain then EOF)
		select {
		case _, ok := <-sub.Messages:
			// May receive buffered messages; eventually ok==false
			if !ok {
				return // closed — good
			}
			// drain remaining
			for msg := range sub.Messages {
				_ = msg
			}
		case <-time.After(3 * time.Second):
			t.Fatal("Messages channel not closed after Close()")
		}
	})
}

// TestSubscriptionContextCancellation tests that canceling the context stops the subscription
func TestSubscriptionContextCancellation(t *testing.T) {
	t.Run("context cancellation stops subscription using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		ctx, cancel := context.WithCancel(context.Background())

		client, conn, err := loadRealRedis()
		require.NoError(t, err)
		defer client.CloseAll(conn)

		sub, err := Subscribe(ctx, client, "test-ctx-cancel")
		require.NoError(t, err)
		require.NotNil(t, sub)

		// Cancel the context — this should trigger Close() internally
		cancel()

		// Messages channel must be closed within a reasonable time
		timeout := time.After(5 * time.Second)
		for {
			select {
			case _, ok := <-sub.Messages:
				if !ok {
					return // closed — good
				}
				// drain any buffered messages
			case <-timeout:
				t.Fatal("Messages channel not closed after context cancellation")
			}
		}
	})

	t.Run("context already canceled before subscribe using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		client, conn, err := loadRealRedis()
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Subscribe with already-canceled context — expect a connection error
		_, err = Subscribe(ctx, client, "test-pre-canceled")
		require.Error(t, err, "expected error when context is already canceled")
	})
}

// TestNextBackoff tests the internal backoff helper
func TestNextBackoff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    time.Duration
		expected time.Duration
	}{
		{1 * time.Second, 2 * time.Second},
		{2 * time.Second, 4 * time.Second},
		{15 * time.Second, 30 * time.Second},
		{16 * time.Second, 30 * time.Second}, // capped
		{30 * time.Second, 30 * time.Second}, // already at max
	}
	for _, tt := range tests {
		result := nextBackoff(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

// TestToInterfaces tests the internal helper that converts []string to []interface{}
func TestToInterfaces(t *testing.T) {
	t.Parallel()

	t.Run("converts non-empty slice", func(t *testing.T) {
		result := toInterfaces([]string{"a", "b", "c"})
		require.Len(t, result, 3)
		assert.Equal(t, "a", result[0])
		assert.Equal(t, "b", result[1])
		assert.Equal(t, "c", result[2])
	})

	t.Run("converts empty slice", func(t *testing.T) {
		result := toInterfaces([]string{})
		assert.Empty(t, result)
	})

	t.Run("converts single element", func(t *testing.T) {
		result := toInterfaces([]string{"only"})
		require.Len(t, result, 1)
		assert.Equal(t, "only", result[0])
	})
}
