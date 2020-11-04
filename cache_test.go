package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
)

// Testing variables
const (
	// testKillDependencyHash   = "a648f768f57e73e2497ccaa113d5ad9e731c5cd8"
	testDependantKey         = "test-dependant-key-name"
	testHashName             = "test-hash-name"
	testIdleTimeout          = 240
	testKey                  = "test-key-name"
	testLocalConnectionURL   = "redis://localhost:6379"
	testMaxActiveConnections = 0
	testMaxConnLifetime      = 0
	testMaxIdleConnections   = 10
	testStringValue          = "test-string-value"
)

// loadMockRedis will load a mocked redis connection
func loadMockRedis() (conn *redigomock.Conn, pool *redis.Pool) {
	conn = redigomock.NewConn()
	pool = &redis.Pool{
		Dial:            func() (redis.Conn, error) { return conn, nil },
		IdleTimeout:     time.Duration(testIdleTimeout) * time.Second,
		MaxActive:       testMaxActiveConnections,
		MaxConnLifetime: testMaxConnLifetime,
		MaxIdle:         testMaxIdleConnections,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, doErr := c.Do(pingCommand)
			return doErr
		},
	}
	return
}

// loadRealRedis will load a real redis connection
func loadRealRedis() (conn redis.Conn, pool *redis.Pool, err error) {
	pool, err = Connect(
		testLocalConnectionURL,
		testMaxActiveConnections,
		testMaxIdleConnections,
		testMaxConnLifetime,
		testIdleTimeout,
		true,
	)
	if err != nil {
		return
	}

	conn = GetConnection(pool)
	return
}

// clearRealRedis will clear a real redis db
func clearRealRedis(conn redis.Conn) error {
	return DestroyCache(conn)
}

// endTest end tests the same way
func endTest(pool *redis.Pool, conn redis.Conn) {
	CloseAll(pool, conn)
}

/*// startTest start all tests the same way
func startTestCustom() (pool *redis.Pool, err error) {
	return Connect(
		testLocalConnectionURL,
		testMaxActiveConnections,
		testMaxIdleConnections,
		testMaxConnLifetime,
		testIdleTimeout,
		true,
		redis.DialKeepAlive(10*time.Second),
	)
}*/

// TestSet is testing the method Set()
func TestSet(t *testing.T) {

	t.Run("set command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		conn, pool := loadMockRedis()
		assert.NotNil(t, pool)
		defer endTest(pool, conn)

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

				// The main command to test
				setCmd := conn.Command(setCommand, test.key, test.value).Expect(test.value)

				// Loop for each dependency
				if len(test.dependencies) > 0 {
					multiCmd := conn.Command(multiCommand)
					for _, dep := range test.dependencies {
						_ = conn.Command(addToSetCommand, dependencyPrefix+dep, test.key)
					}
					exeCmd := conn.Command(executeCommand)

					err := Set(conn, test.key, test.value, test.dependencies...)
					assert.NoError(t, err)
					assert.Equal(t, true, multiCmd.Called)
					assert.Equal(t, true, setCmd.Called)
					assert.Equal(t, true, exeCmd.Called)
				} else {
					err := Set(conn, test.key, test.value, test.dependencies...)
					assert.NoError(t, err)
					assert.Equal(t, true, setCmd.Called)
				}
			})
		}
	})

	t.Run("set command using real redis", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		conn, pool, err := loadRealRedis()
		assert.NotNil(t, pool)
		assert.NoError(t, err)
		defer endTest(pool, conn)

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
				err = Set(conn, test.key, test.value, test.dependencies...)
				assert.NoError(t, err)

				// Validate via getting the data from redis
				var testVal string
				testVal, err = Get(conn, test.key)
				assert.NoError(t, err)
				assert.Equal(t, test.value, testVal)
			})
		}
	})
}

