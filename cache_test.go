package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
)

// Testing variables
const (
	testDependantKey         = "test-dependant-key-name"
	testHashName             = "test-hash-name"
	testIdleTimeout          = 240 * time.Second
	testKey                  = "test-key-name"
	testKillDependencyHash   = "a648f768f57e73e2497ccaa113d5ad9e731c5cd8"
	testLocalConnectionURL   = "redis://localhost:6379"
	testMaxActiveConnections = 0
	testMaxConnLifetime      = 60 * time.Second
	testMaxIdleConnections   = 10
	testStringValue          = "test-string-value"
)

// loadMockRedis will load a mocked redis connection
func loadMockRedis() (client *Client, conn *redigomock.Conn) {
	conn = redigomock.NewConn()
	client = &Client{
		DependencyScriptSha: "",
		Pool: &redis.Pool{
			Dial:            func() (redis.Conn, error) { return conn, nil },
			IdleTimeout:     testIdleTimeout,
			MaxActive:       testMaxActiveConnections,
			MaxConnLifetime: testMaxConnLifetime,
			MaxIdle:         testMaxIdleConnections,
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}
				_, doErr := c.Do(PingCommand)
				return doErr
			},
		},
		ScriptsLoaded: nil,
	}

	return
}

// loadRealRedis will load a real redis connection
func loadRealRedis() (client *Client, conn redis.Conn, err error) {
	client, err = Connect(
		context.Background(),
		testLocalConnectionURL,
		testMaxActiveConnections,
		testMaxIdleConnections,
		testMaxConnLifetime,
		testIdleTimeout,
		true,
		false,
	)
	if err != nil {
		return
	}

	conn, err = client.GetConnectionWithContext(context.Background())
	return
}

// loadRealRedisWithNewRelic will load a real redis connection (new relic enabled)
func loadRealRedisWithNewRelic() (client *Client, conn redis.Conn, err error) {
	client, err = Connect(
		context.Background(),
		testLocalConnectionURL,
		testMaxActiveConnections,
		testMaxIdleConnections,
		testMaxConnLifetime,
		testIdleTimeout,
		true,
		true,
	)
	if err != nil {
		return
	}

	conn, err = client.GetConnectionWithContext(context.Background())
	return
}

// clearRealRedis will clear a real redis db
func clearRealRedis(conn redis.Conn) error {
	return DestroyCacheRaw(conn)
}

// TestSet is testing the method Set()
func TestSet(t *testing.T) {

	t.Run("set command using mocked redis", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase     string
			key          string
			value        string
			dependencies []string
		}{
			{"key with dependencies", testKey, testStringValue, []string{testDependantKey}},
			{"key with no dependencies", testKey, testStringValue, []string{}},
			{"key with empty value", testKey, "", []string{}},
			{"key with spaces", "key name", "some val", []string{}},
			{"key with symbols", ".key name;!()\\", "", []string{}},
			{"key with symbols and value as symbols", ".key name;!()\\", `\ / ; [ ] { }!`, []string{}},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				var commands []*redigomock.Cmd

				// The main command to test
				commands = append(commands, conn.Command(SetCommand, test.key, test.value).Expect(test.value))

				// Loop for each dependency
				if len(test.dependencies) > 0 {
					commands = append(commands, conn.Command(MultiCommand))
					for _, dep := range test.dependencies {
						commands = append(commands, conn.Command(AddToSetCommand, DependencyPrefix+dep, test.key))
					}
					commands = append(commands, conn.Command(ExecuteCommand))

					err := Set(ctx, client, test.key, test.value, test.dependencies...)
					assert.NoError(t, err)
				} else {
					err := Set(ctx, client, test.key, test.value, test.dependencies...)
					assert.NoError(t, err)
				}

				for _, c := range commands {
					assert.Equal(t, true, c.Called)
				}
			})
		}
	})

	t.Run("set command using real redis", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase     string
			key          string
			value        string
			dependencies []string
		}{
			{"key with dependencies", testKey, testStringValue, []string{testDependantKey}},
			{"key with no dependencies", testKey, testStringValue, []string{""}},
			{"key with empty value", testKey, "", []string{""}},
			{"key with spaces", "key name", "some val", []string{""}},
			{"key with symbols", ".key name;!()\\", "", []string{""}},
			{"key with symbols and value as symbols", ".key name;!()\\", `\ / ; [ ] { }!`, []string{""}},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {

				// Start with a fresh db
				err = clearRealRedis(conn)
				assert.NoError(t, err)

				// Run command
				err = Set(context.Background(), client, test.key, test.value, test.dependencies...)
				assert.NoError(t, err)

				// Validate via getting the data from redis
				var testVal string
				testVal, err = Get(context.Background(), client, test.key)
				assert.NoError(t, err)
				assert.Equal(t, test.value, testVal)
			})
		}
	})

	t.Run("set command using real redis (new relic)", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedisWithNewRelic()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase     string
			key          string
			value        string
			dependencies []string
		}{
			{"key with dependencies", testKey, testStringValue, []string{testDependantKey}},
			{"key with no dependencies", testKey, testStringValue, []string{""}},
			{"key with empty value", testKey, "", []string{""}},
			{"key with spaces", "key name", "some val", []string{""}},
			{"key with symbols", ".key name;!()\\", "", []string{""}},
			{"key with symbols and value as symbols", ".key name;!()\\", `\ / ; [ ] { }!`, []string{""}},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {

				// Start with a fresh db
				err = clearRealRedis(conn)
				assert.NoError(t, err)

				// Run command
				err = Set(context.Background(), client, test.key, test.value, test.dependencies...)
				assert.NoError(t, err)

				// Validate via getting the data from redis
				var testVal string
				testVal, err = Get(context.Background(), client, test.key)
				assert.NoError(t, err)
				assert.Equal(t, test.value, testVal)
			})
		}
	})

	t.Run("set cmd, trigger context err", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		client.CloseAll(conn)

		err := Set(context.TODO(), client, "key", "value")
		assert.Error(t, err)
	})
}

