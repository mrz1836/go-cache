package cache

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/mrz1836/go-cache/nrredis"
)

// Define static errors to avoid dynamic error creation
var (
	ErrRedisPoolNil    = errors.New("redis pool is nil")
	ErrMissingRedisURL = errors.New("missing required parameter: redisURL")
)

// Connection defaults and URL schemes
const (
	// defaultRedisPort is used when a URL omits the port
	defaultRedisPort = "6379"

	// secureSchemeRedis and secureSchemeValkey are the TLS-enabling URL schemes.
	// Both Redis and Valkey are wire-compatible; the "s" suffix requests TLS.
	secureSchemeRedis  = "rediss"
	secureSchemeValkey = "valkeys"
)

// Client is used to store the redis.Pool and additional fields/information
type Client struct {
	DependencyScriptSha string       // Stored SHA of the script after loaded
	Pool                nrredis.Pool // Redis pool for the client (get connections)
	ScriptsLoaded       []string     // List of scripts that have been loaded
	mu                  sync.RWMutex // guards Pool and ScriptsLoaded
}

// Close closes the connection pool
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Pool != nil {
		_ = c.Pool.Close()
		c.Pool = nil
	}
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
	c.mu.RLock()
	defer c.mu.RUnlock()
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

// PoolOptions configures a connection pool created by ConnectWithOptions.
//
// URL is required; every other field is optional and its zero value preserves
// the historical behavior of Connect(). The struct is intended to grow
// additively (for example, a future sharded-cluster backend), so callers and
// wrappers can adopt new capabilities without a breaking signature change.
type PoolOptions struct {
	// URL is the connection string.
	//
	// Supported schemes: redis, rediss, valkey, valkeys. The "rediss" and
	// "valkeys" schemes enable TLS automatically. Redis and Valkey are
	// wire-compatible, so both engines use the same URL/config.
	//
	// Format: scheme://[user:password@]host:port[/db]
	URL string

	// Pool sizing (maps 1:1 to the positional parameters of Connect)
	MaxActiveConnections int           // MaxActive (0 = unlimited)
	IdleConnections      int           // MaxIdle
	MaxConnLifetime      time.Duration // 0 = no limit
	IdleTimeout          time.Duration // 0 = no limit

	// Behavior
	DependencyMode  bool // Register the dependency Lua script on connect
	NewRelicEnabled bool // Wrap the pool with New Relic instrumentation

	// TLSConfig, when non-nil, enables TLS and is used for the handshake
	// (custom RootCAs, client certificates, ServerName, InsecureSkipVerify).
	// Leave ServerName empty to let the client default it to the endpoint host
	// — the correct SNI for managed services such as AWS ElastiCache.
	TLSConfig *tls.Config

	// Username and Password provide AUTH credentials. When Username is set a
	// two-arg "AUTH username password" is issued (required by AWS ElastiCache
	// RBAC / Redis ACL users); otherwise a single-arg "AUTH password" is used.
	// These take precedence over any credentials embedded in URL.
	Username string
	Password string

	// DialOptions are extra redigo dial options applied last, so a caller can
	// override any option derived from the fields above.
	DialOptions []redis.DialOption

	// Reserved for a future sharded-cluster backend (additive; unused today):
	//   ClusterMode bool
	//   Addrs       []string
}

// Connect creates a new connection pool connected to the specified url.
//
// Format of URL: redis://localhost:6379
//
// Connect is preserved for backward compatibility and simply forwards to
// ConnectWithOptions. New code should prefer ConnectWithOptions, which also
// supports TLS configuration, ACL/RBAC credentials, and the rediss/valkey(s)
// URL schemes.
func Connect(ctx context.Context, redisURL string,
	maxActiveConnections, idleConnections int,
	maxConnLifetime, idleTimeout time.Duration,
	dependencyMode, newRelicEnabled bool, options ...redis.DialOption,
) (*Client, error) {
	return ConnectWithOptions(ctx, PoolOptions{
		URL:                  redisURL,
		MaxActiveConnections: maxActiveConnections,
		IdleConnections:      idleConnections,
		MaxConnLifetime:      maxConnLifetime,
		IdleTimeout:          idleTimeout,
		DependencyMode:       dependencyMode,
		NewRelicEnabled:      newRelicEnabled,
		DialOptions:          options,
	})
}

