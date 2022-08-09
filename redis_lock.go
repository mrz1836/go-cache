package cache

import (
	"context"
	"errors"

	"github.com/gomodule/redigo/redis"
)

// ErrLockMismatch is the error if the key is locked by someone else
var ErrLockMismatch = errors.New("key is locked with a different secret")

// lockScript is the locking script
const lockScript = `
local v = redis.call("GET", KEYS[1])
if v == false
then
	return redis.call("SET", KEYS[1], ARGV[1], "NX", "EX", ARGV[2]) and 1
else
	if v == ARGV[1]
	then
		return redis.call("SET", KEYS[1], ARGV[1], "EX", ARGV[2]) and 1
	else
		return 0
	end
end
`

// releaseLockScript is the release lock script (removes lock)
const releaseLockScript = `
local v = redis.call("GET",KEYS[1])
if v == false then
	return 1
elseif v == ARGV[1] then
	return redis.call("DEL",KEYS[1])
else
	return 0
end
`

// WriteLock attempts to grab a redis lock
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: WriteLockRaw()
func WriteLock(ctx context.Context, client *Client, name, secret string, ttl int64) (bool, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return false, err
	}
	defer client.CloseConnection(conn)
	return WriteLockRaw(conn, name, secret, ttl)
}

// WriteLockRaw attempts to grab a redis lock
// Uses existing connection (does not close connection)
func WriteLockRaw(conn redis.Conn, name, secret string, ttl int64) (bool, error) {
	script := redis.NewScript(1, lockScript)
	if resp, err := redis.Int(script.Do(conn, name, secret, ttl)); err != nil {
		return false, err
	} else if resp != 0 {
		return true, nil
	}
	return false, ErrLockMismatch
}

// ReleaseLock releases the redis lock
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: ReleaseLockRaw()
func ReleaseLock(ctx context.Context, client *Client, name, secret string) (bool, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return false, err
	}
	defer client.CloseConnection(conn)
	return ReleaseLockRaw(conn, name, secret)
}

// ReleaseLockRaw releases the redis lock
// Uses existing connection (does not close connection)
func ReleaseLockRaw(conn redis.Conn, name, secret string) (bool, error) {
	script := redis.NewScript(1, releaseLockScript)
	if resp, err := redis.Int(script.Do(conn, name, secret)); err != nil {
		return false, err
	} else if resp != 0 {
		return true, nil
	}
	return false, ErrLockMismatch
}
