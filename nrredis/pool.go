package nrredis

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// Pool is an interface for representing a pool of Redis connections
type Pool interface {
	ActiveCount() int
	Close() error
	Get() redis.Conn
	GetContext(ctx context.Context) (redis.Conn, error)
	IdleCount() int
	Stats() redis.PoolStats
}

// Wrap will wrap the existing pool
func Wrap(p Pool, opts ...Option) Pool {
	return &wrappedPool{
		Pool: p,
		cfg:  createConfig(opts),
	}
}

// wrappedPool is a wrapped pool
type wrappedPool struct {
	Pool
	cfg *Config
}

// GetContext will wrap and return a new connection
func (p *wrappedPool) GetContext(ctx context.Context) (conn redis.Conn, err error) {
	if conn, err = p.Pool.GetContext(ctx); err != nil {
		return
	}

	if txn := newrelic.FromContext(ctx); txn != nil {
		conn = wrapConn(conn, txn, p.cfg)
	}

	return
}

// Get will get a connection from the pool
func (p *wrappedPool) Get() redis.Conn {
	return p.Pool.Get()
}

// Close will close the pool
func (p *wrappedPool) Close() error {
	return p.Pool.Close()
}
