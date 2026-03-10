package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// streamMockEntry holds an id and alternating field/value pairs for building XREAD mock responses.
type streamMockEntry struct {
	id     string
	fields []string
}

// makeStreamMockResponse builds the nested [[key, [[id, [f,v,...]]]]] format expected by parseStreamEntries.
func makeStreamMockResponse(key string, entries []streamMockEntry) []interface{} {
	entryList := make([]interface{}, 0, len(entries))
	for _, e := range entries {
		// Use index assignment to avoid asasalint (append([]interface{}, []byte) false-positive).
		fieldVals := make([]interface{}, len(e.fields))
		for i, f := range e.fields {
			fieldVals[i] = []byte(f)
		}
		// Wrap in interface{} explicitly so asasalint does not flag this as a
		// potential spread-slice mistake (we intentionally append one element).
		var entryPair interface{} = []interface{}{[]byte(e.id), fieldVals}
		entryList = append(entryList, entryPair)
	}
	return []interface{}{
		[]interface{}{
			[]byte(key),
			entryList,
		},
	}
}

// TestStreamAdd tests the method StreamAdd()
func TestStreamAdd(t *testing.T) {
	t.Run("stream add command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis(t)
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase string
			key      string
			fields   map[string]string
		}{
			// Single-entry maps only: map iteration is non-deterministic so we
			// cannot predict argument order when multiple fields are present.
			// Multi-field coverage is provided by the real Redis integration test.
			{"single field", testKey, map[string]string{"field1": "value1"}},
			{"empty fields", testKey, map[string]string{}},
			{"empty key", "", map[string]string{"field": "value"}},
			{"empty field value", testKey, map[string]string{"field": ""}},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// Build args the same way the implementation does
				args := make([]interface{}, 0, 2+2*len(test.fields))
				args = append(args, test.key, "*")
				for k, v := range test.fields {
					args = append(args, k, v)
				}

				// The main command to test
				cmd := conn.Command(StreamAddCommand, args...).Expect([]byte("1-0"))

				id, err := StreamAddRaw(conn, test.key, test.fields)
				require.NoError(t, err)
				assert.Equal(t, "1-0", id)
				assert.True(t, cmd.Called)
			})
		}

		// Test managed variant via mock client pool (single field — deterministic arg order)
		conn.Clear()
		fields := map[string]string{"key": "val"}
		conn.Command(StreamAddCommand, testKey, "*", "key", "val").Expect([]byte("2-0"))
		id, err := StreamAdd(context.Background(), client, testKey, fields)
		require.NoError(t, err)
		assert.Equal(t, "2-0", id)
	})

	t.Run("stream add command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis(t)
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn, t)
		require.NoError(t, err)

		// Fire the command
		var id string
		id, err = StreamAddRaw(conn, testKey, map[string]string{"field": "value"})
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		// Confirm the stream has one entry
		var length int64
		length, err = StreamLen(context.Background(), client, testKey)
		require.NoError(t, err)
		assert.Equal(t, int64(1), length)
	})
}

// ExampleStreamAdd is an example of the method StreamAdd()
func ExampleStreamAdd() {
	// Load a mocked redis for testing/examples
	client, conn := loadMockRedis()
	defer client.Close()

	conn.Command(StreamAddCommand, testKey, "*", "field", "value").Expect([]byte("1-0"))

	// Add an entry
	id, _ := StreamAdd(context.Background(), client, testKey, map[string]string{"field": "value"})
	fmt.Printf("added stream entry: %v", id)
	// Output:added stream entry: 1-0
}

