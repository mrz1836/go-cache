package cache

import (
	"fmt"
	"testing"
	"time"
)

// Testing variables
var (
	connectionURL        = "redis://localhost:6379"
	idleTimeout          = 240
	killDependencyHash   = "a648f768f57e73e2497ccaa113d5ad9e731c5cd8"
	maxActiveConnections = 0
	maxConnLifetime      = 0
	maxIdleConnections   = 10
)

// startTest start all tests the same way
func startTest() error {
	if GetPool() == nil {
		return Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)
	}
	return nil
}

// endTest end tests the same way
func endTest() {
	Disconnect()
}

// TestSet is testing the Set() method
func TestSet(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Set the key/value
	err := Set("test-set", "my-value", "another-key")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Check the set via a Get
	val, err := Get("test-set")
	if val != "my-value" {
		t.Fatalf("expected value: %s, got: %s", "my-value", val)
	}
}

// TestSetExp is testing the SetExp() method
func TestSetExp(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Set
	err := SetExp("test-set-exp", "my-value", 2*time.Second, "another-key")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Check the set
	val, err := Get("test-set-exp")
	if val != "my-value" {
		t.Fatalf("expected value: %s, got: %s", "my-value", val)
	}

	// Wait 2 seconds and test
	time.Sleep(time.Second * 3)

	// Check the set expire
	val, err = Get("test-set-exp")
	if val == "my-value" {
		t.Fatalf("expected value: %s, got: %s", "", val)
	}
}

// ExampleSet is an example of Set() method
func ExampleSet() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-set", "my-value", "another-key")
	fmt.Print("set complete")
	//Output: set complete
}

// ExampleSetExp is an example of SetExp() method
func ExampleSetExp() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = SetExp("example-set-exp", "my-value", 2*time.Minute, "another-key")
	fmt.Print("set complete")
	//Output: set complete
}

// TestHashSet is testing the HashSet() method
func TestHashSet(t *testing.T) {

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
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = HashSet("example-hash-set", "test-hash-key", "my-cache-value")
	fmt.Print("set complete")
	//Output: set complete
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
	val, err := HashGet("test-hash-name", "test-hash-key")
	if err != nil {
		t.Fatal(err.Error())
	} else if val != "my-cache-value" {
		t.Fatal("value returned was wrong", val)
	}
}

// ExampleHashGet is an example of HashGet() method
func ExampleHashGet() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = HashSet("example-hash-get", "test-hash-key", "my-cache-value")

	// Get the value
	val, _ := HashGet("example-hash-get", "test-hash-key")
	fmt.Print(val)
	//Output: my-cache-value
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
		{"pair-1", "pair-1-value"},
		{"pair-2", "pair-2-value"},
		{"pair-3", "pair-3-value"},
	}

	// Set the hash map
	err := HashMapSet("test-hash-map-set", pairs, "test-hash-map-1", "test-hash-map-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	val, err := HashGet("test-hash-map-set", "pair-1")
	if err != nil {
		t.Fatal(err.Error())
	} else if val != "pair-1-value" {
		t.Fatal("expected value was wrong")
	}

	// Get a key in the map
	values, err := HashMapGet("test-hash-map-set", "pair-1", "pair-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Got two values?
	if len(values) != 2 {
		t.Fatal("expected 2 values", values, len(values))
	}

	// Test value 1
	if values[0] != "pair-1-value" {
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
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)

	// Disconnect at end
	defer Disconnect()

	// Create pairs
	pairs := [][2]interface{}{
		{"pair-1", "pair-1-value"},
		{"pair-2", "pair-2-value"},
		{"pair-3", "pair-3-value"},
	}

	// Set the hash map
	_ = HashMapSet("example-hash-map-set", pairs, "test-hash-map-1", "test-hash-map-2")
	fmt.Print("set complete")
	//Output: set complete
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
		{"pair-1", "pair-1-value"},
		{"pair-2", "pair-2-value"},
		{"pair-3", "pair-3-value"},
	}

	// Set the hash map
	err := HashMapSetExp("test-hash-map-set-expire", pairs, 2*time.Second, "test-hash-map-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	val, err := HashGet("test-hash-map-set-expire", "pair-1")
	if err != nil {
		t.Fatal(err.Error())
	} else if val != "pair-1-value" {
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
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)

	// Disconnect at end
	defer Disconnect()

	// Create pairs
	pairs := [][2]interface{}{
		{"pair-1", "pair-1-value"},
		{"pair-2", "pair-2-value"},
		{"pair-3", "pair-3-value"},
	}

	// Set the hash map
	_ = HashMapSetExp("example-hash-map-set-exp", pairs, 2*time.Minute, "test-hash-map-1", "test-hash-map-2")
	fmt.Print("set complete")
	//Output: set complete
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
	err := Set("test-get", "my-value")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Get the value
	val, err := Get("test-get")
	if val != "my-value" {
		t.Fatalf("expected value: %s, got: %s", "my-value", val)
	}
}

// ExampleGet is an example of Get() method
func ExampleGet() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-get", "my-value", "another-key")

	// Get the value
	value, _ := Get("example-get")
	fmt.Print(value)
	//Output: my-value
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
	err := Set("test-get-bytes", "my-value")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Get the value
	val, err := GetBytes("test-get-bytes")
	if string(val) != "my-value" {
		t.Fatalf("expected value: %s, got: %s", "my-value", val)
	}
}