// ExampleSet is an example of the method Set()
func ExampleSet() {

	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = Set(context.Background(), client, testKey, testStringValue, testDependantKey)
	fmt.Printf("set: %s value: %s dep key: %s", testKey, testStringValue, testDependantKey)
	// Output:set: test-key-name value: test-string-value dep key: test-dependant-key-name
}

// TestSetExp is testing the method SetExp()
func TestSetExp(t *testing.T) {

	t.Run("set exp command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase     string
			key          string
			value        string
			expiration   time.Duration
			dependencies []string
		}{
			{"key with dependencies", "test-set-exp", testStringValue, 2 * time.Second, []string{testDependantKey}},
			{"key with no dependencies", "test-set2", testStringValue, 2 * time.Second, []string{}},
			{"key with empty value", "test-set3", "", 2 * time.Second, []string{}},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				var commands []*redigomock.Cmd

				// The main command to test
				commands = append(commands, conn.Command(SetExpirationCommand, test.key, int64(test.expiration.Seconds()), test.value).Expect(test.value))

				// Loop for each dependency
				if len(test.dependencies) > 0 {
					commands = append(commands, conn.Command(MultiCommand))
					for _, dep := range test.dependencies {
						commands = append(commands, conn.Command(AddToSetCommand, DependencyPrefix+dep, test.key))
					}
					commands = append(commands, conn.Command(ExecuteCommand))

					err := SetExp(context.Background(), client, test.key, test.value, test.expiration, test.dependencies...)
					assert.NoError(t, err)
				} else {
					err := SetExp(context.Background(), client, test.key, test.value, test.expiration, test.dependencies...)
					assert.NoError(t, err)
				}

				for _, c := range commands {
					assert.Equal(t, true, c.Called)
				}
			})
		}
	})

	t.Run("set exp command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = SetExp(context.Background(), client, testKey, testStringValue, 2*time.Second, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var testVal string
		testVal, err = Get(context.Background(), client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testStringValue, testVal)

		// Wait a few seconds and test
		t.Log("sleeping for 3 seconds...")
		time.Sleep(time.Second * 3)

		// Check that the key is expired
		testVal, err = Get(context.Background(), client, testKey)
		assert.Error(t, err)
		assert.Equal(t, "", testVal)
		assert.Equal(t, redis.ErrNil, err)
	})

	t.Run("set exp cmd, trigger context err", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		client.CloseAll(conn)

		err := SetExp(context.TODO(), client, "key", "value", 10*time.Second)
		assert.Error(t, err)
	})
}

