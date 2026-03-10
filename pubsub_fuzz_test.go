package cache

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fuzzError is a string-backed error type used in fuzz tests to avoid
// inline errors.New() calls that would violate err113.
type fuzzError string

func (e fuzzError) Error() string { return string(e) }

// FuzzPublishRaw fuzzes PublishRaw with arbitrary channel and message strings.
func FuzzPublishRaw(f *testing.F) {
	f.Add("test-channel", "hello")
	f.Add("", "")
	f.Add("channel-with-🔔", "unicode-payload-🎉")
	f.Add("channel\nwith\nnewlines", "message\nwith\nnewlines")
	f.Add("channel\x00\x01\x02", "msg\x00\x01")
	f.Add(strings.Repeat("long-channel-", 50), strings.Repeat("long-message-", 50))

	f.Fuzz(func(t *testing.T, channel, message string) {
		client, conn := loadMockRedis(t)
		defer client.Close()

		conn.Command(PublishCommand, channel, message).Expect(int64(0))

		assert.NotPanics(t, func() {
			count, err := PublishRaw(conn, channel, message)
			if err == nil {
				assert.GreaterOrEqual(t, count, int64(0))
			}
		})
	})
}

// FuzzPublish fuzzes Publish with arbitrary channel and message strings.
func FuzzPublish(f *testing.F) {
	f.Add("test-channel", "hello")
	f.Add("", "")
	f.Add("unicode-🔔", "message-🎉")
	f.Add("chan\x00null", "msg\x00null")
	f.Add(strings.Repeat("ch-", 100), "v")

	f.Fuzz(func(t *testing.T, channel, message string) {
		client, conn := loadMockRedis(t)
		defer client.Close()

		conn.Command(PublishCommand, channel, message).Expect(int64(1))

		ctx := context.Background()
		assert.NotPanics(t, func() {
			count, err := Publish(ctx, client, channel, message)
			if err == nil {
				assert.GreaterOrEqual(t, count, int64(0))
			}
		})
	})
}

// FuzzNextBackoff fuzzes the nextBackoff helper with arbitrary duration inputs.
func FuzzNextBackoff(f *testing.F) {
	f.Add(int64(1_000_000_000)) // 1 second
	f.Add(int64(2_000_000_000))
	f.Add(int64(30_000_000_000)) // 30 seconds (max)
	f.Add(int64(0))
	f.Add(int64(-1))
	f.Add(int64(1 << 62)) // very large

	f.Fuzz(func(t *testing.T, nanos int64) {
		assert.NotPanics(t, func() {
			result := nextBackoff(time.Duration(nanos))
			assert.LessOrEqual(t, result, pubSubReconnectMax)
		})
	})
}

// FuzzToInterfaces fuzzes the toInterfaces helper with arbitrary string inputs.
func FuzzToInterfaces(f *testing.F) {
	f.Add("a", "b", "c")
	f.Add("", "", "")
	f.Add("unicode-🔑", "value-💎", "dep-🏠")
	f.Add("key\x00null", "val\x01ctrl", "dep\x02ctrl")

	f.Fuzz(func(t *testing.T, a, b, c string) {
		assert.NotPanics(t, func() {
			result := toInterfaces([]string{a, b, c})
			assert.Len(t, result, 3)
		})
	})
}

// FuzzIsNetTimeout fuzzes isNetTimeout to ensure no panics on arbitrary error messages.
func FuzzIsNetTimeout(f *testing.F) {
	f.Add("redigo: connection read timeout")
	f.Add("")
	f.Add("some other error")
	f.Add("redigo: connection write timeout")
	f.Add("connection refused")
	f.Add("i/o timeout")

	f.Fuzz(func(t *testing.T, errMsg string) {
		assert.NotPanics(t, func() {
			err := fuzzError(errMsg)
			_ = isNetTimeout(err)
		})
		// nil always returns false
		assert.False(t, isNetTimeout(nil))
	})
}
