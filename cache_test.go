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
	// testPairValue            = "pair-1-value"
	testDependantKey         = "test-dependant-key-name"
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

// TestSet is testing the Set() method
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
			{"key with no dependencies", testKey, testStringValue, []string{""}},
			{"key with empty value", testKey, "", []string{""}},
			{"key with spaces", "key name", "some val", []string{""}},
			{"key with symbols", ".key name;!()\\", "", []string{""}},
			{"key with symbols and value as symbols", ".key name;!()\\", `\ / ; [ ] { }!`, []string{""}},
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

// ExampleSet is an example of Set() method
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

// TestSetExp is testing the SetExp() method
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
			{"key with no dependencies", "test-set2", testStringValue, 2 * time.Second, []string{""}},
			{"key with empty value", "test-set3", "", 2 * time.Second, []string{""}},
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

		// Fire the command
		err = SetExp(conn, testKey, testStringValue, 2*time.Second, testDependantKey)
		assert.NoError(t, err)

		// Check that set worked
		var testVal string
		testVal, err = Get(conn, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testStringValue, testVal)

		// Wait 2 seconds and test
		t.Log("sleeping for 3 seconds...")
		time.Sleep(time.Second * 3)

		// Check that the key is expired
		testVal, err = Get(conn, testKey)
		assert.Error(t, err)
		assert.Equal(t, "", testVal)
	})
}

// ExampleSetExp is an example of SetExp() method
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

