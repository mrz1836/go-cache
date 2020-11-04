package cache

import (
	"errors"

	"github.com/gomodule/redigo/redis"
)

// Delete is an alias for KillByDependency()
func Delete(conn redis.Conn, keys ...string) (total int, err error) {
	return KillByDependency(conn, keys...)
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
