package cache

import (
	"testing"
	"time"
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
	time.Sleep(time.Second * 2)

	val, err = HashGet("test-hash-map-set-expire", "pair-1")
	if err == nil {
		t.Fatal("expected: redigo: nil returned")
	} else if val != "" {
		t.Fatal("expected value to be empty")
	}
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
	time.Sleep(time.Second * 2)

	// Check the set expire
	val, err = Get("test-set-exp")
	if val == "my-value" {
		t.Fatalf("expected value: %s, got: %s", "", val)
	}
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

// TestDependencyManagement tests basic dependency functionality
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
	total, err := KillByDependency("dependent-1")
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

// TestHashDependencyManagement tests HASH dependency functionality
func TestHashDependencyManagement(t *testing.T) {

}