// ExampleSetExp is an example of the method SetExp()
func ExampleSetExp() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = SetExp(context.Background(), client, testKey, testStringValue, 2*time.Minute, testDependantKey)
	fmt.Printf("set: %s value: %s exp: %v dep key: %s", testKey, testStringValue, 2*time.Minute, testDependantKey)
	// Output:set: test-key-name value: test-string-value exp: 2m0s dep key: test-dependant-key-name
}

// TestGet is testing the method Get()
func TestGet(t *testing.T) {

	t.Run("get command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase string
			key      string
			value    interface{}
		}{
			{"valid value", testHashName, testStringValue},
			{"new key", "test-hash-name1", testStringValue},
			{"third key", "test-hash-name2", testStringValue},
			{"fourth key", "test-hash-name3", ""},
			{"no name", "", ""},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				getCmd := conn.Command(GetCommand, test.key).Expect(test.value)

				val, err := Get(context.Background(), client, test.key)
				assert.NoError(t, err)
				assert.Equal(t, true, getCmd.Called)
				assert.Equal(t, test.value, val)
			})
		}
	})

	t.Run("get command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = Set(context.Background(), client, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var testVal string
		testVal, err = Get(context.Background(), client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testStringValue, testVal)
	})

	t.Run("get command using real redis (new relic)", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedisWithNewRelic()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = Set(context.Background(), client, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var testVal string
		testVal, err = Get(context.Background(), client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testStringValue, testVal)
	})

	t.Run("get cmd, trigger context err", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		client.CloseAll(conn)

		val, err := Get(context.TODO(), client, "123456")
		assert.Error(t, err)
		assert.Equal(t, "", val)
	})
}

// ExampleGet is an example of the method Get()
func ExampleGet() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = Set(context.Background(), client, testKey, testStringValue, testDependantKey)

	// Get the value
	_, _ = Get(context.Background(), client, testKey)
	fmt.Printf("got value: %s", testStringValue)
	// Output:got value: test-string-value
}

// TestGetBytes is testing the method GetBytes()
func TestGetBytes(t *testing.T) {

	t.Run("get bytes command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase string
			key      string
			value    string
		}{
			{"valid value", testHashName, testStringValue},
			{"new key", "test-hash-name1", testStringValue},
			{"third key", "test-hash-name2", testStringValue},
			{"fourth key", "test-hash-name3", ""},
			{"no name", "", ""},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				getCmd := conn.Command(GetCommand, test.key).Expect([]byte(test.value))

				val, err := GetBytes(context.Background(), client, test.key)
				assert.NoError(t, err)
				assert.Equal(t, true, getCmd.Called)
				assert.Equal(t, []byte(test.value), val)
			})
		}
	})

	t.Run("get bytes command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = Set(context.Background(), client, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var testVal []byte
		testVal, err = GetBytes(context.Background(), client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, []byte(testStringValue), testVal)
	})

	t.Run("get bytes command using real redis (new relic)", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedisWithNewRelic()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = Set(context.Background(), client, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var testVal []byte
		testVal, err = GetBytes(context.Background(), client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, []byte(testStringValue), testVal)
	})

	t.Run("get bytes cmd, trigger context err", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		client.CloseAll(conn)

		val, err := GetBytes(context.TODO(), client, "123456")
		assert.Error(t, err)
		assert.Equal(t, []byte(nil), val)
	})
}

// ExampleGetBytes is an example of the method GetBytes()
func ExampleGetBytes() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = Set(context.Background(), client, testKey, testStringValue, testDependantKey)

	// Get the value
	_, _ = GetBytes(context.Background(), client, testKey)
	fmt.Printf("got value: %s", testStringValue)
	// Output:got value: test-string-value
}

