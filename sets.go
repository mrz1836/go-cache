package cache

import "github.com/gomodule/redigo/redis"

// SetAdd will add the member to the Set and link a reference to each dependency for the entire Set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetAddRaw()
func SetAdd(client *Client, setName, member interface{}, dependencies ...string) error {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return SetAddRaw(conn, setName, member, dependencies...)
}

// SetAddRaw will add the member to the Set and link a reference to each dependency for the entire Set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/sadd
func SetAddRaw(conn redis.Conn, setName, member interface{}, dependencies ...string) error {
	if _, err := conn.Do(addToSetCommand, setName, member); err != nil {
		return err
	}

	return linkDependencies(conn, setName, dependencies...)
}

// SetAddMany will add many values to an existing set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetAddManyRaw()
func SetAddMany(client *Client, setName string, members ...interface{}) (err error) {
	conn := client.GetConnection()
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

	// Fire the delete
	_, err = conn.Do(addToSetCommand, args...)
	return
}

// SetIsMember returns if the member is part of the set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetIsMemberRaw()
func SetIsMember(client *Client, set, member interface{}) (bool, error) {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return SetIsMemberRaw(conn, set, member)
}

// SetIsMemberRaw returns if the member is part of the set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/sismember
func SetIsMemberRaw(conn redis.Conn, set, member interface{}) (bool, error) {
	return redis.Bool(conn.Do(isMemberCommand, set, member))
}

// SetRemoveMember removes the member from the set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetRemoveMemberRaw()
func SetRemoveMember(client *Client, set, member interface{}) error {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return SetRemoveMemberRaw(conn, set, member)
}

// SetRemoveMemberRaw removes the member from the set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/srem
func SetRemoveMemberRaw(conn redis.Conn, set, member interface{}) (err error) {
	_, err = conn.Do(removeMemberCommand, set, member)
	return
}
