package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHashSet is testing the method HashSet()
func TestHashSet(t *testing.T) {
	t.Run("hash set command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase     string
			hashName     string
			key          string
			value        interface{}
			dependencies []string
		}{
			{"one dependency", testHashName, testKey, testStringValue, []string{testDependantKey}},
			{"new key with dep", "test-hash-name1", testKey, testStringValue, []string{testDependantKey}},
			{"third key", "test-hash-name2", testKey, testStringValue, []string{}},
			{"fourth key", "test-hash-name3", testKey, "", []string{}},
			{"fifth key", "test-hash-name4", "", "", []string{}},
			{"no name", "", "", "", []string{}},
			{"no name or value", "", "", []string{""}, []string{}},
			{"name is symbol", "-", "-", []string{""}, []string{}},
			{"value is a json interface", "-", "-", map[string]string{}, []string{}},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				var commands []*redigomock.Cmd

				// The main command to test
				commands = append(commands, conn.Command(HashKeySetCommand, test.hashName, test.key, test.value).Expect(test.value))

				// Loop for each dependency
				if len(test.dependencies) > 0 {
					commands = append(commands, conn.Command(MultiCommand))
					for _, dep := range test.dependencies {
						commands = append(commands, conn.Command(AddToSetCommand, DependencyPrefix+dep, test.hashName))
					}
					commands = append(commands, conn.Command(ExecuteCommand))

					err := HashSetRaw(conn, test.hashName, test.key, test.value, test.dependencies...)
					require.NoError(t, err)
				} else {
					err := HashSetRaw(conn, test.hashName, test.key, test.value, test.dependencies...)
					require.NoError(t, err)
				}

				for _, c := range commands {
					assert.Equal(t, true, c.Called)
				}
			})
		}
	})

	t.Run("hash set command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Fire the command
		err = HashSet(context.Background(), client, testHashName, testKey, testStringValue, testDependantKey)
		require.NoError(t, err)

		// Check that the command worked
		var testVal string
		testVal, err = HashGet(context.Background(), client, testHashName, testKey)
		require.NoError(t, err)
		assert.Equal(t, testStringValue, testVal)
	})
}

// ExampleHashSet is an example of the method HashSet()
func ExampleHashSet() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = HashSet(context.Background(), client, testHashName, testKey, testStringValue, testDependantKey)
	fmt.Printf("set: %s:%s value: %s dep key: %s", testHashName, testKey, testStringValue, testDependantKey)
	// Output:set: test-hash-name:test-key-name value: test-string-value dep key: test-dependant-key-name
}

// TestHashGet is testing the method HashGet() =
func TestHashGet(t *testing.T) {
	t.Run("hash get command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase string
			hashName string
			key      string
			value    interface{}
		}{
			{"one dependency", testHashName, testKey, testStringValue},
			{"new key with dep", "test-hash-name1", testKey, testStringValue},
			{"third key", "test-hash-name2", testKey, testStringValue},
			{"fourth key", "test-hash-name3", testKey, ""},
			{"fifth key", "test-hash-name4", "", ""},
			{"no name", "", "", ""},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				getCmd := conn.Command(HashGetCommand, test.hashName, test.key).Expect(test.value)

				val, err := HashGetRaw(conn, test.hashName, test.key)
				require.NoError(t, err)
				assert.Equal(t, true, getCmd.Called)
				assert.Equal(t, test.value, val)
			})
		}
	})

	t.Run("hash get command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Fire the command
		err = HashSet(context.Background(), client, testHashName, testKey, testStringValue, testDependantKey)
		require.NoError(t, err)

		// Check that the command worked
		var testVal string
		testVal, err = HashGet(context.Background(), client, testHashName, testKey)
		require.NoError(t, err)
		assert.Equal(t, testStringValue, testVal)
	})
}

// ExampleHashGet is an example of the method HashGet()
func ExampleHashGet() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = HashSet(context.Background(), client, testHashName, testKey, testStringValue, testDependantKey)

	// Get the value
	_, _ = HashGet(context.Background(), client, testHashName, testKey)
	fmt.Printf("got value: %s", testStringValue)
	// Output:got value: test-string-value
}

// TestHashMapSet is testing the method HashMapSet()
func TestHashMapSet(t *testing.T) {
	t.Run("hash map set command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase     string
			hashName     string
			key          string
			dependencies []string
			pairs        [][2]interface{}
		}{
			{
				"one dependency",
				testHashName,
				testKey,
				[]string{testDependantKey},
				[][2]interface{}{
					{"pair-1", "pair-1-value"},
					{"pair-2", "pair-2-value"},
					{"pair-3", "pair-3-value"},
				},
			},
			{
				"new key with dep",
				"test-hash-name1",
				testKey,
				[]string{},
				[][2]interface{}{
					{"pair-1", "pair-1-value"},
					{"pair-2", "pair-2-value"},
					{"pair-3", "pair-3-value"},
				},
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				args := make([]interface{}, 0, 2*len(test.pairs)+1)
				args = append(args, test.hashName)
				for _, pair := range test.pairs {
					args = append(args, pair[0])
					args = append(args, pair[1])
				}

				var commands []*redigomock.Cmd

				// The main command to test
				commands = append(commands, conn.Command(HashMapSetCommand, args...))

				// Loop for each dependency
				if len(test.dependencies) > 0 {
					commands = append(commands, conn.Command(MultiCommand))
					for _, dep := range test.dependencies {
						commands = append(commands, conn.Command(AddToSetCommand, DependencyPrefix+dep, test.hashName))
					}
					commands = append(commands, conn.Command(ExecuteCommand))

					err := HashMapSetRaw(conn, test.hashName, test.pairs, test.dependencies...)
					require.NoError(t, err)
				} else {
					err := HashMapSetRaw(conn, test.hashName, test.pairs, test.dependencies...)
					require.NoError(t, err)
				}

				for _, c := range commands {
					assert.Equal(t, true, c.Called)
				}
			})
		}
	})

	t.Run("hash map set command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Create pairs
		pairs := [][2]interface{}{
			{"pair-1", "pair-1-value"},
			{"pair-2", "pair-2-value"},
			{"pair-3", "pair-3-value"},
		}

		// Set the hash map
		err = HashMapSetRaw(conn, testHashName, pairs, testDependantKey)
		require.NoError(t, err)

		var val string
		val, err = HashGetRaw(conn, testHashName, "pair-1")
		require.NoError(t, err)
		assert.Equal(t, "pair-1-value", val)

		// Get a key in the map
		var values []string
		values, err = HashMapGetRaw(conn, testHashName, "pair-1", "pair-2")
		require.NoError(t, err)

		// Got two values?
		assert.Equal(t, 2, len(values))

		// Test value 1
		assert.Equal(t, "pair-1-value", values[0])

		// Test value 2
		assert.Equal(t, "pair-2-value", values[1])
	})
}

