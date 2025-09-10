package nrredis

import (
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type wrappedConn struct {
	redis.Conn

	txn *newrelic.Transaction
	cfg *Config
}

// wrapConn will wrap a connection with NewRelic support
func wrapConn(c redis.Conn, txn *newrelic.Transaction, cfg *Config) redis.Conn {
	return &wrappedConn{
		Conn: c,
		txn:  txn,
		cfg:  cfg,
	}
}

// Do is a wrapper for the standard method
func (c *wrappedConn) Do(commandName string, args ...interface{}) (interface{}, error) {
	if c.txn != nil {
		seg := c.createSegment(commandName)
		seg.ParameterizedQuery = formatCommand(commandName, args)
		defer seg.End()
	}
	return c.Conn.Do(commandName, args...)
}

// Send is a wrapper for the standard method
func (c *wrappedConn) Send(commandName string, args ...interface{}) error {
	if c.txn != nil {
		seg := c.createSegment(commandName)
		seg.ParameterizedQuery = formatCommand(commandName, args)
		defer seg.End()
	}
	return c.Conn.Send(commandName, args...)
}

// Flush is a wrapper for the standard method
func (c *wrappedConn) Flush() error {
	if c.txn != nil {
		seg := c.createSegment("flush")
		defer seg.End()
	}
	return c.Conn.Flush()
}

// Receive is a wrapper for the standard method
func (c *wrappedConn) Receive() (interface{}, error) {
	if c.txn != nil {
		seg := c.createSegment("receive")
		defer seg.End()
	}
	return c.Conn.Receive()
}

// createSegment will create a new datastore segment for NewRelic
func (c *wrappedConn) createSegment(cmdName string) *newrelic.DatastoreSegment {
	return &newrelic.DatastoreSegment{
		DatabaseName: c.cfg.DBName,
		Host:         c.cfg.Host,
		Operation:    strings.ToLower(cmdName),
		PortPathOrID: c.cfg.PortPathOrID,
		Product:      newrelic.DatastoreRedis,
		StartTime:    c.txn.StartSegmentNow(),
	}
}
