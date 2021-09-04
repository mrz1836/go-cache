package cache

import (
	"context"
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
// The connection must be closed when you're finished
// Deprecated: use GetConnectionWithContext()
func (c *Client) GetConnection() redis.Conn {
	return c.Pool.Get()
}

// GetConnectionWithContext will return a connection from the pool. (convenience method)
// The connection must be closed when you're finished
func (c *Client) GetConnectionWithContext(ctx context.Context) (redis.Conn, error) {
	return c.Pool.GetContext(ctx)
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
//
// Format of URL: redis://localhost:6379
func Connect(redisURL string, maxActiveConnections, idleConnections int, maxConnLifetime, idleTimeout time.Duration,
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
			IdleTimeout:     idleTimeout,
			MaxActive:       maxActiveConnections,
			MaxConnLifetime: maxConnLifetime,
			MaxIdle:         idleConnections,
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}
				_, doErr := c.Do(PingCommand)
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
//
// Preferred method is "Connect()" to create a pool
// Source: "github.com/soveran/redisurl"
// Format of URL: redis://localhost:6379
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
			if _, err = conn.Do(AuthCommand, password); err != nil {
				conn = nil
				return
			}
		}
	}

	// Fire a select on DB
	if len(redisURL.Path) > 1 {
		_, err = conn.Do(SelectCommand, strings.TrimPrefix(redisURL.Path, "/"))
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