// ExampleGetBytes is an example of GetBytes() method
func ExampleGetBytes() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-get-bytes", "my-value", "another-key")

	// Get the value
	value, _ := GetBytes("example-get-bytes")
	fmt.Print(string(value))
	//Output: my-value
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
	err := Set("test-get", "my-value")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Get the value
	keys, err := GetAllKeys()
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
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-get-all-keys", "my-value", "another-key")

	// Get the value
	_, _ = GetAllKeys()
	fmt.Print("found keys")
	//Output: found keys
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
	err := Set("test-exists", "my-value")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Check the set / exists
	exists, err := Exists("test-exists")
	if !exists {
		t.Fatal("expected key to exist")
	}
}

// ExampleExists is an example of Exists() method
func ExampleExists() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-exists", "my-value", "another-key")

	// Get the value
	_, _ = Exists("example-exists")
	fmt.Print("key exists")
	//Output: key exists
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
	err := SetExp("test-set-expire", "my-value", 1*time.Minute, "another-key")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Check the set
	val, err := Get("test-set-expire")
	if val != "my-value" {
		t.Fatalf("expected value: %s, got: %s", "my-value", val)
	}

	// Fire the expire
	err = Expire("test-set-expire", 1*time.Second)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Wait 2 seconds and test
	time.Sleep(time.Second * 2)

	// Check the expire
	val, err = Get("test-set-expire")
	if val == "my-value" {
		t.Fatalf("expected value: %s, got: %s", "", val)
	}
}

// ExampleExpire is an example of Expire() method
func ExampleExpire() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-expire", "my-value", "another-key")

	// Set the expire
	_ = Expire("example-expire", 1*time.Minute)
	fmt.Print("expiration set")
	//Output: expiration set
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
	err := Set("test-destroy", "my-value", "another-key")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Check the set
	val, err := Get("test-destroy")
	if val != "my-value" {
		t.Fatalf("expected value: %s, got: %s", "my-value", val)
	}

	// Fire destroy
	err = DestroyCache()
	if err != nil {
		t.Fatal(err.Error())
	}

	// Check the destroy
	val, err = Get("test-destroy")
	if val == "my-value" {
		t.Fatalf("expected value: %s, got: %s", "", val)
	}
}

// ExampleDestroyCache is an example of DestroyCache() method
func ExampleDestroyCache() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-destroy-cache", "my-value", "another-key")

	// Set the expire
	_ = DestroyCache()
	fmt.Print("cache destroyed")
	//Output: cache destroyed
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

	// Start with a flush
	err := DestroyCache()
	if err != nil {
		t.Fatal(err.Error())
	}

	// Set a key with two dependent keys
	err = Set("test-set-dep", "my-value", "dependent-1", "dependent-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Test for dependent key 1
	ok, err := SetIsMember("depend:dependent-1", "test-set-dep")
	if err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Test for dependent key 2
	ok, err = SetIsMember("depend:dependent-2", "test-set-dep")
	if err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Kill a dependent key
	total, err := Delete("dependent-1")
	if err != nil {
		t.Fatal(err.Error())
	} else if total != 2 {
		t.Fatal("expected 2 keys to be removed", total)
	}

	// Test for main key
	found, err := Exists("test-set-dep")
	if err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}

	// Test for dependency relation
	found, err = Exists("depend:dependent-1")
	if err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}

	// Test for dependent key 2
	ok, err = SetIsMember("depend:dependent-2", "test-set-dep")
	if err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Kill all dependent keys
	total, err = KillByDependency("dependent-1", "dependent-2")
	if err != nil {
		t.Fatal(err.Error())
	} else if total != 1 {
		t.Fatal("expected 1 key to be removed", total)
	}

	// Test for dependency relation
	found, err = Exists("depend:dependent-2")
	if err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}

	// Test for main key
	found, err = Exists("test-set-dep")
	if err != nil {
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
		{"pair-1", "pair-1-value"},
		{"pair-2", "pair-2-value"},
		{"pair-3", "pair-3-value"},
	}

	// Set the hash map
	err := HashMapSet("test-hash-map-dependency", pairs, "test-hash-map-depend-1", "test-hash-map-depend-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	val, err := HashGet("test-hash-map-dependency", "pair-1")
	if err != nil {
		t.Fatal(err.Error())
	} else if val != "pair-1-value" {
		t.Fatal("expected value was wrong")
	}

	// Get a key in the map
	values, err := HashMapGet("test-hash-map-dependency", "pair-1", "pair-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Got two values?
	if len(values) != 2 {
		t.Fatal("expected 2 values", values, len(values))
	}

	// Test for dependent key 1
	ok, err := SetIsMember("depend:test-hash-map-depend-1", "test-hash-map-dependency")
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
	total, err := Delete("test-hash-map-depend-2")
	if err != nil {
		t.Fatal(err.Error())
	} else if total != 2 {
		t.Fatal("expected 2 keys to be removed", total)
	}

	// Test for main key
	found, err := Exists("test-hash-map-dependency")
	if err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}
}
