package cache

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

// StreamEntry represents a single Redis stream entry with an ID and key-value fields
type StreamEntry struct {
	ID     string
	Fields map[string]string
}

// parseStreamEntries parses the nested XREAD response format.
// Redis returns: [[key, [[id, [field, value, field, value, ...]], ...]]]
// We extract all entries across all keys.
func parseStreamEntries(values []interface{}) ([]StreamEntry, error) {
	var entries []StreamEntry
	// outer: list of [key, entries] pairs
	for _, streamRaw := range values {
		streamPair, err := redis.Values(streamRaw, nil)
		if err != nil {
			return nil, err
		}
		if len(streamPair) < 2 {
			continue
		}
		// streamPair[1] = list of [id, fields] pairs
		entryList, err := redis.Values(streamPair[1], nil)
		if err != nil {
			return nil, err
		}
		for _, entryRaw := range entryList {
			entryPair, err := redis.Values(entryRaw, nil)
			if err != nil {
				return nil, err
			}
			if len(entryPair) < 2 {
				continue
			}
			id, err := redis.String(entryPair[0], nil)
			if err != nil {
				return nil, err
			}
			fieldValues, err := redis.Values(entryPair[1], nil)
			if err != nil {
				return nil, err
			}
			fields := make(map[string]string, len(fieldValues)/2)
			for i := 0; i+1 < len(fieldValues); i += 2 {
				k, err := redis.String(fieldValues[i], nil)
				if err != nil {
					return nil, err
				}
				v, err := redis.String(fieldValues[i+1], nil)
				if err != nil {
					return nil, err
				}
				fields[k] = v
			}
			entries = append(entries, StreamEntry{ID: id, Fields: fields})
		}
	}
	return entries, nil
}

// StreamAdd appends an entry with an auto-generated ID to a stream
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: StreamAddRaw()
func StreamAdd(ctx context.Context, client *Client, key string, fields map[string]string) (string, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return "", err
	}
	defer client.CloseConnection(conn)
	return StreamAddRaw(conn, key, fields)
}

// StreamAddRaw appends an entry with an auto-generated ID to a stream
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/xadd
func StreamAddRaw(conn redis.Conn, key string, fields map[string]string) (string, error) {
	args := make([]interface{}, 0, 2+2*len(fields))
	args = append(args, key, "*")
	for k, v := range fields {
		args = append(args, k, v)
	}
	return redis.String(conn.Do(StreamAddCommand, args...))
}

// StreamAddCapped appends an entry to a stream, trimming it to at most maxLen entries
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: StreamAddCappedRaw()
func StreamAddCapped(ctx context.Context, client *Client, key string, maxLen int64, fields map[string]string) (string, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return "", err
	}
	defer client.CloseConnection(conn)
	return StreamAddCappedRaw(conn, key, maxLen, fields)
}

// StreamAddCappedRaw appends an entry to a stream, trimming it to at most maxLen entries
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/xadd
func StreamAddCappedRaw(conn redis.Conn, key string, maxLen int64, fields map[string]string) (string, error) {
	args := make([]interface{}, 0, 4+2*len(fields))
	args = append(args, key, "MAXLEN", "~", maxLen, "*")
	for k, v := range fields {
		args = append(args, k, v)
	}
	return redis.String(conn.Do(StreamAddCommand, args...))
}

// StreamRead reads entries from a stream starting at startID (non-blocking)
// Use "0" for startID to read from the beginning, or "$" for only new entries
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: StreamReadRaw()
func StreamRead(ctx context.Context, client *Client, key, startID string, count int64) ([]StreamEntry, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer client.CloseConnection(conn)
	return StreamReadRaw(conn, key, startID, count)
}

// StreamReadRaw reads entries from a stream starting at startID (non-blocking)
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/xread
func StreamReadRaw(conn redis.Conn, key, startID string, count int64) ([]StreamEntry, error) {
	values, err := redis.Values(conn.Do(StreamReadCommand, "COUNT", count, "STREAMS", key, startID))
	if err != nil {
		return nil, err
	}
	return parseStreamEntries(values)
}

