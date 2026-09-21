package cache

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIfShort skips live-redis tests when running in -short mode
func skipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping live local redis tests")
	}
}

// TestConnectWithOptions tests the method ConnectWithOptions()
func TestConnectWithOptions(t *testing.T) {
	t.Run("missing url", func(t *testing.T) {
		t.Parallel()

		client, err := ConnectWithOptions(context.Background(), PoolOptions{})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrMissingRedisURL)
		assert.Nil(t, client)
	})

	t.Run("valid url builds a lazy pool (no dial)", func(t *testing.T) {
		t.Parallel()

		// A pool is created lazily; no connection is dialed until borrowed, so
		// this succeeds even without a live redis.
		client, err := ConnectWithOptions(context.Background(), PoolOptions{
			URL:                  testLocalConnectionURL,
			MaxActiveConnections: testMaxActiveConnections,
			IdleConnections:      testMaxIdleConnections,
			MaxConnLifetime:      testMaxConnLifetime,
			IdleTimeout:          testIdleTimeout,
		})
		require.NoError(t, err)
		require.NotNil(t, client)
		assert.NotNil(t, client.Pool)
		assert.Empty(t, client.DependencyScriptSha)
		assert.Empty(t, client.ScriptsLoaded)

		client.Close()
	})

	t.Run("new relic wrap", func(t *testing.T) {
		t.Parallel()

		client, err := ConnectWithOptions(context.Background(), PoolOptions{
			URL:             testLocalConnectionURL,
			NewRelicEnabled: true,
		})
		require.NoError(t, err)
		require.NotNil(t, client)
		assert.NotNil(t, client.Pool)

		client.Close()
	})

	t.Run("new relic wrap with bad url", func(t *testing.T) {
		t.Parallel()

		// extractURL fails on a host without a port when new relic is enabled
		client, err := ConnectWithOptions(context.Background(), PoolOptions{
			URL:             "redis://localhost",
			NewRelicEnabled: true,
		})
		require.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("live connection and ping", func(t *testing.T) {
		skipIfShort(t)

		client, err := ConnectWithOptions(context.Background(), PoolOptions{
			URL:             testLocalConnectionURL,
			IdleConnections: testMaxIdleConnections,
		})
		require.NoError(t, err)
		require.NotNil(t, client)

		conn, err := client.GetConnectionWithContext(context.Background())
		require.NoError(t, err)
		defer client.CloseAll(conn)

		pong, err := redis.String(conn.Do(PingCommand))
		require.NoError(t, err)
		assert.Equal(t, "PONG", pong)
	})

	t.Run("live connection with dependency mode", func(t *testing.T) {
		skipIfShort(t)

		client, err := ConnectWithOptions(context.Background(), PoolOptions{
			URL:            testLocalConnectionURL,
			DependencyMode: true,
		})
		require.NoError(t, err)
		require.NotNil(t, client)
		assert.Equal(t, testKillDependencyHash, client.DependencyScriptSha)
		assert.Len(t, client.ScriptsLoaded, 1)

		client.Close()
	})
}

// TestConnectDelegatesToConnectWithOptions asserts Connect() maps its positional
// parameters onto the same pool configuration as ConnectWithOptions().
func TestConnectDelegatesToConnectWithOptions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	legacy, err := Connect(
		ctx,
		testLocalConnectionURL,
		10, 5,
		60*time.Second, 240*time.Second,
		false, false,
	)
	require.NoError(t, err)
	require.NotNil(t, legacy)
	defer legacy.Close()

	modern, err := ConnectWithOptions(ctx, PoolOptions{
		URL:                  testLocalConnectionURL,
		MaxActiveConnections: 10,
		IdleConnections:      5,
		MaxConnLifetime:      60 * time.Second,
		IdleTimeout:          240 * time.Second,
	})
	require.NoError(t, err)
	require.NotNil(t, modern)
	defer modern.Close()

	// When new relic is disabled the pool is a concrete *redis.Pool
	legacyPool, ok := legacy.Pool.(*redis.Pool)
	require.True(t, ok)
	modernPool, ok := modern.Pool.(*redis.Pool)
	require.True(t, ok)

	assert.Equal(t, modernPool.MaxActive, legacyPool.MaxActive)
	assert.Equal(t, modernPool.MaxIdle, legacyPool.MaxIdle)
	assert.Equal(t, modernPool.MaxConnLifetime, legacyPool.MaxConnLifetime)
	assert.Equal(t, modernPool.IdleTimeout, legacyPool.IdleTimeout)
	assert.Equal(t, modernPool.Wait, legacyPool.Wait)
	assert.Equal(t, 10, legacyPool.MaxActive)
	assert.Equal(t, 5, legacyPool.MaxIdle)
}

