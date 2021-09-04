package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

// TestConnect tests the method Connect()
func TestConnect(t *testing.T) {

	t.Run("valid connection, no dependency mode", func(t *testing.T) {
		t.Parallel()

		client, err := Connect(
			context.Background(),
			testLocalConnectionURL,
			testMaxActiveConnections,
			testMaxIdleConnections,
			testMaxConnLifetime,
			testIdleTimeout,
			false,
		)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Pool)
		assert.Equal(t, "", client.DependencyScriptSha)
		assert.Equal(t, 0, len(client.ScriptsLoaded))

		// Close
		client.Close()
	})

	t.Run("valid connection, with dependency mode", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		client, err := Connect(
			context.Background(),
			testLocalConnectionURL,
			testMaxActiveConnections,
			testMaxIdleConnections,
			testMaxConnLifetime,
			testIdleTimeout,
			true,
		)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Pool)
		assert.Equal(t, testKillDependencyHash, client.DependencyScriptSha)
		assert.Equal(t, 1, len(client.ScriptsLoaded))

		// Close
		client.Close()
	})

	t.Run("valid connection, custom options", func(t *testing.T) {
		t.Parallel()

		client, err := Connect(
			context.Background(),
			testLocalConnectionURL,
			testMaxActiveConnections,
			testMaxIdleConnections,
			testMaxConnLifetime,
			testIdleTimeout,
			false,
			redis.DialKeepAlive(10*time.Second),
		)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Pool)
		assert.Equal(t, "", client.DependencyScriptSha)
		assert.Equal(t, 0, len(client.ScriptsLoaded))

		// Close
		client.Close()
	})

	t.Run("invalid connection", func(t *testing.T) {
		t.Parallel()

		client, err := Connect(
			context.Background(),
			"",
			testMaxActiveConnections,
			testMaxIdleConnections,
			testMaxConnLifetime,
			testIdleTimeout,
			false,
		)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

// ExampleConnect is an example of the method Connect()
func ExampleConnect() {

	client, _ := Connect(
		context.Background(),
		testLocalConnectionURL,
		testMaxActiveConnections,
		testMaxIdleConnections,
		testMaxConnLifetime,
		testIdleTimeout,
		false,
	)

	// Close connections at end of request
	defer client.Close()

	fmt.Printf("connected")
	// Output:connected
}

// TestClient_Close tests the method Close()
func TestClient_Close(t *testing.T) {
	t.Run("close a nil pool", func(t *testing.T) {
		t.Parallel()

		client := new(Client)
		assert.NotNil(t, client)
		client.Close()
		assert.Nil(t, client.Pool)
	})

	t.Run("close an active pool", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)
		client.Close()
		assert.Nil(t, client.Pool)
	})
}

// ExampleClient_Close is an example of the method Close()
func ExampleClient_Close() {

	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	fmt.Printf("closed the pool")
	// Output:closed the pool
}

// TestClient_GetConnection tests the method GetConnection()
func TestClient_GetConnection(t *testing.T) {
	t.Run("get a connection", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		client, err := Connect(
			context.Background(),
			testLocalConnectionURL,
			testMaxActiveConnections,
			testMaxIdleConnections,
			testMaxConnLifetime,
			testIdleTimeout,
			false,
		)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Pool)

		var conn redis.Conn
		conn, err = client.GetConnectionWithContext(context.Background())
		assert.NotNil(t, conn)
		assert.NoError(t, err)

		client.Close()
		assert.Nil(t, client.Pool)
	})
}

// ExampleClient_GetConnection is an example of the method GetConnection()
func ExampleClient_GetConnection() {

	client, _ := Connect(
		context.Background(),
		testLocalConnectionURL,
		testMaxActiveConnections,
		testMaxIdleConnections,
		testMaxConnLifetime,
		testIdleTimeout,
		false,
	)

	conn, _ := client.GetConnectionWithContext(context.Background())
	defer client.CloseAll(conn)
	if conn != nil {
		fmt.Printf("got a connection")
	}
	// Output:got a connection
}

// TestClient_CloseConnection tests the method CloseConnection()
func TestClient_CloseConnection(t *testing.T) {
	t.Run("close a nil connection", func(t *testing.T) {
		t.Parallel()

		client := new(Client)
		conn := *new(redis.Conn)
		conn = client.CloseConnection(conn)
		assert.Nil(t, conn)
	})

	t.Run("close an active connection", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		client, err := Connect(
			context.Background(),
			testLocalConnectionURL,
			testMaxActiveConnections,
			testMaxIdleConnections,
			testMaxConnLifetime,
			testIdleTimeout,
			false,
		)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Pool)

		var conn redis.Conn
		conn, err = client.GetConnectionWithContext(context.Background())
		assert.NotNil(t, conn)
		assert.NoError(t, err)

		conn = client.CloseConnection(conn)
		assert.Nil(t, conn)
	})
}

