package cache

import "github.com/gomodule/redigo/redis"

// SetAdd will add the member to the Set and link a reference to each dependency for the entire Set
//
// Spec: https://redis.io/commands/sadd
func SetAdd(conn redis.Conn, setName, member interface{}, dependencies ...string) error {
	if _, err := conn.Do(addToSetCommand, setName, member); err != nil {
		return err
	}

	return linkDependencies(conn, setName, dependencies...)
}

// SetAddMany will add many values to an existing set
//
// Spec: https://redis.io/commands/sadd
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
}

// SetIsMember returns if the member is part of the set
//
// Spec: https://redis.io/commands/sismember
func SetIsMember(conn redis.Conn, set, member interface{}) (bool, error) {
	return redis.Bool(conn.Do(isMemberCommand, set, member))
}

// SetRemoveMember removes the member from the set
//
// Spec: https://redis.io/commands/srem
func SetRemoveMember(conn redis.Conn, set, member interface{}) (err error) {
	_, err = conn.Do(removeMemberCommand, set, member)
	return
}
