package cache

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// HashSet will set the hashKey to the value in the specified hashName and link a
// reference to each dependency for the entire hash
func HashSet(conn redis.Conn, hashName, hashKey string, value interface{}, dependencies ...string) error {
	if _, err := conn.Do(hashKeySetCommand, hashName, hashKey, value); err != nil {
		return err
	}

	return linkDependencies(conn, hashName, dependencies...)
}

// HashGet gets a key from redis via hash
func HashGet(conn redis.Conn, hash, key string) (string, error) {
	return redis.String(conn.Do(hashGetCommand, hash, key))
}

// HashMapGet gets values from a hash map for corresponding keys
func HashMapGet(conn redis.Conn, hashName string, keys ...interface{}) ([]string, error) {

	// Build up the arguments
	keys = append([]interface{}{hashName}, keys...)

	// Fire the command with all keys
	return redis.Strings(conn.Do(hashMapGetCommand, keys...))
}

// HashMapSet will set the hashKey to the value in the specified hashName and link a
// reference to each dependency for the entire hash
func HashMapSet(conn redis.Conn, hashName string, pairs [][2]interface{}, dependencies ...string) error {

	// Set the arguments
	args := make([]interface{}, 0, 2*len(pairs)+1)
	args = append(args, hashName)
	for _, pair := range pairs {
		args = append(args, pair[0])
		args = append(args, pair[1])
	}

	// Set the hash map
	if _, err := conn.Do(hashMapSetCommand, args...); err != nil {
		return err
	}

	// Link and return the error
	return linkDependencies(conn, hashName, dependencies...)
}

// HashMapSetExp will set the hashKey to the value in the specified hashName and link a
// reference to each dependency for the entire hash
func HashMapSetExp(conn redis.Conn, hashName string, pairs [][2]interface{},
	ttl time.Duration, dependencies ...string) error {

	// Set the arguments
	args := make([]interface{}, 0, 2*len(pairs)+1)
	args = append(args, hashName)
	for _, pair := range pairs {
		args = append(args, pair[0], pair[1])
	}

	// Set the hash map
	if _, err := conn.Do(hashMapSetCommand, args...); err != nil {
		return err
	}

	// Fire the expire command
	if _, err := conn.Do(expireCommand, hashName, int64(ttl.Seconds())); err != nil {
		return err
	}

	// Link and return the error
	return linkDependencies(conn, hashName, dependencies...)
}