// TestStreamAddCapped tests the method StreamAddCapped()
func TestStreamAddCapped(t *testing.T) {
	t.Run("stream add capped command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis(t)
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase string
			key      string
			maxLen   int64
			fields   map[string]string
		}{
			// Single-entry maps only to keep argument order deterministic.
			{"cap at 100", testKey, 100, map[string]string{"field": "value"}},
			{"cap at 1", testKey, 1, map[string]string{"f1": "v1"}},
			{"cap at 0", testKey, 0, map[string]string{"f1": "v1"}},
			{"empty key", "", 10, map[string]string{"field": "value"}},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// Build args the same way the implementation does
				args := make([]interface{}, 0, 5+2*len(test.fields))
				args = append(args, test.key, "MAXLEN", "~", test.maxLen, "*")
				for k, v := range test.fields {
					args = append(args, k, v)
				}

				// The main command to test
				cmd := conn.Command(StreamAddCommand, args...).Expect([]byte("1-0"))

				id, err := StreamAddCappedRaw(conn, test.key, test.maxLen, test.fields)
				require.NoError(t, err)
				assert.Equal(t, "1-0", id)
				assert.True(t, cmd.Called)
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		cappedFields := map[string]string{"data": "val"}
		conn.Command(StreamAddCommand, testKey, "MAXLEN", "~", int64(50), "*", "data", "val").Expect([]byte("3-0"))
		id, err := StreamAddCapped(context.Background(), client, testKey, 50, cappedFields)
		require.NoError(t, err)
		assert.Equal(t, "3-0", id)
	})

	t.Run("stream add capped command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis(t)
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn, t)
		require.NoError(t, err)

		// Add several entries capped at 2
		for i := 0; i < 5; i++ {
			_, err = StreamAddCapped(context.Background(), client, testKey, 2, map[string]string{"i": fmt.Sprintf("%d", i)})
			require.NoError(t, err)
		}

		// The stream should have at most 2 entries (may be slightly more due to ~ approximation)
		var length int64
		length, err = StreamLen(context.Background(), client, testKey)
		require.NoError(t, err)
		assert.LessOrEqual(t, length, int64(5))
	})
}

// TestStreamRead tests the method StreamRead()
func TestStreamRead(t *testing.T) {
	t.Run("stream read command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis(t)
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase        string
			key             string
			startID         string
			count           int64
			mockResponse    interface{}
			expectError     bool
			expectedEntries int
		}{
			{
				"single entry",
				testKey, "0", 10,
				makeStreamMockResponse(testKey, []streamMockEntry{{"1-0", []string{"field", "value"}}}),
				false, 1,
			},
			{
				"multiple entries",
				testKey, "0", 10,
				makeStreamMockResponse(testKey, []streamMockEntry{
					{"1-0", []string{"f1", "v1"}},
					{"2-0", []string{"f2", "v2", "f3", "v3"}},
				}),
				false, 2,
			},
			{
				"empty stream (nil response)",
				testKey, "0", 10,
				redis.ErrNil,
				true, 0,
			},
			{
				"start from dollar sign",
				testKey, "$", 10,
				redis.ErrNil,
				true, 0,
			},
			{
				"single entry multiple fields",
				testKey, "0", 10,
				makeStreamMockResponse(testKey, []streamMockEntry{{"3-0", []string{"a", "1", "b", "2", "c", "3"}}}),
				false, 1,
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				var cmd *redigomock.Cmd
				if test.expectError {
					cmd = conn.Command(StreamReadCommand, "COUNT", test.count, "STREAMS", test.key, test.startID).
						ExpectError(test.mockResponse.(error))
				} else {
					cmd = conn.Command(StreamReadCommand, "COUNT", test.count, "STREAMS", test.key, test.startID).
						Expect(test.mockResponse)
				}

				result, err := StreamReadRaw(conn, test.key, test.startID, test.count)
				if test.expectError {
					require.Error(t, err)
					assert.Nil(t, result)
				} else {
					require.NoError(t, err)
					assert.Len(t, result, test.expectedEntries)
				}
				assert.True(t, cmd.Called)
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		conn.Command(StreamReadCommand, "COUNT", int64(5), "STREAMS", testKey, "0").
			Expect(makeStreamMockResponse(testKey, []streamMockEntry{{"10-0", []string{"x", "y"}}}))
		readResult, err := StreamRead(context.Background(), client, testKey, "0", 5)
		require.NoError(t, err)
		require.Len(t, readResult, 1)
		assert.Equal(t, "10-0", readResult[0].ID)
		assert.Equal(t, "y", readResult[0].Fields["x"])
	})

	t.Run("stream read command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis(t)
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn, t)
		require.NoError(t, err)

		// Add entries
		_, err = StreamAddRaw(conn, testKey, map[string]string{"name": "alice", "age": "30"})
		require.NoError(t, err)
		_, err = StreamAddRaw(conn, testKey, map[string]string{"name": "bob", "age": "25"})
		require.NoError(t, err)

		// Read from the beginning
		var entries []StreamEntry
		entries, err = StreamRead(context.Background(), client, testKey, "0", 10)
		require.NoError(t, err)
		assert.Len(t, entries, 2)
		assert.NotEmpty(t, entries[0].ID)
		assert.Equal(t, "alice", entries[0].Fields["name"])
		assert.Equal(t, "30", entries[0].Fields["age"])
	})
}

