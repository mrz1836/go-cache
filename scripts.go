package cache

import "github.com/gomodule/redigo/redis"

// RegisterScripts register all scripts
func RegisterScripts(pool *redis.Pool) (err error) {
	conn := GetConnection(pool)
	defer CloseConnection(conn)
	killByDependencySha, err = RegisterScript(conn, killByDependencyLua)
	return
}

// RegisterScript register a script
func RegisterScript(conn redis.Conn, script string) (string, error) {
	return redis.String(conn.Do(scriptCommand, loadCommand, script))
}

// DidRegisterKillByDependencyScript returns true if the script has a sha from redis
func DidRegisterKillByDependencyScript() bool {
	return killByDependencySha != ""
}

// killByDependencySha is the SHA of the script
var killByDependencySha string

// killByDependencyLua is a script for kill related dependencies
var killByDependencyLua = `
--@begin=lua@
redis.replicate_commands()
local number_of_keys = table.getn(ARGV)
local all_keys = {}
for _, key in ipairs(ARGV) do
	table.insert(all_keys, key)
	local set = redis.call("SMEMBERS", key)
	for _, v in ipairs(set) do
	  table.insert(all_keys, v)
	end
end
return redis.call("DEL", unpack(all_keys))
--@end=lua@
`
