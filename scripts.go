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
		c.DependencyScriptSha, err = RegisterScript(c, killByDependencyLua)
	}
	return
}

// RegisterScript register a new script
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: RegisterScriptRaw()
func RegisterScript(client *Client, script string) (string, error) {
	conn := client.GetConnection()
	defer client.CloseConnection(conn)
	return RegisterScriptRaw(client, conn, script)
}

// RegisterScriptRaw register a new script
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/script-load
func RegisterScriptRaw(client *Client, conn redis.Conn, script string) (sha string, err error) {
	if sha, err = redis.String(conn.Do(ScriptCommand, LoadCommand, script)); err != nil {
		return
	}
	client.ScriptsLoaded = append(client.ScriptsLoaded, sha)
	return
}

// killByDependencySha is the SHA of the below script
const killByDependencySha = "a648f768f57e73e2497ccaa113d5ad9e731c5cd8"

// killByDependencyLua is a script for kill related dependencies
//
// Editing this script requires a new SHA above
var killByDependencyLua = `
--@begin=lua@
redis.replicate_commands()
local number_of_keys = table.getn(ARGV)
local all_keys = {}
for _, key in ipairs(ARGV) do
	table.insert(all_keys, key)
	local set = redis.call("` + MembersCommand + `", key)
	for _, v in ipairs(set) do
	  table.insert(all_keys, v)
	end
end
return redis.call("` + DeleteCommand + `", unpack(all_keys))
--@end=lua@
`