// TestDialFromOptions tests the dial function builder directly.
func TestDialFromOptions(t *testing.T) {
	t.Run("bad url returns parse error", func(t *testing.T) {
		t.Parallel()

		conn, err := dialFromOptions(PoolOptions{URL: "redis://user:pass{DEf1=ghi@domain.com"})()
		require.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("plaintext scheme connects", func(t *testing.T) {
		skipIfShort(t)

		conn, err := dialFromOptions(PoolOptions{URL: testLocalConnectionURL})()
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer func() { _ = conn.Close() }()

		pong, err := redis.String(conn.Do(PingCommand))
		require.NoError(t, err)
		assert.Equal(t, "PONG", pong)
	})

	t.Run("rediss scheme enables tls (fails against plaintext server)", func(t *testing.T) {
		skipIfShort(t)

		// Dialing TLS against a plaintext redis fails the handshake, which
		// proves the secure scheme turned TLS on.
		conn, err := dialFromOptions(PoolOptions{
			URL:         "rediss://localhost:6379",
			DialOptions: []redis.DialOption{redis.DialConnectTimeout(2 * time.Second), redis.DialTLSHandshakeTimeout(time.Second)},
		})()
		require.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("valkeys scheme enables tls (fails against plaintext server)", func(t *testing.T) {
		skipIfShort(t)

		conn, err := dialFromOptions(PoolOptions{
			URL:         "valkeys://localhost:6379",
			DialOptions: []redis.DialOption{redis.DialConnectTimeout(2 * time.Second), redis.DialTLSHandshakeTimeout(time.Second)},
		})()
		require.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("tls config enables tls (fails against plaintext server)", func(t *testing.T) {
		skipIfShort(t)

		conn, err := dialFromOptions(PoolOptions{
			URL:         testLocalConnectionURL,
			TLSConfig:   &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // negative test against a plaintext server
			DialOptions: []redis.DialOption{redis.DialConnectTimeout(2 * time.Second), redis.DialTLSHandshakeTimeout(time.Second)},
		})()
		require.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("select database from path", func(t *testing.T) {
		skipIfShort(t)

		conn, err := dialFromOptions(PoolOptions{URL: "redis://localhost:6379/1"})()
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer func() { _ = conn.Close() }()

		pong, err := redis.String(conn.Do(PingCommand))
		require.NoError(t, err)
		assert.Equal(t, "PONG", pong)
	})

	t.Run("bad credentials fail auth", func(t *testing.T) {
		skipIfShort(t)

		// The local test server has no auth configured, so a two-arg ACL AUTH
		// is rejected — proving credentials are sent on dial.
		conn, err := dialFromOptions(PoolOptions{ //nolint:gosec // fake test-only credentials in URL
			URL:         "redis://baduser:badpass@localhost:6379",
			DialOptions: []redis.DialOption{redis.DialConnectTimeout(2 * time.Second), redis.DialTLSHandshakeTimeout(time.Second)},
		})()
		require.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("explicit credentials override url", func(t *testing.T) {
		skipIfShort(t)

		// Explicit bad credentials are applied after (and win over) URL creds,
		// so the dial fails on a no-auth server.
		conn, err := dialFromOptions(PoolOptions{
			URL:         testLocalConnectionURL,
			Username:    "baduser",
			Password:    "badpass",
			DialOptions: []redis.DialOption{redis.DialConnectTimeout(2 * time.Second), redis.DialTLSHandshakeTimeout(time.Second)},
		})()
		require.Error(t, err)
		assert.Nil(t, conn)
	})
}