// TestGetAllKeys is testing the method GetAllKeys()
func TestGetAllKeys(t *testing.T) {

	t.Run("get all keys command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		conn.Clear()

		// The main command to test
		getCmd := conn.Command(KeysCommand, AllKeysCommand).Expect([]interface{}{[]byte(testKey)})

		val, err := GetAllKeys(context.Background(), client)
		assert.NoError(t, err)
		assert.Equal(t, true, getCmd.Called)
		assert.Equal(t, []string{testKey}, val)
	})

	t.Run("get all keys command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = Set(context.Background(), client, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var keys []string
		keys, err = GetAllKeys(context.Background(), client)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(keys))
	})

	t.Run("get all keys command using real redis (new relic)", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedisWithNewRelic()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = Set(context.Background(), client, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var keys []string
		keys, err = GetAllKeys(context.Background(), client)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(keys))
	})

	t.Run("get all keys cmd, trigger context err", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		client.CloseAll(conn)

		val, err := GetAllKeys(context.TODO(), client)
		assert.Error(t, err)
		assert.Equal(t, []string(nil), val)
	})
}

// ExampleGetAllKeys is an example of the method GetAllKeys()
func ExampleGetAllKeys() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = Set(context.Background(), client, testKey, testStringValue, testDependantKey)

	// Get the keys
	_, _ = GetAllKeys(context.Background(), client)
	fmt.Printf("found keys: %d", len([]string{testKey, testDependantKey}))
	// Output:found keys: 2
}

// TestExists is testing the method Exists()
func TestExists(t *testing.T) {

	t.Run("exists command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		conn.Clear()

		// todo: add table tests

		// The main command to test
		existsCmd := conn.Command(ExistsCommand, testKey).Expect(interface{}(int64(1)))

		val, err := Exists(context.Background(), client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, true, existsCmd.Called)
		assert.Equal(t, true, val)
	})

	t.Run("exists command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = Set(context.Background(), client, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var found bool
		found, err = Exists(context.Background(), client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, true, found)
	})

	t.Run("exists cmd, trigger context err", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		client.CloseAll(conn)

		found, err := Exists(context.TODO(), client, "key")
		assert.Error(t, err)
		assert.Equal(t, false, found)
	})
}

// ExampleExists is an example of the method Exists()
func ExampleExists() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = Set(context.Background(), client, testKey, testStringValue, testDependantKey)

	// Get the value
	_, _ = Exists(context.Background(), client, testKey)
	fmt.Print("key exists")
	// Output:key exists
}

// TestExpire is testing the method Expire()
func TestExpire(t *testing.T) {

	t.Run("expire command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase   string
			key        string
			expiration time.Duration
		}{
			{"regular key", "test-set-exp", 2 * time.Second},
			{"lots of time", "test-set2", 200 * time.Hour},
			{"no time", "test-set3", 0},
			{"no key name", "", 2 * time.Second},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				expireCmd := conn.Command(ExpireCommand, test.key, int64(test.expiration.Seconds()))

				err := Expire(context.Background(), client, test.key, test.expiration)
				assert.NoError(t, err)
				assert.Equal(t, true, expireCmd.Called)
			})
		}
	})

	t.Run("expire command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = SetExp(context.Background(), client, testKey, testStringValue, 5*time.Second, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var testVal string
		testVal, err = Get(context.Background(), client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testStringValue, testVal)

		// Expire
		err = Expire(context.Background(), client, testKey, 1*time.Second)
		if err != nil {
			t.Fatal(err.Error())
		}

		// Wait a few seconds and test
		t.Log("sleeping for 2 seconds...")
		time.Sleep(time.Second * 2)

		// Check that the key is expired
		testVal, err = Get(context.Background(), client, testKey)
		assert.Error(t, err)
		assert.Equal(t, redis.ErrNil, err)
		assert.Equal(t, "", testVal)
	})

	t.Run("expire cmd, trigger context err", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		client.CloseAll(conn)

		err := Expire(context.TODO(), client, "key", 10*time.Second)
		assert.Error(t, err)
	})
}

// ExampleExpire is an example of the method Expire()
func ExampleExpire() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = Set(context.Background(), client, testKey, testStringValue, testDependantKey)

	// Fire the command
	_ = Expire(context.Background(), client, testKey, 1*time.Minute)
	fmt.Printf("expiration on key: %s set for: %v", testKey, 1*time.Minute)
	// Output:expiration on key: test-key-name set for: 1m0s
}

