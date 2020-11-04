package cache

import "github.com/gomodule/redigo/redis"

// RegisterScripts will register all required scripts for additional functionality
// This method runs on Connect()
func (c *Client) RegisterScripts() (err error) {

	// Open a new connection
	conn := c.GetConnection()
	defer c.CloseConnection(conn)

	// Load dependency script if not loaded
	if len(c.DependencyScriptSha) == 0 {
		c.DependencyScriptSha, err = RegisterScript(c, conn, killByDependencyLua)
	}
	return
}

// RegisterScript register a new script
func RegisterScript(client *Client, conn redis.Conn, script string) (sha string, err error) {
	if sha, err = redis.String(conn.Do(scriptCommand, loadCommand, script)); err != nil {
		return
	}
	client.ScriptsLoaded = append(client.ScriptsLoaded, sha)
	return
}

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
