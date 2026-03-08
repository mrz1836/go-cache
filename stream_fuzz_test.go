package cache

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// FuzzStreamAdd fuzzes the StreamAdd function with key, field name, and field value inputs
func FuzzStreamAdd(f *testing.F) {
	// Seed corpus with interesting values
	f.Add("test-stream", "field1", "value1")
	f.Add("", "", "")
	f.Add("unicode-stream-🌊", "unicode-field-🔑", "unicode-value-✨")
	f.Add("stream with spaces", "field with spaces", "value with spaces")
	f.Add("stream\nwith\nnewlines", "field\nname", "value\ndata")
	f.Add("stream\x00\x01\x02", "field\x00\x01", "val\x00\x01")
	f.Add(strings.Repeat("long-stream-", 100), strings.Repeat("long-field-", 50), strings.Repeat("long-value-", 50))
	f.Add("s", "f", "v")
	f.Add("mystream", "$", "special-id-value")
	f.Add("mystream", "0", "zero-start-id")

	f.Fuzz(func(t *testing.T, key, fieldName, fieldValue string) {
		client, conn := loadMockRedis()
		defer client.Close()

		fields := map[string]string{fieldName: fieldValue}

		// Build expected args the same way StreamAddRaw does
		conn.Command(StreamAddCommand, key, "*", fieldName, fieldValue).Expect([]byte("1-0"))

		ctx := context.Background()

		assert.NotPanics(t, func() {
			id, err := StreamAdd(ctx, client, key, fields)
			if err == nil {
				assert.IsType(t, "", id)
			}
		})
	})
}

// FuzzStreamRead fuzzes the StreamRead function with key, startID, and count inputs
func FuzzStreamRead(f *testing.F) {
	// Seed corpus with interesting values
	f.Add("test-stream", "0", int64(10))
	f.Add("", "0", int64(0))
	f.Add("unicode-stream-🌊", "$", int64(1))
	f.Add("stream with spaces", "0-0", int64(100))
	f.Add("stream\nwith\nnewlines", "0", int64(5))
	f.Add("stream\x00\x01\x02", "0", int64(1))
	f.Add(strings.Repeat("long-stream-", 50), "0", int64(999))
	f.Add("mystream", "$", int64(1))
	f.Add("mystream", "0", int64(1))
	f.Add("mystream", "1234567890-0", int64(10))
	f.Add("mystream", "*", int64(10))

	f.Fuzz(func(t *testing.T, key, startID string, count int64) {
		client, conn := loadMockRedis()
		defer client.Close()

		// Return empty result from the mock (nil → ErrNil from Redis, or an empty response)
		conn.Command(StreamReadCommand, "COUNT", count, "STREAMS", key, startID).Expect([]interface{}{})

		ctx := context.Background()

		assert.NotPanics(t, func() {
			result, err := StreamRead(ctx, client, key, startID, count)
			if err == nil {
				assert.IsType(t, []StreamEntry{}, result)
			}
		})
	})
}