// TestDestroyCache is testing the method DestroyCache()
func TestDestroyCache(t *testing.T) {

	t.Run("destroy cache / flush all command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		conn.Clear()

		// The main command to test
		destroyCmd := conn.Command(FlushAllCommand)

		err := DestroyCache(context.Background(), client)
		assert.NoError(t, err)
		assert.Equal(t, true, destroyCmd.Called)
	})

	t.Run("destroy cache / flush all command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = Set(context.Background(), client, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Test getting a value
		var val string
		val, err = Get(context.Background(), client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, val, testStringValue)

		// Check that the command worked
		err = DestroyCache(context.Background(), client)
		assert.NoError(t, err)

		// Value should not exist
		val, err = Get(context.Background(), client, testKey)
		assert.Error(t, err)
		assert.Equal(t, err, redis.ErrNil)
		assert.Equal(t, val, "")
	})

	t.Run("destroy cache cmd, trigger context err", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		client.CloseAll(conn)

		err := DestroyCache(context.TODO(), client)
		assert.Error(t, err)
	})
}

// ExampleDestroyCache is an example of the method DestroyCache()
func ExampleDestroyCache() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Fire the command
	_ = DestroyCache(context.Background(), client)
	fmt.Print("cache destroyed")
	// Output:cache destroyed
}

// TestGetList test the method GetList()
func TestGetList(t *testing.T) {

	t.Run("get list command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase           string
			key                string
			inputList          []string
			expectedList       []interface{}
			expectedStringList []string
		}{
			{
				"empty list",
				"test-set",
				[]string{""},
				[]interface{}{""},
				[]string{""},
			},
			{
				"one item",
				"test-set",
				[]string{"1"},
				[]interface{}{[]byte("1")},
				[]string{"1"},
			},
			{
				"multiple items",
				"test-set",
				[]string{"1", "1"},
				[]interface{}{[]byte("1"), []byte("1")},
				[]string{"1", "1"},
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				getCmd := conn.Command(ListRangeCommand, test.key, 0, -1).Expect(test.expectedList)

				list, err := GetList(context.Background(), client, test.key)
				assert.NoError(t, err)
				assert.Equal(t, true, getCmd.Called)
				assert.Equal(t, test.expectedStringList, list)
			})
		}
	})

	t.Run("get list command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = SetList(context.Background(), client, testKey, []string{testStringValue})
		assert.NoError(t, err)

		// Check that the command worked
		var list []string
		list, err = GetList(context.Background(), client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, []string{testStringValue}, list)
	})

	t.Run("get list cmd, trigger context err", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		client.CloseAll(conn)

		val, err := GetList(context.TODO(), client, "123456")
		assert.Error(t, err)
		assert.Equal(t, []string(nil), val)
	})
}

// ExampleGetList is an example of the method GetList()
func ExampleGetList() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = SetList(context.Background(), client, testKey, []string{testStringValue})

	// Fire the command
	_, _ = GetList(context.Background(), client, testKey)
	fmt.Printf("got list: %v", []string{testStringValue})
	// Output:got list: [test-string-value]
}

// TestSetList test the method SetList()
func TestSetList(t *testing.T) {

	t.Run("set list command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase  string
			key       string
			inputList []string
		}{
			{
				"empty list",
				"test-set",
				[]string{""},
			},
			{
				"one item",
				"test-set",
				[]string{"1"},
			},
			{
				"multiple items",
				"test-set",
				[]string{"1", "1"},
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// Create the arguments
				args := make([]interface{}, len(test.inputList)+1)
				args[0] = test.key

				// Loop members
				for i, param := range test.inputList {
					args[i+1] = param
				}

				// The main command to test
				setCmd := conn.Command(ListPushCommand, args...)

				err := SetList(context.Background(), client, test.key, test.inputList)
				assert.NoError(t, err)
				assert.Equal(t, true, setCmd.Called)
			})
		}
	})

	t.Run("set list command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = SetList(context.Background(), client, testKey, []string{testStringValue})
		assert.NoError(t, err)

		// Check that the command worked
		var list []string
		list, err = GetList(context.Background(), client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, []string{testStringValue}, list)
	})

	t.Run("set list cmd, trigger context err", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		client.CloseAll(conn)

		err := SetList(context.TODO(), client, "123456", []string{"test", "test1"})
		assert.Error(t, err)
	})
}

