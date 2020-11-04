// Package cache is a simple redis cache dependency system on-top of the famous redigo package
//
// If you have any suggestions or comments, please feel free to open an issue on
// this GitHub repository!
//
// By @MrZ1836
package cache

import (
	"errors"
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

// Get gets a key from redis
func Get(conn redis.Conn, key string) (string, error) {
	return redis.String(conn.Do(getCommand, key))
}

// GetBytes gets a key from redis in bytes
func GetBytes(conn redis.Conn, key string) ([]byte, error) {
	return redis.Bytes(conn.Do(getCommand, key))
}

// GetList returns a []string stored in redis list
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
func GetAllKeys(conn redis.Conn) (keys []string, err error) {
	return redis.Strings(conn.Do(keysCommand, allKeysCommand))
}

// Set will set the key in redis and keep a reference to each dependency
// value can be both a string or []byte
func Set(conn redis.Conn, key string, value interface{}, dependencies ...string) error {
	if _, err := conn.Do(setCommand, key, value); err != nil {
		return err
	}

	return linkDependencies(conn, key, dependencies...)
}

// SetExp will set the key in redis and keep a reference to each dependency
// value can be both a string or []byte
func SetExp(conn redis.Conn, key string, value interface{}, ttl time.Duration, dependencies ...string) error {
	if _, err := conn.Do(setExpirationCommand, key, int64(ttl.Seconds()), value); err != nil {
		return err
	}

	return linkDependencies(conn, key, dependencies...)
}

// Exists checks if a key is present or not
func Exists(conn redis.Conn, key string) (bool, error) {
	return redis.Bool(conn.Do(existsCommand, key))
}

// Expire sets the expiration for a given key
func Expire(conn redis.Conn, key string, duration time.Duration) (err error) {
	_, err = conn.Do(expireCommand, key, int64(duration.Seconds()))
	return
}

// Delete is an alias for KillByDependency()
func Delete(conn redis.Conn, keys ...string) (total int, err error) {
	return KillByDependency(conn, keys...)
}

// DeleteWithoutDependency will remove keys without using dependency script
func DeleteWithoutDependency(conn redis.Conn, keys ...string) (total int, err error) {
	for _, key := range keys {
		if _, err = conn.Do(deleteCommand, key); err != nil {
			return
		}
		total++
	}

	return
}

// SetAdd will add the member to the Set and link a reference to each dependency for the entire Set
func SetAdd(conn redis.Conn, setName, member interface{}, dependencies ...string) error {
	if _, err := conn.Do(addToSetCommand, setName, member); err != nil {
		return err
	}

	return linkDependencies(conn, setName, dependencies...)
}

// SetAddMany will add many values to an existing set
func SetAddMany(conn redis.Conn, setName string, members ...interface{}) (err error) {

	// Create the arguments
	args := make([]interface{}, len(members)+1)
	args[0] = setName

	// Loop members
	for i, key := range members {
		args[i+1] = key
	}

	// Fire the delete
	_, err = conn.Do(addToSetCommand, args...)
	return

	// Link and return the error //todo: add dependencies back?
	// return linkDependencies(conn, setName, dependencies...)
}

// SetIsMember returns if the member is part of the set
func SetIsMember(conn redis.Conn, set, member interface{}) (bool, error) {
	return redis.Bool(conn.Do(isMemberCommand, set, member))
}

// SetRemoveMember removes the member from the set
func SetRemoveMember(conn redis.Conn, set, member interface{}) (err error) {
	_, err = conn.Do(removeMemberCommand, set, member)
	return
}

// DestroyCache will flush the entire redis server
// It only removes keys, not scripts
func DestroyCache(conn redis.Conn) (err error) {
	_, err = conn.Do(flushAllCommand)
	return
}

// KillByDependency removes all keys which are listed as depending on the key(s)
// Also: Delete()
func KillByDependency(conn redis.Conn, keys ...string) (total int, err error) {

	// Do we have keys to kill?
	if len(keys) == 0 {
		return
	}

	// Create the arguments
	args := make([]interface{}, len(keys)+2)
	deleteArgs := make([]interface{}, len(keys))

	args[0] = killByDependencySha
	args[1] = 0

	// Loop keys
	for i, key := range keys {
		args[i+2] = dependencyPrefix + key
		deleteArgs[i] = key
	}

	// Create the script
	if total, err = redis.Int(conn.Do(evalCommand, args...)); err != nil {
		return
	}

	// Fire the delete
	_, err = conn.Do(deleteCommand, deleteArgs...)
	return
}

// linkDependencies links any dependencies
func linkDependencies(conn redis.Conn, key interface{}, dependencies ...string) (err error) {

	// No dependencies given
	if len(dependencies) == 0 {
		return
	}

	// Send the multi command
	if err = conn.Send(multiCommand); err != nil {
		return
	}

	// Add all to the set
	for _, dependency := range dependencies {
		if err = conn.Send(addToSetCommand, dependencyPrefix+dependency, key); err != nil {
			return
		}
	}

	// Fire the exec command (ignore nil error response?) // todo: test against live redis vs mock
	if _, err = redis.Values(conn.Do(executeCommand)); errors.Is(err, redis.ErrNil) {
		err = nil
	}
	return
}
