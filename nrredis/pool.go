package nrredis

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// Pool is an interface for representing a pool of Redis connections
type Pool interface {
	GetContext(ctx context.Context) (redis.Conn, error)
	Get() redis.Conn
	Close() error
	Stats() redis.PoolStats
	ActiveCount() int
	IdleCount() int
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

func (p *wrappedPool) Get() redis.Conn {
	return p.Pool.Get()
}

func (p *wrappedPool) Close() error {
	return p.Pool.Close()
}
