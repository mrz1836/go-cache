package cache

import (
	"errors"

	"github.com/gomodule/redigo/redis"
)

// ErrLockMismatch is the error if the key is locked by someone else
var ErrLockMismatch = errors.New("key is locked with a different secret")

// lockScript is the locking script
const lockScript = `
local v = redis.call("GET", KEYS[1])
if v == false or v == ARGV[1]
then
	return redis.call("SET", KEYS[1], ARGV[1], "EX", ARGV[2]) and 1
else
	return 0
end
`

// unlockScript is the unlocking script
const unlockScript = `
local v = redis.call("GET",KEYS[1])
if v == false then
	return 1
elseif v == ARGV[1] then
	return redis.call("DEL",KEYS[1])
else
	return 0
end
`

// WriteLock attempts to grab a redis lock.
func WriteLock(name, secret string, ttl int64) (locked bool, err error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

	script := redis.NewScript(1, lockScript)
	var resp int
	resp, err = redis.Int(script.Do(conn, name, secret, ttl))
	if err != nil {
		return
	} else if resp != 0 {
		locked = true
		return
	}
	err = ErrLockMismatch
	return
}

// ReleaseLock releases the redis lock
func ReleaseLock(name, secret string) (released bool, err error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

	script := redis.NewScript(1, unlockScript)
	var resp int
	resp, err = redis.Int(script.Do(conn, name, secret))
	if err != nil {
		return
	} else if resp != 0 {
		released = true
		return
	}
	err = ErrLockMismatch
	return
}