// connWithContext is satisfied by redigo network connections that support context-aware Do.
// Using this avoids spawning a goroutine and eliminates the concurrent Close/Do race on
// the pool's activeConn.state when context cancellation is needed.
type connWithContext interface {
	DoContext(ctx context.Context, commandName string, args ...interface{}) (interface{}, error)
}

// StreamReadBlock reads entries from a stream, blocking until data is available or blockMs elapses
// Respects context cancellation via DoContext when supported, or by closing the connection.
// Use blockMs=0 to block indefinitely.
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: StreamReadBlockRaw()
func StreamReadBlock(ctx context.Context, client *Client, key, startID string, count, blockMs int64) ([]StreamEntry, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return immediately if context is already done — avoids spawning a goroutine
	// that would race with a concurrent Close on the pool's activeConn.
	if ctxErr := ctx.Err(); ctxErr != nil {
		client.CloseConnection(conn)
		return nil, ctxErr
	}

	// Prefer DoContext: it handles cancellation via socket deadline from a single
	// goroutine, avoiding the activeConn.state race entirely.
	// activeConn exposes DoContext but delegates to the underlying conn; if the
	// underlying conn does not implement ConnWithContext (e.g. mock connections),
	// DoContext returns errContextNotSupported — detect and fall through.
	if cwt, ok := conn.(connWithContext); ok {
		values, doErr := redis.Values(cwt.DoContext(ctx, StreamReadCommand, "BLOCK", blockMs, "COUNT", count, "STREAMS", key, startID))
		if doErr == nil || doErr.Error() != "redis: connection does not support ConnWithContext" {
			// DoContext executed (success or a real Redis error) — we are done.
			client.CloseConnection(conn)
			if doErr != nil {
				return nil, doErr
			}
			return parseStreamEntries(values)
		}
		// Fall through to goroutine path — DoContext is not actually supported.
	}

	// Fallback for connections that do not support DoContext (e.g. mock connections):
	// spawn a goroutine and close the connection to unblock it on cancellation.
	type result struct {
		entries []StreamEntry
		err     error
	}
	ch := make(chan result, 1)
	go func() {
		entries, goroutineErr := StreamReadBlockRaw(conn, key, startID, count, blockMs)
		ch <- result{entries, goroutineErr}
	}()

	select {
	case <-ctx.Done():
		_ = conn.Close() // unblocks the goroutine; error intentionally ignored
		<-ch             // wait for goroutine to exit
		client.CloseConnection(conn)
		return nil, ctx.Err()
	case r := <-ch:
		client.CloseConnection(conn)
		return r.entries, r.err
	}
}

// StreamReadBlockRaw reads entries from a stream with blocking support
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/xread
func StreamReadBlockRaw(conn redis.Conn, key, startID string, count, blockMs int64) ([]StreamEntry, error) {
	values, err := redis.Values(conn.Do(StreamReadCommand, "BLOCK", blockMs, "COUNT", count, "STREAMS", key, startID))
	if err != nil {
		return nil, err
	}
	return parseStreamEntries(values)
}

// StreamTrim trims the stream to at most maxLen entries, removing oldest entries first
// Returns the number of entries removed
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: StreamTrimRaw()
func StreamTrim(ctx context.Context, client *Client, key string, maxLen int64) (int64, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return 0, err
	}
	defer client.CloseConnection(conn)
	return StreamTrimRaw(conn, key, maxLen)
}

// StreamTrimRaw trims the stream to at most maxLen entries, removing oldest entries first
// Returns the number of entries removed
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/xtrim
func StreamTrimRaw(conn redis.Conn, key string, maxLen int64) (int64, error) {
	return redis.Int64(conn.Do(StreamTrimCommand, key, "MAXLEN", maxLen))
}

// StreamLen returns the number of entries in a stream
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: StreamLenRaw()
func StreamLen(ctx context.Context, client *Client, key string) (int64, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return 0, err
	}
	defer client.CloseConnection(conn)
	return StreamLenRaw(conn, key)
}

// StreamLenRaw returns the number of entries in a stream
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/xlen
func StreamLenRaw(conn redis.Conn, key string) (int64, error) {
	return redis.Int64(conn.Do(StreamLenCommand, key))
}
