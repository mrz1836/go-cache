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
