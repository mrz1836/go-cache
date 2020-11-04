package cache

import (
	"errors"

	"github.com/gomodule/redigo/redis"
)

// Delete is an alias for KillByDependency()
func Delete(client *Client, conn redis.Conn, keys ...string) (total int, err error) {
	return KillByDependency(client, conn, keys...)
}

// KillByDependency removes all keys which are listed as depending on the key(s)
// Alias: Delete()
func KillByDependency(client *Client, conn redis.Conn, keys ...string) (total int, err error) {

	// Do we have keys to kill?
	if len(keys) == 0 {
		return
	}

	// Create the arguments
	args := make([]interface{}, len(keys)+2)
	deleteArgs := make([]interface{}, len(keys))

	args[0] = client.DependencyScriptSha
	args[1] = 0

	// Loop keys
	for i, key := range keys {
		args[i+2] = dependencyPrefix + key
		deleteArgs[i] = key
	}

	// Run the script
	if total, err = redis.Int(conn.Do(evalCommand, args...)); err != nil {
		return
	}

	// Fire the delete
	var deleted int
	if deleted, err = redis.Int(conn.Do(deleteCommand, deleteArgs...)); err != nil {
		return
	}
	total += deleted

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