// ExampleSetList is an example of the method SetList()
func ExampleSetList() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = SetList(context.Background(), client, testKey, []string{testStringValue})

	// Fire the command
	_, _ = GetList(context.Background(), client, testKey)
	fmt.Printf("got list: %v", []string{testStringValue})
	// Output:got list: [test-string-value]
}

// TestDeleteWithoutDependency test the method DeleteWithoutDependency()
func TestDeleteWithoutDependency(t *testing.T) {

	t.Run("delete without using dependencies using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase     string
			keys         []string
			totalDeleted int
		}{
			{
				"empty list",
				[]string{},
				0,
			},
			{
				"one item",
				[]string{testKey},
				1,
			},
			{
				"multiple items",
				[]string{testKey, testKey + "2"},
				2,
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				var commands []*redigomock.Cmd
				for _, key := range test.keys {
					cmd := conn.Command(DeleteCommand, key)
					commands = append(commands, cmd)
				}

				total, err := DeleteWithoutDependency(context.Background(), client, test.keys...)
				assert.NoError(t, err)
				assert.Equal(t, test.totalDeleted, total)
				for _, c := range commands {
					assert.Equal(t, true, c.Called)
				}
			})
		}
	})

	t.Run("delete without using dependencies using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Set a key
		err = Set(context.Background(), client, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Fire the command
		var total int
		total, err = DeleteWithoutDependency(context.Background(), client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)

		// Check that the command worked
		var val string
		val, err = Get(context.Background(), client, testKey)
		assert.Error(t, err)
		assert.Equal(t, redis.ErrNil, err)
		assert.Equal(t, "", val)
	})

	t.Run("expire cmd, trigger context err", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		client.CloseAll(conn)

		total, err := DeleteWithoutDependency(context.TODO(), client, "key")
		assert.Error(t, err)
		assert.Equal(t, 0, total)
	})
}

// ExampleDeleteWithoutDependency is an example of the method DeleteWithoutDependency()
func ExampleDeleteWithoutDependency() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = Set(context.Background(), client, testKey, testStringValue)
	_ = Set(context.Background(), client, testKey+"2", testStringValue)

	// Delete keys
	_, _ = DeleteWithoutDependency(context.Background(), client, testKey, testKey+"2")
	fmt.Printf("deleted keys: %d", 2)
	// Output:deleted keys: 2
}

