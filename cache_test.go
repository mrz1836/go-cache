package cache

import (
	"testing"
	"time"
)

// TestSet is testing the Set() method
func TestSet(t *testing.T) {
	// Create a local connection
	err := Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer func() {
		Disconnect()
	}()

	// Set
	err = Set("test-set", "my-value", "another-key")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Check the set
	val, err := Get("test-set")
	if val != "my-value" {
		t.Fatalf("expected value: %s, got: %s", "my-value", val)
	}
}

// TestGet is testing the Get() method
func TestGet(t *testing.T) {
	// Create a local connection
	err := Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer func() {
		Disconnect()
	}()

	// Set
	err = Set("test-get", "my-value")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Get the value
	val, err := Get("test-get")
	if val != "my-value" {
		t.Fatalf("expected value: %s, got: %s", "my-value", val)
	}
}

// TestGetBytes is testing the Get() method
func TestGetBytes(t *testing.T) {
	// Create a local connection
	err := Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer func() {
		Disconnect()
	}()

	// Set
	err = Set("test-get-bytes", "my-value")
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
	err := Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer func() {
		Disconnect()
	}()

	// Set
	err = Set("test-get", "my-value")
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
	err := Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer func() {
		Disconnect()
	}()

	// Set
	err = Set("test-exists", "my-value")
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
	err := Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer func() {
		Disconnect()
	}()

	// Set
	err = SetExp("test-set-exp", "my-value", 2*time.Second, "another-key")
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
	err := Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer func() {
		Disconnect()
	}()

	// Set
	err = SetExp("test-set-expire", "my-value", 1*time.Minute, "another-key")
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
	err := Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer func() {
		Disconnect()
	}()

	// Set
	err = Set("test-destroy", "my-value", "another-key")
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