// ExampleSet is an example of the method Set()
func ExampleSet() {

	// Load a mocked redis for testing/examples
	conn, pool := loadMockRedis()

	// Close connections at end of request
	defer CloseAll(pool, conn)

	// Set the key/value
	_ = Set(conn, testKey, testStringValue, testDependantKey)
	fmt.Printf("set: %s value: %s dep key: %s", testKey, testStringValue, testDependantKey)
	// Output:set: test-key-name value: test-string-value dep key: test-dependant-key-name
}

// TestSetExp is testing the method SetExp()
func TestSetExp(t *testing.T) {

	t.Run("set exp command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		conn, pool := loadMockRedis()
		assert.NotNil(t, pool)
		defer endTest(pool, conn)

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

				// The main command to test
				setCmd := conn.Command(setExpirationCommand, test.key, int64(test.expiration.Seconds()), test.value).Expect(test.value)

				// Loop for each dependency
				if len(test.dependencies) > 0 {
					multiCmd := conn.Command(multiCommand)
					for _, dep := range test.dependencies {
						_ = conn.Command(addToSetCommand, dependencyPrefix+dep, test.key)
					}
					exeCmd := conn.Command(executeCommand)

					err := SetExp(conn, test.key, test.value, test.expiration, test.dependencies...)
					assert.NoError(t, err)
					assert.Equal(t, true, multiCmd.Called)
					assert.Equal(t, true, setCmd.Called)
					assert.Equal(t, true, exeCmd.Called)

				} else {
					err := SetExp(conn, test.key, test.value, test.expiration, test.dependencies...)
					assert.NoError(t, err)
					assert.Equal(t, true, setCmd.Called)
				}
			})
		}
	})

	t.Run("set exp command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		conn, pool, err := loadRealRedis()
		assert.NotNil(t, pool)
		assert.NoError(t, err)
		defer endTest(pool, conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = SetExp(conn, testKey, testStringValue, 2*time.Second, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var testVal string
		testVal, err = Get(conn, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testStringValue, testVal)

		// Wait a few seconds and test
		t.Log("sleeping for 3 seconds...")
		time.Sleep(time.Second * 3)

		// Check that the key is expired
		testVal, err = Get(conn, testKey)
		assert.Error(t, err)
		assert.Equal(t, "", testVal)
	})
}

// ExampleSetExp is an example of the method SetExp()
func ExampleSetExp() {
	// Load a mocked redis for testing/examples
	conn, pool := loadMockRedis()

	// Close connections at end of request
	defer CloseAll(pool, conn)

	// Set the key/value
	_ = SetExp(conn, testKey, testStringValue, 2*time.Minute, testDependantKey)
	fmt.Printf("set: %s value: %s exp: %v dep key: %s", testKey, testStringValue, 2*time.Minute, testDependantKey)
	// Output:set: test-key-name value: test-string-value exp: 2m0s dep key: test-dependant-key-name
}

