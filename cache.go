// Package cache is a simple redis cache dependency system on-top of the famous redigo package
//
// If you have any suggestions or comments, please feel free to open an issue on
// this GitHub repository!
//
// By @MrZ1836
package cache

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// Package constants (commands)
const (
	addToSetCommand      string = "SADD"
	allKeysCommand       string = "*"
	authCommand          string = "AUTH"
	deleteCommand        string = "DEL"
	dependencyPrefix     string = "depend:"
	evalCommand          string = "EVALSHA"
	executeCommand       string = "EXEC"
	existsCommand        string = "EXISTS"
	expireCommand        string = "EXPIRE"
	flushAllCommand      string = "FLUSHALL"
	getCommand           string = "GET"
	hashGetCommand       string = "HGET"
	hashKeySetCommand    string = "HSET"
	hashMapGetCommand    string = "HMGET"
	hashMapSetCommand    string = "HMSET"
	isMemberCommand      string = "SISMEMBER"
	keysCommand          string = "KEYS"
	listPushCommand      string = "RPUSH"
	listRangeCommand     string = "LRANGE"
	loadCommand          string = "LOAD"
	multiCommand         string = "MULTI"
	pingCommand          string = "PING"
	removeMemberCommand  string = "SREM"
	scriptCommand        string = "SCRIPT"
	selectCommand        string = "SELECT"
	setCommand           string = "SET"
	setExpirationCommand string = "SETEX"
)

// Get gets a key from redis in string format
//
// Spec: https://redis.io/commands/get
func Get(conn redis.Conn, key string) (string, error) {
	return redis.String(conn.Do(getCommand, key))
}

// GetBytes gets a key from redis formatted in bytes
//
// Spec: https://redis.io/commands/get
func GetBytes(conn redis.Conn, key string) ([]byte, error) {
	return redis.Bytes(conn.Do(getCommand, key))
}

// GetList returns a []string stored in redis list
//
// Spec: https://redis.io/commands/lrange
func GetList(conn redis.Conn, key string) (list []string, err error) {

	// This command takes two parameters specifying the range: 0 start, -1 is the end of the list
	var values []interface{}
	if values, err = redis.Values(conn.Do(listRangeCommand, key, 0, -1)); err != nil {
		return
	}

	// Scan slice by value, return with destination
	err = redis.ScanSlice(values, &list)
	return
}

// SetList saves a slice as a redis list (appends)
//
// Spec: https://redis.io/commands/rpush
func SetList(conn redis.Conn, key string, slice []string) (err error) {

	// Create the arguments
	args := make([]interface{}, len(slice)+1)
	args[0] = key

	// Loop members
	for i, param := range slice {
		args[i+1] = param
	}

	// Fire the set command
	_, err = conn.Do(listPushCommand, args...)
	return
}

// GetAllKeys returns a []string of keys
//
// Spec: https://redis.io/commands/keys
func GetAllKeys(conn redis.Conn) (keys []string, err error) {
	return redis.Strings(conn.Do(keysCommand, allKeysCommand))
}

// Set will set the key in redis and keep a reference to each dependency
// value can be both a string or []byte
//
// Spec: https://redis.io/commands/set
func Set(conn redis.Conn, key string, value interface{}, dependencies ...string) error {
	if _, err := conn.Do(setCommand, key, value); err != nil {
		return err
	}

	return linkDependencies(conn, key, dependencies...)
}

// SetExp will set the key in redis and keep a reference to each dependency
// value can be both a string or []byte
//
// Spec: https://redis.io/commands/setex
func SetExp(conn redis.Conn, key string, value interface{}, ttl time.Duration, dependencies ...string) error {
	if _, err := conn.Do(setExpirationCommand, key, int64(ttl.Seconds()), value); err != nil {
		return err
	}

	return linkDependencies(conn, key, dependencies...)
}

// Exists checks if a key is present or not
//
// Spec: https://redis.io/commands/exists
func Exists(conn redis.Conn, key string) (bool, error) {
	return redis.Bool(conn.Do(existsCommand, key))
}

// Expire sets the expiration for a given key
//
// Spec: https://redis.io/commands/expire
func Expire(conn redis.Conn, key string, duration time.Duration) (err error) {
	_, err = conn.Do(expireCommand, key, int64(duration.Seconds()))
	return
}

// DeleteWithoutDependency will remove keys without using dependency script
//
// Spec: https://redis.io/commands/del
func DeleteWithoutDependency(conn redis.Conn, keys ...string) (total int, err error) {
	for _, key := range keys {
		if _, err = conn.Do(deleteCommand, key); err != nil {
			return
		}
		total++
	}

	return
}

// DestroyCache will flush the entire redis server
// It only removes keys, not scripts
//
// Spec: https://redis.io/commands/flushall
func DestroyCache(conn redis.Conn) (err error) {
	_, err = conn.Do(flushAllCommand)
	return
}