// ConnectWithOptions creates a new connection pool from the given options.
//
// It is the preferred entry point for creating a pool. See PoolOptions for the
// supported URL schemes, TLS and authentication behavior.
func ConnectWithOptions(ctx context.Context, opts PoolOptions) (client *Client, err error) {
	// Required param for dial
	if len(opts.URL) == 0 {
		return nil, ErrMissingRedisURL
	}

	// Create the pool
	redisPool := redis.Pool{
		Dial:            dialFromOptions(opts),
		IdleTimeout:     opts.IdleTimeout,
		MaxActive:       opts.MaxActiveConnections,
		MaxConnLifetime: opts.MaxConnLifetime,
		MaxIdle:         opts.IdleConnections,
		Wait:            true,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, doErr := c.Do(PingCommand)
			return doErr
		},
	}

	// Wrap if NewRelic is enabled
	if opts.NewRelicEnabled {
		var host, database, port string
		if host, database, port, err = extractURL(opts.URL); err != nil {
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

	// Register scripts if enabled
	if opts.DependencyMode {
		if err = client.RegisterScripts(ctx); err != nil {
			client.Close()
			return nil, err
		}
	}

	return client, err
}

// dialFromOptions builds the pool's dial function from the given options.
//
// Connections are created through redigo's dial machinery so that URL schemes
// (redis/rediss/valkey/valkeys), TLS, ACL/RBAC authentication and database
// selection are handled consistently. Options are layered so callers always
// win: URL-derived options first, then explicit Username/Password, then
// opts.DialOptions last.
//
// The returned closure intentionally uses redis.Dial (background context)
// rather than the connect-time ctx: the pool creates connections lazily and
// may do so long after ConnectWithOptions returns, so binding a request-scoped
// context here would break connection creation once that context is canceled.
func dialFromOptions(opts PoolOptions) func() (redis.Conn, error) {
	return func() (redis.Conn, error) {
		u, err := url.Parse(opts.URL)
		if err != nil {
			return nil, err
		}

		// Resolve the dial address. Default the port only when a host is
		// present; an empty host is left as-is so it fails as before rather
		// than silently dialing localhost.
		address := u.Host
		if u.Hostname() != "" && u.Port() == "" {
			address = net.JoinHostPort(u.Hostname(), defaultRedisPort)
		}

		var dopts []redis.DialOption

		// TLS: enabled by a secure scheme or by supplying a TLS config. Leaving
		// TLSConfig.ServerName empty lets redigo default it to the endpoint host
		// (correct SNI for managed services like AWS ElastiCache).
		if u.Scheme == secureSchemeRedis || u.Scheme == secureSchemeValkey || opts.TLSConfig != nil {
			dopts = append(dopts, redis.DialUseTLS(true))
			if opts.TLSConfig != nil {
				dopts = append(dopts, redis.DialTLSConfig(opts.TLSConfig))
			}
		}

		// AUTH from URL user-info (mirrors redigo's DialURL semantics):
		//   scheme://user:pass@host -> AUTH user pass (two-arg, ACL/RBAC)
		//   scheme://:pass@host     -> AUTH pass      (single-arg)
		if u.User != nil {
			username := u.User.Username()
			if password, ok := u.User.Password(); ok {
				if username != "" {
					dopts = append(dopts, redis.DialUsername(username))
				}
				dopts = append(dopts, redis.DialPassword(password))
			} else if username != "" {
				// A lone user-info token is treated as the password (redis-cli compatible)
				dopts = append(dopts, redis.DialPassword(username))
			}
		}

		// Explicit credentials take precedence over anything in the URL
		if opts.Username != "" {
			dopts = append(dopts, redis.DialUsername(opts.Username))
		}
		if opts.Password != "" {
			dopts = append(dopts, redis.DialPassword(opts.Password))
		}

		// SELECT database from the URL path (e.g. redis://host/2)
		if db := strings.TrimPrefix(u.Path, "/"); db != "" {
			if n, convErr := strconv.Atoi(db); convErr == nil && n != 0 {
				dopts = append(dopts, redis.DialDatabase(n))
			}
		}

		// Caller-supplied options are applied last so they can override
		dopts = append(dopts, opts.DialOptions...)

		return redis.Dial("tcp", address, dopts...)
	}
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
				_ = conn.Close()
				return nil, err
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