// TestGet is testing the method Get()
func TestGet(t *testing.T) {

	t.Run("get command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		conn, pool := loadMockRedis()
		assert.NotNil(t, pool)
		defer endTest(pool, conn)

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
				getCmd := conn.Command(getCommand, test.key).Expect(test.value)

				val, err := Get(conn, test.key)
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
		conn, pool, err := loadRealRedis()
		assert.NotNil(t, pool)
		assert.NoError(t, err)
		defer endTest(pool, conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = Set(conn, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var testVal string
		testVal, err = Get(conn, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testStringValue, testVal)
	})
}

// ExampleGet is an example of the method Get()
func ExampleGet() {
	// Load a mocked redis for testing/examples
	conn, pool := loadMockRedis()

	// Close connections at end of request
	defer CloseAll(pool, conn)

	// Set the key/value
	_ = Set(conn, testKey, testStringValue, testDependantKey)

	// Get the value
	_, _ = Get(conn, testKey)
	fmt.Printf("got value: %s", testStringValue)
	// Output:got value: test-string-value
}

// TestGetBytes is testing the method GetBytes()
func TestGetBytes(t *testing.T) {

	t.Run("get bytes command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		conn, pool := loadMockRedis()
		assert.NotNil(t, pool)
		defer endTest(pool, conn)

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
				getCmd := conn.Command(getCommand, test.key).Expect([]byte(test.value))

				val, err := GetBytes(conn, test.key)
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
		conn, pool, err := loadRealRedis()
		assert.NotNil(t, pool)
		assert.NoError(t, err)
		defer endTest(pool, conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = Set(conn, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var testVal []byte
		testVal, err = GetBytes(conn, testKey)
		assert.NoError(t, err)
		assert.Equal(t, []byte(testStringValue), testVal)
	})
}

// ExampleGetBytes is an example of the method GetBytes()
func ExampleGetBytes() {
	// Load a mocked redis for testing/examples
	conn, pool := loadMockRedis()

	// Close connections at end of request
	defer CloseAll(pool, conn)

	// Set the key/value
	_ = Set(conn, testKey, testStringValue, testDependantKey)

	// Get the value
	_, _ = GetBytes(conn, testKey)
	fmt.Printf("got value: %s", testStringValue)
	// Output:got value: test-string-value
}

// TestGetAllKeys is testing the method GetAllKeys()
func TestGetAllKeys(t *testing.T) {

	t.Run("get all keys command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		conn, pool := loadMockRedis()
		assert.NotNil(t, pool)
		defer endTest(pool, conn)

		conn.Clear()

		// The main command to test
		getCmd := conn.Command(keysCommand, allKeysCommand).Expect([]interface{}{[]byte(testKey)})

		val, err := GetAllKeys(conn)
		assert.NoError(t, err)
		assert.Equal(t, true, getCmd.Called)
		assert.Equal(t, []string{testKey}, val)
	})

	t.Run("get all keys command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		conn, pool, err := loadRealRedis()
		assert.NotNil(t, pool)
		assert.NoError(t, err)
		defer endTest(pool, conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = Set(conn, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var keys []string
		keys, err = GetAllKeys(conn)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(keys))
	})
}

// ExampleGetAllKeys is an example of the method GetAllKeys()
func ExampleGetAllKeys() {
	// Load a mocked redis for testing/examples
	conn, pool := loadMockRedis()

	// Close connections at end of request
	defer CloseAll(pool, conn)

	// Set the key/value
	_ = Set(conn, testKey, testStringValue, testDependantKey)

	// Get the keys
	_, _ = GetAllKeys(conn)
	fmt.Printf("found keys: %d", len([]string{testKey, testDependantKey}))
	// Output:found keys: 2
}

// TestExists is testing the method Exists()
func TestExists(t *testing.T) {

	t.Run("exists command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		conn, pool := loadMockRedis()
		assert.NotNil(t, pool)
		defer endTest(pool, conn)

		conn.Clear()

		// todo: add table tests

		// The main command to test
		existsCmd := conn.Command(existsCommand, testKey).Expect(interface{}(int64(1)))

		val, err := Exists(conn, testKey)
		assert.NoError(t, err)
		assert.Equal(t, true, existsCmd.Called)
		assert.Equal(t, true, val)
	})

	t.Run("exists command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		conn, pool, err := loadRealRedis()
		assert.NotNil(t, pool)
		assert.NoError(t, err)
		defer endTest(pool, conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = Set(conn, testKey, testStringValue, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var found bool
		found, err = Exists(conn, testKey)
		assert.NoError(t, err)
		assert.Equal(t, true, found)
	})
}

// ExampleExists is an example of Exists() method
func ExampleExists() {
	// Load a mocked redis for testing/examples
	conn, pool := loadMockRedis()

	// Close connections at end of request
	defer CloseAll(pool, conn)

	// Set the key/value
	_ = Set(conn, testKey, testStringValue, testDependantKey)

	// Get the value
	_, _ = Exists(conn, testKey)
	fmt.Print("key exists")
	// Output:key exists
}

// TestExpire is testing the method Expire()
func TestExpire(t *testing.T) {

	t.Run("expire command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		conn, pool := loadMockRedis()
		assert.NotNil(t, pool)
		defer endTest(pool, conn)

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
				expireCmd := conn.Command(expireCommand, test.key, int64(test.expiration.Seconds()))

				err := Expire(conn, test.key, test.expiration)
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
		conn, pool, err := loadRealRedis()
		assert.NotNil(t, pool)
		assert.NoError(t, err)
		defer endTest(pool, conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = SetExp(conn, testKey, testStringValue, 5*time.Second, testDependantKey)
		assert.NoError(t, err)

		// Check that the command worked
		var testVal string
		testVal, err = Get(conn, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testStringValue, testVal)

		// Expire
		err = Expire(conn, testKey, 1*time.Second)
		if err != nil {
			t.Fatal(err.Error())
		}

		// Wait a few seconds and test
		t.Log("sleeping for 2 seconds...")
		time.Sleep(time.Second * 2)

		// Check that the key is expired
		testVal, err = Get(conn, testKey)
		assert.Error(t, err)
		assert.Equal(t, "", testVal)
	})
}

// ExampleExpire is an example of Expire() method
func ExampleExpire() {
	// Load a mocked redis for testing/examples
	conn, pool := loadMockRedis()

	// Close connections at end of request
	defer CloseAll(pool, conn)

	// Set the key/value
	_ = Set(conn, testKey, testStringValue, testDependantKey)

	// Set the expire
	_ = Expire(conn, testKey, 1*time.Minute)
	fmt.Printf("expiration on key: %s set for: %v", testKey, 1*time.Minute)
	// Output:expiration on key: test-key-name set for: 1m0s
}

/*
// TestDestroyCache is testing the DestroyCache() method
func TestDestroyCache(t *testing.T) {

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Set
	err := Set("test-destroy", testStringValue, "another-key")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Check the set
	val, _ := Get("test-destroy")
	if val != testStringValue {
		t.Fatalf("expected value: %s, got: %s", testStringValue, val)
	}

	// Fire destroy
	if err = DestroyCache(); err != nil {
		t.Fatal(err.Error())
	}

	// Check the destroy
	if val, _ = Get("test-destroy"); val == testStringValue {
		t.Fatalf("expected value: %s, got: %s", "", val)
	}
}

// ExampleDestroyCache is an example of DestroyCache() method
func ExampleDestroyCache() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-destroy-cache", testStringValue, "another-key")

	// Set the expire
	_ = DestroyCache()
	fmt.Print("cache destroyed")
	// Output: cache destroyed
}

// ExampleDelete is an example of Delete() method
func ExampleDelete() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-destroy-cache", testStringValue, "another-key")

	// Delete keys
	total, _ := Delete("another-key", "another-key-2")
	fmt.Print(total, " deleted keys")
	// Output: 2 deleted keys
}

// ExampleDeleteWithoutDependency is an example of DeleteWithoutDependency() method
func ExampleDeleteWithoutDependency() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-destroy-cache-1", testStringValue)
	_ = Set("example-destroy-cache-2", testStringValue)

	// Delete keys
	total, _ := DeleteWithoutDependency("example-destroy-cache-1", "example-destroy-cache-2")
	fmt.Print(total, " deleted keys")
	// Output: 2 deleted keys
}

// TestKillByDependency will test KillByDependency()
func TestKillByDependency(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Test no keys
	_, err := KillByDependency()
	if err != nil {
		t.Errorf("error occurred %s", err.Error())
	}

	// Test  keys
	if _, err = KillByDependency("key1"); err != nil {
		t.Errorf("error occurred %s", err.Error())
	}

	// Test  keys
	if _, err = KillByDependency("key1", "key two; another"); err != nil {
		t.Errorf("error occurred %s", err.Error())
	}
}

// ExampleKillByDependency is an example of KillByDependency() method
func ExampleKillByDependency() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-destroy-cache", testStringValue, "another-key")

	// Delete keys
	_, _ = KillByDependency("another-key", "another-key-2")
	fmt.Print("deleted keys")
	// Output: deleted keys
}

// TestDependencyManagement tests basic dependency functionality
// Tests a myriad of methods
func TestDependencyManagement(t *testing.T) {

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Set a key with two dependent keys
	err := Set("test-set-dep", testStringValue, "dependent-1", "dependent-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Test for dependent key 1
	var ok bool
	if ok, err = SetIsMember("depend:dependent-1", "test-set-dep"); err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Test for dependent key 2
	if ok, err = SetIsMember("depend:dependent-2", "test-set-dep"); err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Kill a dependent key
	var total int
	if total, err = Delete("dependent-1"); err != nil {
		t.Fatal(err.Error())
	} else if total != 2 {
		t.Fatal("expected 2 keys to be removed", total)
	}

	// Test for main key
	var found bool
	if found, err = Exists("test-set-dep"); err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}

	// Test for dependency relation
	if found, err = Exists("depend:dependent-1"); err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}

	// Test for dependent key 2
	if ok, err = SetIsMember("depend:dependent-2", "test-set-dep"); err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Kill all dependent keys
	if total, err = KillByDependency("dependent-1", "dependent-2"); err != nil {
		t.Fatal(err.Error())
	} else if total != 1 {
		t.Fatal("expected 1 key to be removed", total)
	}

	// Test for dependency relation
	if found, err = Exists("depend:dependent-2"); err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}

	// Test for main key
	if found, err = Exists("test-set-dep"); err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}
}

// TestHashMapDependencyManagement tests HASH map dependency functionality
// Tests a myriad of methods
func TestHashMapDependencyManagement(t *testing.T) {

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Create pairs
	pairs := [][2]interface{}{
		{"pair-1", testPairValue},
		{"pair-2", "pair-2-value"},
		{"pair-3", "pair-3-value"},
	}

	// Set the hash map
	err := HashMapSet("test-hash-map-dependency", pairs, "test-hash-map-depend-1", "test-hash-map-depend-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	var val string
	val, err = HashGet("test-hash-map-dependency", "pair-1")
	if err != nil {
		t.Fatal(err.Error())
	} else if val != testPairValue {
		t.Fatal("expected value was wrong")
	}

	// Get a key in the map
	var values []string
	values, err = HashMapGet("test-hash-map-dependency", "pair-1", "pair-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Got two values?
	if len(values) != 2 {
		t.Fatal("expected 2 values", values, len(values))
	}

	// Test for dependent key 1
	var ok bool
	ok, err = SetIsMember("depend:test-hash-map-depend-1", "test-hash-map-dependency")
	if err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Test for dependent key 2
	ok, err = SetIsMember("depend:test-hash-map-depend-2", "test-hash-map-dependency")
	if err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Kill a dependent key
	var total int
	total, err = Delete("test-hash-map-depend-2")
	if err != nil {
		t.Fatal(err.Error())
	} else if total != 2 {
		t.Fatal("expected 2 keys to be removed", total)
	}

	// Test for main key
	var found bool
	found, err = Exists("test-hash-map-dependency")
	if err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}
}

// TestSetAddMany test the SetAddMany() method
func TestSetAddMany(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	var empty []interface{}
	empty = append(empty, "one", "two", "three")

	err := SetAddMany("test-set", empty...)
	if err != nil {
		t.Fatal("error setting members", err.Error())
	}

	// Test the set
	var found bool
	found, err = SetIsMember("test-set", "two")
	if err != nil {
		t.Fatal("error in SetIsMember", err.Error())
	} else if !found {
		t.Fatal("failed to find a member that should exist")
	}

	// Test for not finding a value
	found, err = SetIsMember("test-set", "not-here")
	if err != nil {
		t.Fatal("error in SetIsMember", err.Error())
	} else if found {
		t.Fatal("found a member that should NOT exist")
	}
}

// TestSetRemoveMember test the SetRemoveMember() method
func TestSetRemoveMember(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	var empty []interface{}
	empty = append(empty, "one", "two", "three")

	err := SetAddMany("test-set", empty...)
	if err != nil {
		t.Fatal("error setting members", err.Error())
	}

	// Try to delete
	err = SetRemoveMember("test-set", "two")
	if err != nil {
		t.Fatal("error SetRemoveMember", err.Error())
	}

	// Test for not finding a value
	var found bool
	found, err = SetIsMember("test-set", "two")
	if err != nil {
		t.Fatal("error in SetIsMember", err.Error())
	} else if found {
		t.Fatal("found a member that should NOT exist")
	}
}

// TestSetAdd test the SetAdd() method
func TestSetAdd(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	var tests = []struct {
		name          string
		member        interface{}
		dependencies  string
		expectedError bool
	}{
		{"test-set", testStringValue, "another-key", false},
		{"test-set", []string{"one", "two", "three"}, "another-key", false},
		{"test-set", []int{1, 2, 3}, "another-key", false},
		{"test-set", "", "another-key", false},
		{"test-set", "", "", false},
		{"", "", "", false},
	}

	// Test all
	for _, test := range tests {
		if err := SetAdd(test.name, test.member, test.dependencies); err != nil && !test.expectedError {
			t.Errorf("%s Failed: [%s] inputted and [%v], error [%s]", t.Name(), test.name, test.member, err.Error())
		} else if err == nil && test.expectedError {
			t.Errorf("%s Failed: [%s] inputted and [%v], error was expected but did not occur", t.Name(), test.name, test.member)
		}
	}
}

// TestSetList test the SetList() method
func TestSetList(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	var tests = []struct {
		key           string
		list          []string
		expectedError bool
	}{
		{"test-set", []string{"val1", "val2"}, false},
		{"test-set-bad-empty", []string{""}, false},
		{"test-set-bad-no-list", []string{}, true},
	}

	// Test all
	for _, test := range tests {
		if err := SetList(test.key, test.list); err != nil && !test.expectedError {
			t.Errorf("%s Failed: [%s] inputted and [%v], error [%s]", t.Name(), test.key, test.list, err.Error())
		} else if err == nil && test.expectedError {
			t.Errorf("%s Failed: [%s] inputted and [%v], error was expected but did not occur", t.Name(), test.key, test.list)
		}
	}
}

// TestGetList test the GetList() method
func TestGetList(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	var tests = []struct {
		key           string
		input         []string
		expected      []string
		expectedError bool
	}{
		{"test-set", []string{}, []string{}, false},
		{"test-set", []string{"1"}, []string{"1"}, false},
		{"test-set", []string{"1", "1"}, []string{"1", "1", "1"}, false},
		{"test-set", []string{}, []string{"1", "1", "1"}, false},
		{"test-set", []string{""}, []string{"1", "1", "1", ""}, false},
	}

	// Test all
	for _, test := range tests {

		_ = SetList(test.key, test.input)

		if list, err := GetList(test.key); err != nil && !test.expectedError {
			t.Errorf("%s Failed: [%s] inputted and expected [%v], result [%s] error [%s]", t.Name(), test.input, test.expected, list, err.Error())
		} else if err == nil && test.expectedError {
			t.Errorf("%s Failed: [%s] inputted and expected  [%v], result [%s], error was expected but did not occur", t.Name(), test.input, test.expected, list)
		} else if len(list) != len(test.expected) {
			t.Errorf("%s Failed: [%s] inputted and expected [%v], result [%s]", t.Name(), test.input, test.expected, list)
		}
	}
}

func TestGetOrSetWithExpirationGob(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

}
*/
