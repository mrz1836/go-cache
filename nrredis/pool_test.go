package nrredis

import (
	"context"
	"errors"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Define static test errors
var (
	ErrConnectionError = errors.New("connection error")
	ErrPoolClosed      = errors.New("pool closed")
)

// ---- Mock Definitions ----

type mockPool struct {
	mock.Mock
}

// ActiveCount returns the number of active connections in the pool.
func (m *mockPool) ActiveCount() int {
	args := m.Called()
	return args.Int(0)
}

// Close closes the pool and releases all connections.
func (m *mockPool) Close() error {
	return m.Called().Error(0)
}

// Get returns a connection from the pool.
func (m *mockPool) Get() redis.Conn {
	args := m.Called()
	return args.Get(0).(redis.Conn)
}

// GetContext returns a connection from the pool with the provided context.
func (m *mockPool) GetContext(ctx context.Context) (redis.Conn, error) {
	args := m.Called(ctx)
	return args.Get(0).(redis.Conn), args.Error(1)
}

// IdleCount returns the number of idle connections in the pool.
func (m *mockPool) IdleCount() int {
	args := m.Called()
	return args.Int(0)
}

// Stats returns the statistics of the pool.
func (m *mockPool) Stats() redis.PoolStats {
	args := m.Called()
	return args.Get(0).(redis.PoolStats)
}

// ---- Helper Mock Conn for Pool ----

type dummyConn struct {
	redis.Conn
}

// Close closes the connection.
func (d *dummyConn) Close() error { return nil }

// Err returns any error that occurred on the connection.
func (d *dummyConn) Err() error { return nil }

// Do will execute a command on the Redis server.
func (d *dummyConn) Do(_ string, _ ...interface{}) (interface{}, error) {
	return nil, ErrPoolClosed
}

// Send sends a command to the Redis server.
func (d *dummyConn) Send(_ string, _ ...interface{}) error {
	return nil
}

// Flush flushes the commands sent to the Redis server.
func (d *dummyConn) Flush() error {
	return nil
}

// Receive receives a response from the Redis server.
func (d *dummyConn) Receive() (interface{}, error) {
	return nil, ErrPoolClosed
}

// ---- Tests ----

// TestWrap_ReturnsWrappedPool tests that the Wrap function returns a wrapped pool.
func TestWrap_ReturnsWrappedPool(t *testing.T) {
	mockP := new(mockPool)
	p := Wrap(mockP)

	require.IsType(t, &wrappedPool{}, p)
}

// TestWrappedPool_GetContext_WithTransaction tests the GetContext method of the wrapped pool when a transaction is present.
func TestWrappedPool_GetContext_WithTransaction(t *testing.T) {
	ctx := newrelic.NewContext(context.Background(), &newrelic.Transaction{})
	conn := new(dummyConn)

	mockP := new(mockPool)
	mockP.On("GetContext", ctx).Return(conn, nil)

	p := Wrap(mockP)
	wrapped := p.(*wrappedPool)

	result, err := wrapped.GetContext(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.IsType(t, &wrappedConn{}, result)

	mockP.AssertExpectations(t)
}

// TestWrappedPool_GetContext_WithoutTransaction tests the GetContext method of the wrapped pool when no transaction is present.
func TestWrappedPool_GetContext_WithoutTransaction(t *testing.T) {
	ctx := context.Background()
	conn := new(dummyConn)

	mockP := new(mockPool)
	mockP.On("GetContext", ctx).Return(conn, nil)

	p := Wrap(mockP)
	wrapped := p.(*wrappedPool)

	result, err := wrapped.GetContext(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, conn, result)

	mockP.AssertExpectations(t)
}

// TestWrappedPool_GetContext_Error tests the GetContext method of the wrapped pool when an error occurs.
func TestWrappedPool_GetContext_Error(t *testing.T) {
	ctx := context.Background()
	mockP := new(mockPool)

	// Avoid asserting nil through interface conversion, use Run instead
	mockP.On("GetContext", mock.Anything).Run(func(_ mock.Arguments) {
		// No need to do anything in here; we simulate the call
	}).Return((*dummyConn)(nil), ErrConnectionError) // typed nil with a concrete type

	p := Wrap(mockP)
	wrapped := p.(*wrappedPool)

	result, err := wrapped.GetContext(ctx)
	require.Error(t, err)
	require.Nil(t, result)

	mockP.AssertExpectations(t)
}

// TestWrappedPool_Get tests the Get method of the wrapped pool.
func TestWrappedPool_Get(t *testing.T) {
	conn := new(dummyConn)
	mockP := new(mockPool)
	mockP.On("Get").Return(conn)

	p := Wrap(mockP)
	wrapped := p.(*wrappedPool)

	result := wrapped.Get()
	require.Equal(t, conn, result)

	mockP.AssertExpectations(t)
}

// TestWrappedPool_Close tests the Close method of the wrapped pool.
func TestWrappedPool_Close(t *testing.T) {
	mockP := new(mockPool)
	mockP.On("Close").Return(nil)

	p := Wrap(mockP)
	wrapped := p.(*wrappedPool)

	err := wrapped.Close()
	require.NoError(t, err)

	mockP.AssertExpectations(t)
}
