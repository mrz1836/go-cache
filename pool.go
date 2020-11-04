package cache

import (
	"errors"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Client is used to store the redis.Pool and additional fields/information
type Client struct {
	DependencyScriptSha string      // Stored SHA of the script after loaded
	Pool                *redis.Pool // Redis pool for the client (get connections)
	ScriptsLoaded       []string    // List of scripts that have been loaded
}

// Close closes the connection pool
func (c *Client) Close() {
	if c.Pool != nil {
		_ = c.Pool.Close()
	}
	c.Pool = nil
}

// CloseAll closes the connection pool and given connection
func (c *Client) CloseAll(conn redis.Conn) redis.Conn {
	c.Close()
	return c.CloseConnection(conn)
}

// GetConnection will return a connection from the pool. (convenience method)
// The connection must be closed when done with use to return it to the pool
func (c *Client) GetConnection() redis.Conn {
	return c.Pool.Get()
}

// CloseConnection will close a previously open connection
func (c *Client) CloseConnection(conn redis.Conn) redis.Conn {
	return CloseConnection(conn)
}

// CloseConnection will close a connection
func CloseConnection(conn redis.Conn) redis.Conn {
	if conn != nil {
		_ = conn.Close()
	}
	return nil
}

// Connect creates a new connection pool connected to the specified url
// URL Format: redis://localhost:6379
func Connect(redisURL string, maxActiveConnections, idleConnections, maxConnLifetime, idleTimeout int,
	dependencyMode bool, options ...redis.DialOption) (client *Client, err error) {

	// Required param for dial
	if len(redisURL) == 0 {
		err = errors.New("missing required parameter: redisURL")
		return
	}

	// Create a new redis client (pool)
	client = &Client{
		Pool: &redis.Pool{
			Dial:            buildDialer(redisURL, options...),
			IdleTimeout:     time.Duration(idleTimeout) * time.Second,
			MaxActive:       maxActiveConnections,
			MaxConnLifetime: time.Duration(maxConnLifetime) * time.Second,
			MaxIdle:         idleConnections,
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}
				_, doErr := c.Do(pingCommand)
				return doErr
			},
		},
		ScriptsLoaded:       nil,
		DependencyScriptSha: "",
	}

	// Cleanup
	cleanUp(client.Pool)

	// Register scripts if enabled
	if dependencyMode {
		err = client.RegisterScripts()
	}

	return
}

// ConnectToURL connects via REDIS_URL and returns a single connection
// Preferred method is Connect() to create a pool
// Source: github.com/soveran/redisurl
// URL Format: redis://localhost:6379
func ConnectToURL(connectToURL string, options ...redis.DialOption) (conn redis.Conn, err error) {

	// Parse the URL
	var redisURL *url.URL
	if redisURL, err = url.Parse(connectToURL); err != nil {
		return
	}

	// Create the connection
	if conn, err = redis.Dial("tcp", redisURL.Host, options...); err != nil {
		return
	}

	// Attempt authentication if needed
	if redisURL.User != nil {
		if password, ok := redisURL.User.Password(); ok {
			if _, err = conn.Do(authCommand, password); err != nil {
				conn = nil
				return
			}
		}
	}

	// Fire a select on DB
	if len(redisURL.Path) > 1 {
		_, err = conn.Do(selectCommand, strings.TrimPrefix(redisURL.Path, "/"))
	}

	return
}

// buildDialer will build a redis connection from URL
func buildDialer(url string, options ...redis.DialOption) func() (redis.Conn, error) {
	return func() (redis.Conn, error) {
		return ConnectToURL(url, options...)
	}
}

// cleanUp is fired after the pool is created
// Source: https://github.com/pete911/examples-redigo
// todo: is this really needed?
func cleanUp(pool *redis.Pool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func(pool *redis.Pool) {
		<-c
		_ = pool.Close()
		os.Exit(0)
	}(pool)
}
