package cache

import (
	"testing"

	"github.com/gomodule/redigo/redis"
)

// TestConnectToURL test the ConnectToURL() method
func TestConnectToURL(t *testing.T) {

	// Connect to url string
	c, err := ConnectToURL(connectionURL)
	if err != nil {
		t.Errorf("Error returned")
	} else if c == nil {
		t.Errorf("Client was nil")
	}

	// Close the connection
	defer func() {
		_ = c.Close()
	}()

	// Try to ping
	pong, err := redis.String(c.Do(pingCommand))
	if err != nil {
		t.Errorf("Call to %s returned an error: %v", pingCommand, err)
	}

	// Got a pong?
	if pong != "PONG" {
		t.Errorf("Wanted PONG, got %v\n", pong)
	}
}

// TestConnect tests the connect method
func TestConnect(t *testing.T) {

	// Create a local connection
	err := Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer func() {
		Disconnect()
	}()

	// Get a connection
	c := GetConnection()

	// Close
	defer func() {
		_ = c.Close()
	}()

	// Test our only script
	if !DidRegisterKillByDependencyScript() {
		t.Fatal("Did not register the script")
	}
}

// TestGetPool test getting a pool
func TestGetPool(t *testing.T) {
	// Create a local connection
	err := Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer func() {
		Disconnect()
	}()

	// Get the pool
	if p := GetPool(); p == nil {
		t.Fatal("expected to get pool")
	}
}

// TestDisconnect test disconnecting the pool
func TestDisconnect(t *testing.T) {
	// Create a local connection
	err := Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect
	Disconnect()

	// Test pool
	if p := GetPool(); p != nil {
		t.Fatal("pool expected to be nil")
	}
}