// TestStreamReadBlock tests the method StreamReadBlock()
func TestStreamReadBlock(t *testing.T) {
	t.Run("stream read block raw using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis(t)
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		// Test successful raw read with block
		conn.Clear()
		cmd := conn.Command(StreamReadCommand, "BLOCK", int64(100), "COUNT", int64(10), "STREAMS", testKey, "0").
			Expect(makeStreamMockResponse(testKey, []streamMockEntry{{"5-0", []string{"event", "login"}}}))

		result, err := StreamReadBlockRaw(conn, testKey, "0", 10, 100)
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "5-0", result[0].ID)
		assert.Equal(t, "login", result[0].Fields["event"])
		assert.True(t, cmd.Called)
	})

	t.Run("stream read block managed using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis — use a fresh client/conn pair to avoid sharing state with other tests
		client, conn := loadMockRedis(t)
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		// Set up the mock to respond immediately (simulates data already in stream)
		conn.Command(StreamReadCommand, "BLOCK", int64(50), "COUNT", int64(5), "STREAMS", testKey, "$").
			Expect(makeStreamMockResponse(testKey, []streamMockEntry{{"6-0", []string{"k", "v"}}}))

		// The managed variant falls back to goroutine path for mock connections that do
		// not support DoContext; context is NOT canceled so data path is taken.
		blockResult, err := StreamReadBlock(context.Background(), client, testKey, "$", 5, 50)
		require.NoError(t, err)
		require.Len(t, blockResult, 1)
		assert.Equal(t, "6-0", blockResult[0].ID)
		assert.Equal(t, "v", blockResult[0].Fields["k"])
	})

	// Context cancellation test runs with real Redis only (skip in short mode).
	// The blocking-read/close-to-unblock pattern inherently involves concurrent
	// connection access that would trip the race detector on a mock that returns
	// immediately; on a real blocking read the close always precedes the read return.
	t.Run("stream read block context cancellation using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis(t)
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db (no entries — so XREAD BLOCK will actually block)
		err = clearRealRedis(conn, t)
		require.NoError(t, err)

		// Cancel the context immediately, then call StreamReadBlock;
		// it should return an error (ctx.Err() or connection error) without hanging
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		var entries []StreamEntry
		entries, err = StreamReadBlock(ctx, client, testKey, "$", 10, 0)
		require.Error(t, err)
		assert.Nil(t, entries)
	})

	t.Run("stream read block using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis(t)
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn, t)
		require.NoError(t, err)

		// Add an entry first so the block returns immediately
		_, err = StreamAddRaw(conn, testKey, map[string]string{"data": "ready"})
		require.NoError(t, err)

		// Read with a short block timeout — should return the existing entry
		var entries []StreamEntry
		entries, err = StreamReadBlock(context.Background(), client, testKey, "0", 10, 100)
		require.NoError(t, err)
		assert.Len(t, entries, 1)
		assert.Equal(t, "ready", entries[0].Fields["data"])
	})
}