// ExampleHashMapSet is an example of the method HashMapSet()
func ExampleHashMapSet() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Create pairs
	pairs := [][2]interface{}{
		{"pair-1", "pair-1-value"},
		{"pair-2", "pair-2-value"},
		{"pair-3", "pair-3-value"},
	}

	// Set the hash map
	_ = HashMapSet(context.Background(), client, testHashName, pairs, testDependantKey)
	fmt.Printf("set: %s pairs: %d dep key: %s", testHashName, len(pairs), testDependantKey)
	// Output:set: test-hash-name pairs: 3 dep key: test-dependant-key-name
}

// TestHashMapSetExp is testing the method HashMapSetExp()
func TestHashMapSetExp(t *testing.T) {
	t.Run("hash map set exp command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase     string
			hashName     string
			key          string
			dependencies []string
			pairs        [][2]interface{}
			expiration   time.Duration
		}{
			{
				"one dependency",
				testHashName,
				testKey,
				[]string{testDependantKey},
				[][2]interface{}{
					{"pair-1", "pair-1-value"},
					{"pair-2", "pair-2-value"},
					{"pair-3", "pair-3-value"},
				},
				2 * time.Second,
			},
			{
				"new key with dep",
				"test-hash-name1",
				testKey,
				[]string{},
				[][2]interface{}{
					{"pair-1", "pair-1-value"},
					{"pair-2", "pair-2-value"},
					{"pair-3", "pair-3-value"},
				},
				2 * time.Second,
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				args := make([]interface{}, 0, 2*len(test.pairs)+1)
				args = append(args, test.hashName)
				for _, pair := range test.pairs {
					args = append(args, pair[0], pair[1])
				}

				var commands []*redigomock.Cmd

				// The main command to test
				commands = append(commands, conn.Command(HashMapSetCommand, args...))
				commands = append(commands, conn.Command(ExpireCommand, test.hashName, int64(test.expiration.Seconds())))

				// Loop for each dependency
				if len(test.dependencies) > 0 {
					commands = append(commands, conn.Command(MultiCommand))
					for _, dep := range test.dependencies {
						commands = append(commands, conn.Command(AddToSetCommand, DependencyPrefix+dep, test.hashName))
					}
					commands = append(commands, conn.Command(ExecuteCommand))

					err := HashMapSetExp(context.Background(), client, test.hashName, test.pairs, test.expiration, test.dependencies...)
					require.NoError(t, err)
				} else {
					err := HashMapSetExp(context.Background(), client, test.hashName, test.pairs, test.expiration, test.dependencies...)
					require.NoError(t, err)
				}

				for _, c := range commands {
					assert.Equal(t, true, c.Called)
				}
			})
		}
	})

	t.Run("hash map set exp command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Create pairs
		pairs := [][2]interface{}{
			{"pair-1", "pair-1-value"},
			{"pair-2", "pair-2-value"},
			{"pair-3", "pair-3-value"},
		}

		// Set the hash map
		err = HashMapSetExpRaw(conn, testHashName, pairs, 2*time.Second, testDependantKey)
		require.NoError(t, err)

		var val string
		val, err = HashGetRaw(conn, testHashName, "pair-1")
		require.NoError(t, err)
		assert.Equal(t, "pair-1-value", val)

		// Get a key in the map
		var values []string
		values, err = HashMapGetRaw(conn, testHashName, "pair-1", "pair-2")
		require.NoError(t, err)

		// Got two values?
		assert.Equal(t, 2, len(values))

		// Test value 1
		assert.Equal(t, "pair-1-value", values[0])

		// Test value 2
		assert.Equal(t, "pair-2-value", values[1])

		// Wait a few seconds and test
		t.Log("sleeping for 3 seconds...")
		time.Sleep(time.Second * 3)

		values, err = HashMapGetRaw(conn, testHashName, "pair-1")
		require.NoError(t, err)
		assert.Equal(t, []string{""}, values)
	})
}

// ExampleHashMapSetExp is an example of the method HashMapSetExp()
func ExampleHashMapSetExp() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Create pairs
	pairs := [][2]interface{}{
		{"pair-1", "pair-1-value"},
		{"pair-2", "pair-2-value"},
		{"pair-3", "pair-3-value"},
	}

	// Set the hash map
	_ = HashMapSetExp(context.Background(), client, testHashName, pairs, 5*time.Second, testDependantKey)
	fmt.Printf("set: %s pairs: %d dep key: %s exp: %v", testHashName, len(pairs), testDependantKey, 5*time.Second)
	// Output:set: test-hash-name pairs: 3 dep key: test-dependant-key-name exp: 5s
}