// TestSetToJSON is testing the method SetToJSON()
func TestSetToJSON(t *testing.T) {

	t.Run("set to json command using mocked redis (valid)", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase     string
			key          string
			modelData    interface{}
			dependencies []string
		}{
			{"key with dependencies", testKey, struct {
				testFieldString string
				testFieldInt    int
				testFieldFloat  float64
				testFieldBool   bool
			}{
				testFieldString: "test-value",
				testFieldInt:    123,
				testFieldFloat:  123.123,
				testFieldBool:   true,
			},
				[]string{testDependantKey},
			},
			{"key with no dependencies", testKey, struct {
				testFieldString string
				testFieldInt    int
				testFieldFloat  float64
				testFieldBool   bool
			}{
				testFieldString: "test-value",
				testFieldInt:    123,
				testFieldFloat:  123.123,
				testFieldBool:   true,
			},
				[]string{},
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				responseBytes, err := json.Marshal(&test.modelData)
				assert.NoError(t, err)

				var commands []*redigomock.Cmd

				// The main command to test
				commands = append(commands, conn.Command(SetCommand, test.key, string(responseBytes)))

				// Loop for each dependency
				if len(test.dependencies) > 0 {
					commands = append(commands, conn.Command(MultiCommand))
					for _, dep := range test.dependencies {
						commands = append(commands, conn.Command(AddToSetCommand, DependencyPrefix+dep, test.key))
					}
					commands = append(commands, conn.Command(ExecuteCommand))

					err = SetToJSONRaw(conn, test.key, test.modelData, 0, test.dependencies...)
					assert.NoError(t, err)
				} else {
					err = SetToJSONRaw(conn, test.key, test.modelData, 0, test.dependencies...)
					assert.NoError(t, err)
				}

				for _, c := range commands {
					assert.Equal(t, true, c.Called)
				}
			})
		}
	})

	t.Run("set to json command using mocked redis (valid) (exp)", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase     string
			key          string
			modelData    interface{}
			dependencies []string
			expiration   time.Duration
		}{
			{"key with dependencies", testKey, struct {
				testFieldString string
				testFieldInt    int
				testFieldFloat  float64
				testFieldBool   bool
			}{
				testFieldString: "test-value",
				testFieldInt:    123,
				testFieldFloat:  123.123,
				testFieldBool:   true,
			},
				[]string{testDependantKey},
				10 * time.Second,
			},
			{"key with no dependencies", testKey, struct {
				testFieldString string
				testFieldInt    int
				testFieldFloat  float64
				testFieldBool   bool
			}{
				testFieldString: "test-value",
				testFieldInt:    123,
				testFieldFloat:  123.123,
				testFieldBool:   true,
			},
				[]string{},
				10 * time.Second,
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				responseBytes, err := json.Marshal(&test.modelData)
				assert.NoError(t, err)

				var commands []*redigomock.Cmd

				// The main command to test
				commands = append(commands, conn.Command(SetExpirationCommand, test.key, int64(test.expiration.Seconds()), string(responseBytes)))

				// Loop for each dependency
				if len(test.dependencies) > 0 {
					commands = append(commands, conn.Command(MultiCommand))
					for _, dep := range test.dependencies {
						commands = append(commands, conn.Command(AddToSetCommand, DependencyPrefix+dep, test.key))
					}
					commands = append(commands, conn.Command(ExecuteCommand))

					err = SetToJSONRaw(conn, test.key, test.modelData, test.expiration, test.dependencies...)
					assert.NoError(t, err)
				} else {
					err = SetToJSONRaw(conn, test.key, test.modelData, test.expiration, test.dependencies...)
					assert.NoError(t, err)
				}

				for _, c := range commands {
					assert.Equal(t, true, c.Called)
				}
			})
		}
	})

	t.Run("set to json command using mocked redis (invalid json)", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase     string
			key          string
			modelData    interface{}
			dependencies []string
		}{
			{
				"json error - infinite",
				testKey,
				math.Inf(1),
				[]string{},
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				err := SetToJSON(context.Background(), client, test.key, test.modelData, 0, test.dependencies...)
				assert.Error(t, err)
			})
		}
	})

	t.Run("set to json command using real redis (valid)", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase     string
			key          string
			modelData    interface{}
			dependencies []string
			expiration   time.Duration
		}{
			{"key with dependencies", testKey, struct {
				testFieldString string
				testFieldInt    int
				testFieldFloat  float64
				testFieldBool   bool
			}{
				testFieldString: "test-value",
				testFieldInt:    123,
				testFieldFloat:  123.123,
				testFieldBool:   true,
			},
				[]string{testDependantKey},
				10 * time.Second,
			},
			{"key with no dependencies", testKey, struct {
				testFieldString string
				testFieldInt    int
				testFieldFloat  float64
				testFieldBool   bool
			}{
				testFieldString: "test-value",
				testFieldInt:    123,
				testFieldFloat:  123.123,
				testFieldBool:   true,
			},
				[]string{},
				10 * time.Second,
			},
			{"key with no exp", testKey, struct {
				testFieldString string
				testFieldInt    int
				testFieldFloat  float64
				testFieldBool   bool
			}{
				testFieldString: "test-value",
				testFieldInt:    123,
				testFieldFloat:  123.123,
				testFieldBool:   true,
			},
				[]string{},
				0,
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {

				// Start with a fresh db
				err = clearRealRedis(conn)
				assert.NoError(t, err)

				var responseBytes []byte
				responseBytes, err = json.Marshal(&test.modelData)
				assert.NoError(t, err)

				// Run command
				err = SetToJSONRaw(conn, test.key, test.modelData, test.expiration, test.dependencies...)
				assert.NoError(t, err)

				// Validate via getting the data from redis
				var testVal string
				testVal, err = GetRaw(conn, test.key)
				assert.NoError(t, err)
				assert.Equal(t, string(responseBytes), testVal)
			})
		}
	})

	t.Run("set to json cmd, trigger context err", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		client.CloseAll(conn)

		err := SetToJSON(context.TODO(), client, "123456", nil, 10*time.Second)
		assert.Error(t, err)
	})
}