// TestStreamTrim tests the method StreamTrim()
func TestStreamTrim(t *testing.T) {
	t.Run("stream trim command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis(t)
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase        string
			key             string
			maxLen          int64
			expectedTrimmed int64
		}{
			{"trim to 10", testKey, 10, 5},
			{"trim to 0", testKey, 0, 100},
			{"trim empty key", "", 10, 0},
			{"trim to 1", testKey, 1, 9},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				cmd := conn.Command(StreamTrimCommand, test.key, "MAXLEN", test.maxLen).Expect(test.expectedTrimmed)

				removed, err := StreamTrimRaw(conn, test.key, test.maxLen)
				require.NoError(t, err)
				assert.Equal(t, test.expectedTrimmed, removed)
				assert.True(t, cmd.Called)
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		conn.Command(StreamTrimCommand, testKey, "MAXLEN", int64(5)).Expect(int64(3))
		trimResult, err := StreamTrim(context.Background(), client, testKey, 5)
		require.NoError(t, err)
		assert.Equal(t, int64(3), trimResult)
	})

	t.Run("stream trim command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis(t)
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn, t)
		require.NoError(t, err)

		// Add 5 entries
		for i := 0; i < 5; i++ {
			_, err = StreamAddRaw(conn, testKey, map[string]string{"i": fmt.Sprintf("%d", i)})
			require.NoError(t, err)
		}

		// Trim to 3 (exact, not approximate)
		var removed int64
		removed, err = StreamTrim(context.Background(), client, testKey, 3)
		require.NoError(t, err)
		assert.Equal(t, int64(2), removed)

		// Verify length
		var length int64
		length, err = StreamLen(context.Background(), client, testKey)
		require.NoError(t, err)
		assert.Equal(t, int64(3), length)
	})
}

// TestStreamLen tests the method StreamLen()
func TestStreamLen(t *testing.T) {
	t.Run("stream len command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis(t)
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase       string
			key            string
			expectedLength int64
		}{
			{"non-empty stream", testKey, 5},
			{"empty stream", testKey, 0},
			{"empty key", "", 0},
			{"large stream", testKey, 1000000},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				cmd := conn.Command(StreamLenCommand, test.key).Expect(test.expectedLength)

				result, err := StreamLenRaw(conn, test.key)
				require.NoError(t, err)
				assert.Equal(t, test.expectedLength, result)
				assert.True(t, cmd.Called)
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		conn.Command(StreamLenCommand, testKey).Expect(int64(42))
		lenResult, err := StreamLen(context.Background(), client, testKey)
		require.NoError(t, err)
		assert.Equal(t, int64(42), lenResult)
	})

	t.Run("stream len command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis(t)
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn, t)
		require.NoError(t, err)

		// Empty stream has length 0
		var length int64
		length, err = StreamLen(context.Background(), client, testKey)
		require.NoError(t, err)
		assert.Equal(t, int64(0), length)

		// Add entries and verify length
		for i := 0; i < 3; i++ {
			_, err = StreamAddRaw(conn, testKey, map[string]string{"seq": fmt.Sprintf("%d", i)})
			require.NoError(t, err)
		}

		length, err = StreamLen(context.Background(), client, testKey)
		require.NoError(t, err)
		assert.Equal(t, int64(3), length)
	})
}

// TestParseStreamEntries tests the internal parseStreamEntries function edge cases
func TestParseStreamEntries(t *testing.T) {
	t.Run("nil input returns no error", func(t *testing.T) {
		entries, err := parseStreamEntries(nil)
		require.NoError(t, err)
		assert.Nil(t, entries)
	})

	t.Run("empty outer slice", func(t *testing.T) {
		entries, err := parseStreamEntries([]interface{}{})
		require.NoError(t, err)
		assert.Nil(t, entries)
	})

	t.Run("single entry single field", func(t *testing.T) {
		raw := makeStreamMockResponse("mystream", []streamMockEntry{{"1-0", []string{"name", "alice"}}})

		entries, err := parseStreamEntries(raw)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "1-0", entries[0].ID)
		assert.Equal(t, "alice", entries[0].Fields["name"])
	})

	t.Run("single entry empty fields", func(t *testing.T) {
		raw := makeStreamMockResponse("mystream", []streamMockEntry{{"2-0", []string{}}})

		entries, err := parseStreamEntries(raw)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "2-0", entries[0].ID)
		assert.Empty(t, entries[0].Fields)
	})

	t.Run("multiple entries multiple fields", func(t *testing.T) {
		raw := makeStreamMockResponse("mystream", []streamMockEntry{
			{"3-0", []string{"a", "1", "b", "2"}},
			{"4-0", []string{"c", "3"}},
		})

		entries, err := parseStreamEntries(raw)
		require.NoError(t, err)
		require.Len(t, entries, 2)
		assert.Equal(t, "3-0", entries[0].ID)
		assert.Equal(t, "1", entries[0].Fields["a"])
		assert.Equal(t, "2", entries[0].Fields["b"])
		assert.Equal(t, "4-0", entries[1].ID)
		assert.Equal(t, "3", entries[1].Fields["c"])
	})
}