/*

// TestHashSet is testing the HashSet() method
func TestHashSet(t *testing.T) {

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	var tests = []struct {
		name          string
		key           string
		value         interface{}
		dependencies  string
		expectedError bool
	}{
		{"test-hash-name", "test-hash-key", "my-cache-value", "another-key", false},
		{"test-hash-name1", "test-hash-key", "my-cache-value", "another-key", false},
		{"test-hash-name2", "test-hash-key", "my-cache-value", "", false},
		{"test-hash-name3", "test-hash-key", "", "", false},
		{"test-hash-name4", "", "", "", false},
		{"", "", "", "", false},
		{"", "", []string{""}, "", false},
		{"-", "-", []string{""}, "-", false},
		{"-", "-", map[string]string{}, "-", false},
	}

	// Test all
	for _, test := range tests {
		if err := HashSet(test.name, test.key, test.value, test.dependencies); err != nil && !test.expectedError {
			t.Errorf("%s Failed: [%s] inputted and [%s] and [%v], error [%s]", t.Name(), test.name, test.key, test.value, err.Error())
		} else if err == nil && test.expectedError {
			t.Errorf("%s Failed: [%s] inputted and [%s] and [%v], error was expected but did not occur", t.Name(), test.name, test.key, test.value)
		}
	}

	// Get the value
	val, err := HashGet("test-hash-name", "test-hash-key")
	if err != nil {
		t.Fatal(err.Error())
	} else if val != "my-cache-value" {
		t.Fatal("value returned was wrong", val)
	}
}

// ExampleHashSet is an example of HashSet() method
func ExampleHashSet() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = HashSet("example-hash-set", "test-hash-key", "my-cache-value")
	fmt.Print("set complete")
	// Output: set complete
}

// TestHashGet is testing the HashGet() method
func TestHashGet(t *testing.T) {

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Set the hash
	err := HashSet("test-hash-name", "test-hash-key", "my-cache-value")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Get the value
	var val string
	val, err = HashGet("test-hash-name", "test-hash-key")
	if err != nil {
		t.Fatal(err.Error())
	} else if val != "my-cache-value" {
		t.Fatal("value returned was wrong", val)
	}
}

// ExampleHashGet is an example of HashGet() method
func ExampleHashGet() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = HashSet("example-hash-get", "test-hash-key", "my-cache-value")

	// Get the value
	val, _ := HashGet("example-hash-get", "test-hash-key")
	fmt.Print(val)
	// Output: my-cache-value
}

// TestHashMapSet is testing the HashMapSet() method
func TestHashMapSet(t *testing.T) {

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
	err := HashMapSet("test-hash-map-set", pairs, "test-hash-map-1", "test-hash-map-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	var val string
	val, err = HashGet("test-hash-map-set", "pair-1")
	if err != nil {
		t.Fatal(err.Error())
	} else if val != testPairValue {
		t.Fatal("expected value was wrong")
	}

	// Get a key in the map
	var values []string
	values, err = HashMapGet("test-hash-map-set", "pair-1", "pair-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Got two values?
	if len(values) != 2 {
		t.Fatal("expected 2 values", values, len(values))
	}

	// Test value 1
	if values[0] != testPairValue {
		t.Fatal("expected value", values[0])
	}

	// Test value 2
	if values[1] != "pair-2-value" {
		t.Fatal("expected value", values[1])
	}
}

// ExampleHashMapSet is an example of HashMapSet() method
func ExampleHashMapSet() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Create pairs
	pairs := [][2]interface{}{
		{"pair-1", testPairValue},
		{"pair-2", "pair-2-value"},
		{"pair-3", "pair-3-value"},
	}

	// Set the hash map
	_ = HashMapSet("example-hash-map-set", pairs, "test-hash-map-1", "test-hash-map-2")
	fmt.Print("set complete")
	// Output: set complete
}

// TestHashMapSetExp is testing the HashMapSetExp() method
func TestHashMapSetExp(t *testing.T) {

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
	err := HashMapSetExp("test-hash-map-set-expire", pairs, 2*time.Second, "test-hash-map-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	var val string
	val, err = HashGet("test-hash-map-set-expire", "pair-1")
	if err != nil {
		t.Fatal(err.Error())
	} else if val != testPairValue {
		t.Fatal("expected value was wrong")
	}

	// Wait 2 seconds and test
	time.Sleep(time.Second * 3)

	val, err = HashGet("test-hash-map-set-expire", "pair-1")
	if err == nil {
		t.Fatal("expected error: redigo: nil returned", err)
	} else if val != "" {
		t.Fatal("expected value to be empty")
	}
}

// ExampleHashMapSetExp is an example of HashMapSetExp() method
func ExampleHashMapSetExp() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Create pairs
	pairs := [][2]interface{}{
		{"pair-1", testPairValue},
		{"pair-2", "pair-2-value"},
		{"pair-3", "pair-3-value"},
	}

	// Set the hash map
	_ = HashMapSetExp("example-hash-map-set-exp", pairs, 2*time.Minute, "test-hash-map-1", "test-hash-map-2")
	fmt.Print("set complete")
	// Output: set complete
}

// TestGet is testing the Get() method
func TestGet(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Set
	err := Set("test-get", testStringValue)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Get the value
	val, _ := Get("test-get")
	if val != testStringValue {
		t.Fatalf("expected value: %s, got: %s", testStringValue, val)
	}
}

// ExampleGet is an example of Get() method
func ExampleGet() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-get", testStringValue, "another-key")

	// Get the value
	value, _ := Get("example-get")
	fmt.Print(value)
	// Output: test-string-value
}

// TestGetBytes is testing the GetBytes() method
func TestGetBytes(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Set
	err := Set("test-get-bytes", testStringValue)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Get the value
	val, _ := GetBytes("test-get-bytes")
	if string(val) != testStringValue {
		t.Fatalf("expected value: %s, got: %s", testStringValue, val)
	}
}

// ExampleGetBytes is an example of GetBytes() method
func ExampleGetBytes() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-get-bytes", testStringValue, "another-key")

	// Get the value
	value, _ := GetBytes("example-get-bytes")
	fmt.Print(string(value))
	// Output: test-string-value
}

// TestGetAllKeys is testing the GetAllKeys() method
func TestGetAllKeys(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Set
	err := Set("test-get", testStringValue)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Get the value
	var keys []string
	keys, err = GetAllKeys()
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(keys) == 0 {
		t.Fatal("expected to have at least 1 key")
	}
}

// ExampleGetAllKeys is an example of GetAllKeys() method
func ExampleGetAllKeys() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-get-all-keys", testStringValue, "another-key")

	// Get the value
	_, _ = GetAllKeys()
	fmt.Print("found keys")
	// Output: found keys
}

// TestExists is testing the Exists() method
func TestExists(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Set
	err := Set("test-exists", testStringValue)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Check the set / exists
	exists, _ := Exists("test-exists")
	if !exists {
		t.Fatal("expected key to exist")
	}
}

// ExampleExists is an example of Exists() method
func ExampleExists() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-exists", testStringValue, "another-key")

	// Get the value
	_, _ = Exists("example-exists")
	fmt.Print("key exists")
	// Output: key exists
}

// TestExpire is testing the Expire() method
func TestExpire(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Set
	err := SetExp("test-set-expire", testStringValue, 1*time.Minute, "another-key")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Check the set
	val, _ := Get("test-set-expire")
	if val != testStringValue {
		t.Fatalf("expected value: %s, got: %s", testStringValue, val)
	}

	// Fire the expire
	err = Expire("test-set-expire", 1*time.Second)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Wait 2 seconds and test
	time.Sleep(time.Second * 2)

	// Check the expire
	val, _ = Get("test-set-expire")
	if val == testStringValue {
		t.Fatalf("expected value: %s, got: %s", "", val)
	}
}

// ExampleExpire is an example of Expire() method
func ExampleExpire() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-expire", testStringValue, "another-key")

	// Set the expire
	_ = Expire("example-expire", 1*time.Minute)
	fmt.Print("expiration set")
	// Output: expiration set
}

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
