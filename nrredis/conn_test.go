package nrredis

import (
	"errors"
	"testing"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---- Mock Definitions ----

type mockConn struct {
	mock.Mock
}

// Do with execute a command on the Redis server.
func (m *mockConn) Do(commandName string, args ...interface{}) (interface{}, error) {
	callArgs := m.Called(append([]interface{}{commandName}, args...)...)
	return callArgs.Get(0), callArgs.Error(1)
}

// Send sends a command to the Redis server.
func (m *mockConn) Send(commandName string, args ...interface{}) error {
	callArgs := m.Called(append([]interface{}{commandName}, args...)...)
	return callArgs.Error(0)
}

// Flush flushes the commands sent to the Redis server.
func (m *mockConn) Flush() error {
	return m.Called().Error(0)
}

// Receive receives a response from the Redis server.
func (m *mockConn) Receive() (interface{}, error) {
	return m.Called().Get(0), m.Called().Error(1)
}

// Close closes the connection to the Redis server.
func (m *mockConn) Close() error { return nil }

// Err returns any error that occurred on the connection.
func (m *mockConn) Err() error { return nil }

// DoWithTimeout executes a command with a timeout.
func (m *mockConn) DoWithTimeout(timeout int, cmd string, args ...interface{}) (interface{}, error) {
	return nil, nil
}

// SendWithTimeout sends a command with a timeout.
func (m *mockConn) SendWithTimeout(timeout int, cmd string, args ...interface{}) error {
	return nil
}

// ---- Mock Transaction ----

type mockTxn struct{}

// StartSegment creates a new segment for the transaction.
func (m *mockTxn) StartSegmentNow() newrelic.SegmentStartTime {
	return newrelic.SegmentStartTime{}
}

// ---- Test Helpers ----

func testTxn() *newrelic.Transaction {
	return &newrelic.Transaction{}
}

func testCfg() *Config {
	return &Config{
		DBName:       "redis",
		Host:         "localhost",
		PortPathOrID: "6379",
	}
}

// ---- Tests ----

// TestWrappedConn_Do tests the Do method of the wrapped connection.
func TestWrappedConn_Do(t *testing.T) {
	mockRedis := new(mockConn)
	mockRedis.On("Do", "GET", "key").Return("value", nil)

	conn := wrapConn(mockRedis, testTxn(), testCfg())
	result, err := conn.Do("GET", "key")

	require.NoError(t, err)
	require.Equal(t, "value", result)
	mockRedis.AssertExpectations(t)
}

// TestWrappedConn_Send tests the Send method of the wrapped connection.
func TestWrappedConn_Send(t *testing.T) {
	mockRedis := new(mockConn)
	mockRedis.On("Send", "SET", "key", "value").Return(nil)

	conn := wrapConn(mockRedis, testTxn(), testCfg())
	err := conn.Send("SET", "key", "value")

	require.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

// TestWrappedConn_Flush tests the Flush method of the wrapped connection.
func TestWrappedConn_Flush(t *testing.T) {
	mockRedis := new(mockConn)
	mockRedis.On("Flush").Return(nil)

	conn := wrapConn(mockRedis, testTxn(), testCfg())
	err := conn.Flush()

	require.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

// TestWrappedConn_Receive tests the Receive method of the wrapped connection.
func TestWrappedConn_Receive(t *testing.T) {
	mockRedis := new(mockConn)
	mockRedis.On("Receive").Return("PONG", nil)

	conn := wrapConn(mockRedis, testTxn(), testCfg())
	result, err := conn.Receive()

	require.NoError(t, err)
	require.Equal(t, "PONG", result)
	mockRedis.AssertExpectations(t)
}

// TestWrappedConn_Do_WithError tests the Do method of the wrapped connection with an error.
func TestWrappedConn_Do_WithError(t *testing.T) {
	mockRedis := new(mockConn)
	mockRedis.On("Do", "GET", "missing").Return(nil, errors.New("not found"))

	conn := wrapConn(mockRedis, testTxn(), testCfg())
	result, err := conn.Do("GET", "missing")

	require.Error(t, err)
	require.Nil(t, result)
	mockRedis.AssertExpectations(t)
}
