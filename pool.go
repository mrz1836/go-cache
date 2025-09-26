package cache

import (
	"context"
	"errors"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/mrz1836/go-cache/nrredis"
)

// Define static errors to avoid dynamic error creation
var (
	ErrRedisPoolNil    = errors.New("redis pool is nil")
	ErrMissingRedisURL = errors.New("missing required parameter: redisURL")
)

// Client is used to store the redis.Pool and additional fields/information
type Client struct {
	DependencyScriptSha string // Stored SHA of the script after loaded
	// Pool                *redis.Pool // Redis pool for the client (get connections)
	Pool          nrredis.Pool // Redis pool for the client (get connections)
	ScriptsLoaded []string     // List of scripts that have been loaded
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
	if c.Pool != nil {
		return c.Pool.GetContext(ctx)
	}
	return nil, ErrRedisPoolNil
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
func Connect(ctx context.Context, redisURL string,
	maxActiveConnections, idleConnections int,
	maxConnLifetime, idleTimeout time.Duration,
	dependencyMode, newRelicEnabled bool, options ...redis.DialOption,
) (client *Client, err error) {
	// Required param for dial
	if len(redisURL) == 0 {
		err = ErrMissingRedisURL
		return nil, err
	}

	// Create the pool
	redisPool := redis.Pool{
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
	}

	// Wrap if NewRelic is enabled
	if newRelicEnabled {
		var host, database, port string
		if host, database, port, err = extractURL(redisURL); err != nil {
			return nil, err
		}

		client = &Client{
			Pool: nrredis.Wrap(
				&redisPool,
				nrredis.WithDBName(database),
				nrredis.WithHost(host),
				nrredis.WithPortPathOrID(port),
			),
			ScriptsLoaded: nil,
		}
	} else {
		client = &Client{
			Pool:          &redisPool,
			ScriptsLoaded: nil,
		}
	}

	// Cleanup
	cleanUp(client.Pool)

	// Register scripts if enabled
	if dependencyMode {
		if err = client.RegisterScripts(ctx); err != nil {
			client.Close()
			return nil, err
		}
	}

	return client, err
}

// ConnectToURL connects via REDIS_URL and returns a single connection
//
// Deprecated: use Connect()
// Preferred method is "Connect()" to create a pool
// Source: "github.com/soveran/redisurl"
// Format of URL: redis://localhost:6379
func ConnectToURL(connectToURL string, options ...redis.DialOption) (conn redis.Conn, err error) {
	// Parse the URL
	var redisURL *url.URL
	if redisURL, err = url.Parse(connectToURL); err != nil {
		return conn, err
	}

	// Create the connection
	if conn, err = redis.Dial("tcp", redisURL.Host, options...); err != nil {
		return conn, err
	}

	// Attempt authentication if needed
	if redisURL.User != nil {
		if password, ok := redisURL.User.Password(); ok {
			if _, err = conn.Do(AuthCommand, password); err != nil {
				conn = nil
				return conn, err
			}
		}
	}

	// Fire a select on DB
	if len(redisURL.Path) > 1 {
		_, err = conn.Do(SelectCommand, strings.TrimPrefix(redisURL.Path, "/"))
	}

	return conn, err
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
func cleanUp(pool nrredis.Pool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL) //nolint:staticcheck // ignore
	go func(pool nrredis.Pool) {
		<-c
		_ = pool.Close()
		os.Exit(0)
	}(pool)
}

// extractURL will extract the parts of the redis url
func extractURL(redisURL string) (host, database, port string, err error) {
	// Parse the URL
	var u *url.URL
	if u, err = url.Parse(redisURL); err != nil {
		return host, database, port, err
	}

	// Split the host and port
	if host, port, err = net.SplitHostPort(u.Host); err != nil {
		return host, database, port, err
	}

	// Set the database
	database = strings.ReplaceAll(u.RequestURI(), "/", "")
	return host, database, port, err
}
