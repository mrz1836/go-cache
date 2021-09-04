package cache

import (
	"context"
	"errors"

	"github.com/gomodule/redigo/redis"
)

// Delete is an alias for KillByDependency()
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: DeleteRaw()
func Delete(ctx context.Context, client *Client, keys ...string) (total int, err error) {
	var conn redis.Conn
	conn, err = client.GetConnectionWithContext(ctx)
	if err != nil {
		return
	}
	defer client.CloseConnection(conn)
	return DeleteRaw(conn, keys...)
}

// DeleteRaw is an alias for KillByDependency()
// Uses existing connection (does not close connection)
func DeleteRaw(conn redis.Conn, keys ...string) (total int, err error) {
	return KillByDependencyRaw(conn, keys...)
}

// KillByDependency removes all keys which are listed as depending on the key(s)
// Alias: Delete()
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: KillByDependencyRaw()
//
// Commands used:
// https://redis.io/commands/eval
// https://redis.io/commands/del
func KillByDependency(ctx context.Context, client *Client, keys ...string) (int, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return 0, err
	}
	defer client.CloseConnection(conn)
	return KillByDependencyRaw(conn, keys...)
}

// KillByDependencyRaw removes all keys which are listed as depending on the key(s)
// Alias: Delete()
//
// Commands used:
// https://redis.io/commands/eval
// https://redis.io/commands/del
func KillByDependencyRaw(conn redis.Conn, keys ...string) (total int, err error) {

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
		args[i+2] = DependencyPrefix + key
		deleteArgs[i] = key
	}

	// Run the script
	if total, err = redis.Int(conn.Do(EvalCommand, args...)); err != nil {
		return
	}

	// Fire the delete command
	var deleted int
	if deleted, err = redis.Int(conn.Do(DeleteCommand, deleteArgs...)); err != nil {
		return
	}
	total += deleted

	return
}

// linkDependencies links any dependencies
//
// Commands used:
// https://redis.io/commands/multi
// https://redis.io/commands/sadd
// https://redis.io/commands/exec
func linkDependencies(conn redis.Conn, key interface{}, dependencies ...string) (err error) {

	// No dependencies given
	if len(dependencies) == 0 {
		return
	}

	// Send the multi command
	if err = conn.Send(MultiCommand); err != nil {
		return
	}

	// Add all to the set
	for _, dependency := range dependencies {
		if err = conn.Send(AddToSetCommand, DependencyPrefix+dependency, key); err != nil {
			return
		}
	}

	// Fire the exec command (ignore nil error response?)
	if _, err = redis.Values(conn.Do(ExecuteCommand)); errors.Is(err, redis.ErrNil) {
		// todo: test against live redis vs mock (is =nil needed)
		err = nil
	}
	return
}
