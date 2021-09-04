package cache

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

// SetAdd will add the member to the Set and link a reference to each dependency for the entire Set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetAddRaw()
func SetAdd(ctx context.Context, client *Client, setName, member interface{}, dependencies ...string) error {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return err
	}
	defer client.CloseConnection(conn)
	return SetAddRaw(conn, setName, member, dependencies...)
}

// SetAddRaw will add the member to the Set and link a reference to each dependency for the entire Set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/sadd
func SetAddRaw(conn redis.Conn, setName, member interface{}, dependencies ...string) error {
	if _, err := conn.Do(AddToSetCommand, setName, member); err != nil {
		return err
	}

	return linkDependencies(conn, setName, dependencies...)
}

// SetAddMany will add many values to an existing set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetAddManyRaw()
func SetAddMany(ctx context.Context, client *Client, setName string, members ...interface{}) error {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return err
	}
	defer client.CloseConnection(conn)
	return SetAddManyRaw(conn, setName, members...)
}

// SetAddManyRaw will add many values to an existing set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/sadd
func SetAddManyRaw(conn redis.Conn, setName string, members ...interface{}) (err error) {

	// Create the arguments
	args := make([]interface{}, len(members)+1)
	args[0] = setName

	// Loop members
	for i, key := range members {
		args[i+1] = key
	}

	// Fire the delete command
	_, err = conn.Do(AddToSetCommand, args...)
	return
}

// SetIsMember returns if the member is part of the set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetIsMemberRaw()
func SetIsMember(ctx context.Context, client *Client, set, member interface{}) (bool, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return false, err
	}
	defer client.CloseConnection(conn)
	return SetIsMemberRaw(conn, set, member)
}

// SetIsMemberRaw returns if the member is part of the set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/sismember
func SetIsMemberRaw(conn redis.Conn, set, member interface{}) (bool, error) {
	return redis.Bool(conn.Do(IsMemberCommand, set, member))
}

// SetRemoveMember removes the member from the set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetRemoveMemberRaw()
func SetRemoveMember(ctx context.Context, client *Client, set, member interface{}) error {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return err
	}
	defer client.CloseConnection(conn)
	return SetRemoveMemberRaw(conn, set, member)
}

// SetRemoveMemberRaw removes the member from the set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/srem
func SetRemoveMemberRaw(conn redis.Conn, set, member interface{}) (err error) {
	_, err = conn.Do(RemoveMemberCommand, set, member)
	return
}

// SetMembers will fetch all members in the list
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetMembersRaw()
func SetMembers(ctx context.Context, client *Client, set interface{}) ([]string, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer client.CloseConnection(conn)
	return SetMembersRaw(conn, set)
}

// SetMembersRaw will fetch all members in the list
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/smembers
func SetMembersRaw(conn redis.Conn, set interface{}) ([]string, error) {
	return redis.Strings(conn.Do(MembersCommand, set))
}