// ExampleClient_CloseConnection is an example of the method CloseConnection()
func ExampleClient_CloseConnection() {

	// Load a mocked redis for testing/examples
	client, conn := loadMockRedis()

	// Close after finished
	defer client.CloseConnection(conn)

	// Got a connection?
	if conn != nil {
		fmt.Printf("got a connection and closed")
	}
	// Output:got a connection and closed
}

// TestClient_CloseAll tests the method CloseAll()
func TestClient_CloseAll(t *testing.T) {
	t.Run("close a nil connection", func(t *testing.T) {
		t.Parallel()

		client := new(Client)
		assert.NotNil(t, client)
		conn := *new(redis.Conn)
		client.CloseAll(conn)
		assert.Nil(t, conn)
		assert.Nil(t, client.Pool)
	})

	t.Run("close an active connection", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		client, err := Connect(
			context.Background(),
			testLocalConnectionURL,
			testMaxActiveConnections,
			testMaxIdleConnections,
			testMaxConnLifetime,
			testIdleTimeout,
			false,
		)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Pool)

		var conn redis.Conn
		conn, err = client.GetConnectionWithContext(context.Background())
		assert.NotNil(t, conn)
		assert.NoError(t, err)

		conn = client.CloseAll(conn)
		assert.Nil(t, conn)
	})
}

// ExampleClient_CloseAll is an example of the method CloseAll()
func ExampleClient_CloseAll() {

	// Load a mocked redis for testing/examples
	client, conn := loadMockRedis()

	// Close connections at end of request
	defer client.CloseAll(conn)

	// Got a connection?
	if conn != nil {
		fmt.Printf("got a connection and closed")
	}
	// Output:got a connection and closed
}

// TestConnectToURL tests the method ConnectToURL()
func TestConnectToURL(t *testing.T) {
	t.Run("bad url (format)", func(t *testing.T) {
		t.Parallel()

		c, err := ConnectToURL("redis://user:pass{DEf1=ghi@domain.com")
		assert.Error(t, err)
		assert.Nil(t, c)
	})

	t.Run("bad url (file)", func(t *testing.T) {
		t.Parallel()

		c, err := ConnectToURL("foo.html")
		assert.Error(t, err)
		assert.Nil(t, c)
	})

	t.Run("cannot connect (bad host)", func(t *testing.T) {
		t.Parallel()

		c, err := ConnectToURL("redis://doesnotexist.com")
		assert.Error(t, err)
		assert.Nil(t, c)
	})

	t.Run("cannot connect (bad port)", func(t *testing.T) {
		t.Parallel()

		c, err := ConnectToURL("redis://doesnotexist.com:6379", redis.DialConnectTimeout(2*time.Second))
		assert.Error(t, err)
		assert.Nil(t, c)
	})

	t.Run("cannot connect (bad authentication)", func(t *testing.T) {
		t.Parallel()

		c, err := ConnectToURL("redis://user:pass@localhost:6379", redis.DialConnectTimeout(2*time.Second))
		assert.Error(t, err)
		assert.Nil(t, c)
	})

	t.Run("bad path", func(t *testing.T) {
		t.Parallel()

		c, err := ConnectToURL("redis://localhost:6379/pathDb", redis.DialConnectTimeout(2*time.Second))
		assert.Error(t, err)
		assert.NotNil(t, c)
		CloseConnection(c)
	})

	t.Run("valid local connection", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		c, err := ConnectToURL(testLocalConnectionURL)
		assert.NoError(t, err)
		assert.NotNil(t, c)
		defer CloseConnection(c)

		// Try to ping
		var pong string
		pong, err = redis.String(c.Do(PingCommand))
		assert.NoError(t, err)
		assert.Equal(t, "PONG", pong)
	})

	t.Run("valid local connection - dial options", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		c, err := ConnectToURL(testLocalConnectionURL, redis.DialUseTLS(false), redis.DialKeepAlive(3*time.Second))
		assert.NoError(t, err)
		assert.NotNil(t, c)
		defer CloseConnection(c)

		// Try to ping
		var pong string
		pong, err = redis.String(c.Do(PingCommand))
		assert.NoError(t, err)
		assert.Equal(t, "PONG", pong)
	})
}

// ExampleConnectToURL is an example of the method ConnectToURL()
func ExampleConnectToURL() {

	c, _ := ConnectToURL(testLocalConnectionURL)

	// Close connections at end of request
	defer CloseConnection(c)

	fmt.Printf("connected")
	// Output:connected
}
