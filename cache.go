package cache

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Package constants
const (
	addToSetCommand      string = "SADD"
	deleteCommand        string = "DEL"
	dependencyPrefix     string = "depend:"
	evalCommand          string = "EVALSHA"
	executeCommand       string = "EXEC"
	expireCommand        string = "EXPIRE"
	flushAllCommand      string = "FLUSHALL"
	getCommand           string = "GET"
	hashKeySet           string = "HSET"
	hashMapSet           string = "HMSET"
	isMemberCommand      string = "SISMEMBER"
	multiCommand         string = "MULTI"
	pingCommand          string = "PING"
	removeMemberCommand  string = "SREM"
	setCommand           string = "SET"
	setExpirationCommand string = "SETEX"
)

//GobCacheCreator can be used to generate the content to store. If the content
//is not found, the creator will be invoked and the result will be stored and
//returned
type GobCacheCreator func() ([]byte, error)

// Set will set the key in redis and keep a reference to each dependency
func Set(key string, value interface{}, dependencies ...string) (err error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

	// Fire the set command
	if _, err = conn.Do(setCommand, key, value); err != nil {
		return
	}

	// Link and return the error
	return linkDependencies(conn, key, dependencies...)
}

// SetExp will set the key in redis and keep a reference to each dependency
func SetExp(key string, value interface{}, ttl time.Duration, dependencies ...string) (err error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

	// Fire the set expiration
	_, err = conn.Do(setExpirationCommand, key, int64(ttl.Seconds()), value)
	if err != nil {
		return
	}

	// Link and return the error
	return linkDependencies(conn, key, dependencies...)
}

// HSet will set the hashKey to the value in the specified hashName and link a
// reference to each dependency for the entire hash
func HSet(hashName, hashKey string, value interface{}, dependencies ...string) (err error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

	// Set the hash key
	if _, err = conn.Do(hashKeySet, hashName, hashKey, value); err != nil {
		return
	}

	// Link and return the error
	return linkDependencies(conn, hashName, dependencies...)
}

// HMSet will set the hashKey to the value in the specified hashName and link a
// reference to each dependency for the entire hash
func HMSet(hashName string, pairs [][2]interface{}, dependencies ...string) (err error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

	// Set the arguments
	args := make([]interface{}, 0, 2*len(pairs)+1)
	args = append(args, hashName)
	for _, pair := range pairs {
		args = append(args, pair[0])
		args = append(args, pair[1])
	}

	// Set the hash map
	if _, err = conn.Do(hashMapSet, args...); err != nil {
		return
	}

	// Link and return the error
	return linkDependencies(conn, hashName, dependencies...)
}

// HMSetExp will set the hashKey to the value in the specified hashName and link a
// reference to each dependency for the entire hash
func HMSetExp(hashName string, pairs [][2]interface{}, ttl time.Duration, dependencies ...string) (err error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

	// Set the arguments
	args := make([]interface{}, 0, 2*len(pairs)+1)
	args = append(args, hashName)
	for _, pair := range pairs {
		args = append(args, pair[0])
		args = append(args, pair[1])
	}

	// Set the hash map
	if _, err = conn.Do(hashMapSet, args...); err != nil {
		return
	}

	// Fire the expire command
	if _, err = conn.Do(expireCommand, hashName, ttl.Seconds()); err != nil {
		return
	}

	// Link and return the error
	return linkDependencies(conn, hashName, dependencies...)
}

// SAdd will add the member to the Set and link a reference to each dependency
// for the entire Set
func SAdd(setName, member interface{}, dependencies ...string) (err error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

	// Add member to set
	if _, err = conn.Do(addToSetCommand, setName, member); err != nil {
		return
	}

	// Link and return the error
	return linkDependencies(conn, setName, dependencies...)
}

// SIsMember returns if the member is part of the set
func SIsMember(set, member interface{}) (bool, error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

	// Check if is member
	return redis.Bool(conn.Do(isMemberCommand, set, member))
}

// SRem removes the member from the set
func SRem(set, member interface{}) (err error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

	// Remove and return
	_, err = conn.Do(removeMemberCommand, set, member)
	return
}

// GetOrSetWithExpirationGob will return the cached value for the key or use the
// GobCacheCreator to create and insert the value into the cache. If the expiration
// time is set to a value greater than zero, the key will be set to expire in
// the provided duration
func GetOrSetWithExpirationGob(key string, fn GobCacheCreator, duration time.Duration, dependencies ...string) (data []byte, err error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

	//Get from redis
	data, err = redis.Bytes(conn.Do(getCommand, key))

	//Set the string in redis
	if err != nil {

		//Set the value
		data, err = fn()
		if err != nil {
			return
		}

		//No data?!
		if len(data) == 0 {
			err = fmt.Errorf("value is empty for key: %s", key)
			return
		}

		//Go routine to set the key and expiration
		go func(key string, data []byte, duration time.Duration, dependencies []string) {

			// Create a new connection and defer closing
			conn := GetConnection()
			defer func() {
				_ = conn.Close()
			}()

			//Set an expiration time if found
			if duration > 0 {
				_ = SetExp(key, data, duration, dependencies...) //todo: handle the error?
			} else {
				_ = Set(key, data, dependencies...) //todo: handle the error?
			}
		}(key, data, duration, dependencies)
	}

	//Return the value
	return
}

// DestroyCache will flush the entire redis server. It only removes keys, not
// scripts
func DestroyCache() (err error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

	// Fire the command
	_, err = conn.Do(flushAllCommand)
	return
}

// KillByDependency removes all keys which are listed as depending on the key(s)
func KillByDependency(keys ...string) (total int, err error) {

	// Create a new connection and defer closing
	conn := GetConnection()
	defer func() {
		_ = conn.Close()
	}()

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
	total, err = redis.Int(conn.Do(evalCommand, args...))
	if err != nil {
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

	// Fire the exec command
	_, err = redis.Values(conn.Do(executeCommand))
	return
}
