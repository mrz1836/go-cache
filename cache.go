// Package cache is a simple redis cache dependency system on-top of the famous redigo package
//
// If you have any suggestions or comments, please feel free to open an issue on
// this GitHub repository!
//
// By @MrZ1836
package cache

import (
	"encoding/json"
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
	membersCommand       string = "SMEMBERS"
	multiCommand         string = "MULTI"
	pingCommand          string = "PING"
	removeMemberCommand  string = "SREM"
	scriptCommand        string = "SCRIPT"
	selectCommand        string = "SELECT"
	setCommand           string = "SET"
	setExpirationCommand string = "SETEX"
)

// Get gets a key from redis in string format
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: GetRaw()
func Get(client *Client, key string) (string, error) {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return GetRaw(conn, key)
}

// GetRaw gets a key from redis in string format
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/get
func GetRaw(conn redis.Conn, key string) (string, error) {
	return redis.String(conn.Do(getCommand, key))
}

// GetBytes gets a key from redis formatted in bytes
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: GetBytesRaw()
func GetBytes(client *Client, key string) ([]byte, error) {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return GetBytesRaw(conn, key)
}

// GetBytesRaw gets a key from redis formatted in bytes
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/get
func GetBytesRaw(conn redis.Conn, key string) ([]byte, error) {
	return redis.Bytes(conn.Do(getCommand, key))
}

// GetList returns a []string stored in redis list
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: GetListRaw()
func GetList(client *Client, key string) ([]string, error) {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return GetListRaw(conn, key)
}

// GetListRaw returns a []string stored in redis list
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/lrange
func GetListRaw(conn redis.Conn, key string) (list []string, err error) {

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
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetListRaw()
func SetList(client *Client, key string, slice []string) error {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return SetListRaw(conn, key, slice)
}

// SetListRaw saves a slice as a redis list (appends)
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/rpush
func SetListRaw(conn redis.Conn, key string, slice []string) (err error) {

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
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: GetAllKeysRaw()
func GetAllKeys(client *Client) (keys []string, err error) {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return GetAllKeysRaw(conn)
}

// GetAllKeysRaw returns a []string of keys
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/keys
func GetAllKeysRaw(conn redis.Conn) (keys []string, err error) {
	return redis.Strings(conn.Do(keysCommand, allKeysCommand))
}

// Set will set the key in redis and keep a reference to each dependency
// value can be both a string or []byte
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetRaw()
func Set(client *Client, key string, value interface{}, dependencies ...string) error {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return SetRaw(conn, key, value, dependencies...)
}

// SetRaw will set the key in redis and keep a reference to each dependency
// value can be both a string or []byte
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/set
func SetRaw(conn redis.Conn, key string, value interface{}, dependencies ...string) error {
	if _, err := conn.Do(setCommand, key, value); err != nil {
		return err
	}

	return linkDependencies(conn, key, dependencies...)
}

// SetExp will set the key in redis and keep a reference to each dependency
// value can be both a string or []byte
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetExpRaw()
func SetExp(client *Client, key string, value interface{}, ttl time.Duration, dependencies ...string) error {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return SetExpRaw(conn, key, value, ttl, dependencies...)
}

// SetExpRaw will set the key in redis and keep a reference to each dependency
// value can be both a string or []byte
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/setex
func SetExpRaw(conn redis.Conn, key string, value interface{}, ttl time.Duration, dependencies ...string) error {
	if _, err := conn.Do(setExpirationCommand, key, int64(ttl.Seconds()), value); err != nil {
		return err
	}

	return linkDependencies(conn, key, dependencies...)
}

// Exists checks if a key is present or not
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: ExistsRaw()
func Exists(client *Client, key string) (bool, error) {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return ExistsRaw(conn, key)
}

// ExistsRaw checks if a key is present or not
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/exists
func ExistsRaw(conn redis.Conn, key string) (bool, error) {
	return redis.Bool(conn.Do(existsCommand, key))
}

// Expire sets the expiration for a given key
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: ExpireRaw()
func Expire(client *Client, key string, duration time.Duration) error {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return ExpireRaw(conn, key, duration)
}

// ExpireRaw sets the expiration for a given key
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/expire
func ExpireRaw(conn redis.Conn, key string, duration time.Duration) (err error) {
	_, err = conn.Do(expireCommand, key, int64(duration.Seconds()))
	return
}

// DeleteWithoutDependency will remove keys without using dependency script
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: DeleteWithoutDependencyRaw()
func DeleteWithoutDependency(client *Client, keys ...string) (int, error) {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return DeleteWithoutDependencyRaw(conn, keys...)
}

// DeleteWithoutDependencyRaw will remove keys without using dependency script
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/del
func DeleteWithoutDependencyRaw(conn redis.Conn, keys ...string) (total int, err error) {
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
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: DestroyCacheRaw()
func DestroyCache(client *Client) error {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return DestroyCacheRaw(conn)
}

// DestroyCacheRaw will flush the entire redis server
// It only removes keys, not scripts
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/flushall
func DestroyCacheRaw(conn redis.Conn) (err error) {
	_, err = conn.Do(flushAllCommand)
	return
}

// SetToJSON stores the struct data (Struct->JSON) into redis under a key
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SetToJSONRaw()
func SetToJSON(client *Client, keyName string, modelData interface{},
	ttl time.Duration, dependencies ...string) error {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return SetToJSONRaw(conn, keyName, modelData, ttl, dependencies...)
}

// SetToJSONRaw stores the struct data (Struct->JSON) into redis under a key
// Uses existing connection (does not close connection)
//
// Uses methods: SetExpRaw() or SetRaw()
func SetToJSONRaw(conn redis.Conn, keyName string, modelData interface{},
	ttl time.Duration, dependencies ...string) (err error) {
	var responseBytes []byte
	if responseBytes, err = json.Marshal(&modelData); err != nil {
		return
	}
	if ttl > 0 {
		err = SetExpRaw(conn, keyName, string(responseBytes), ttl, dependencies...)
	} else {
		err = SetRaw(conn, keyName, string(responseBytes), dependencies...)
	}
	return
}
