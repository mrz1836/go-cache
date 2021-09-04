package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClient_RegisterScripts tests the method RegisterScripts()
func TestClient_RegisterScripts(t *testing.T) {

	t.Run("valid client - run register", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		client, err := Connect(
			context.Background(),
			testLocalConnectionURL,
			testMaxActiveConnections,
			testMaxIdleConnections,
			testMaxConnLifetime,
			testIdleTimeout,
			false,
		)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Pool)
		assert.Equal(t, "", client.DependencyScriptSha)
		assert.Equal(t, 0, len(client.ScriptsLoaded))
		defer client.Close()

		// Run register
		err = client.RegisterScripts(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, testKillDependencyHash, client.DependencyScriptSha)
		assert.Equal(t, 1, len(client.ScriptsLoaded))
	})

	t.Run("valid client - run register 2 times", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		client, err := Connect(
			context.Background(),
			testLocalConnectionURL,
			testMaxActiveConnections,
			testMaxIdleConnections,
			testMaxConnLifetime,
			testIdleTimeout,
			false,
		)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Pool)
		assert.Equal(t, "", client.DependencyScriptSha)
		assert.Equal(t, 0, len(client.ScriptsLoaded))
		defer client.Close()

		// Run register
		err = client.RegisterScripts(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, testKillDependencyHash, client.DependencyScriptSha)
		assert.Equal(t, 1, len(client.ScriptsLoaded))

		// Run again (should skip)
		err = client.RegisterScripts(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, testKillDependencyHash, client.DependencyScriptSha)
		assert.Equal(t, 1, len(client.ScriptsLoaded))
	})
}

// ExampleClient_RegisterScripts is an example of the method RegisterScripts()
func ExampleClient_RegisterScripts() {

	// Load a mocked redis for testing/examples
	client, conn := loadMockRedis()

	// Close connections at end of request
	defer client.CloseAll(conn)

	// Register known scripts
	_ = client.RegisterScripts(context.Background())

	fmt.Printf("scripts registered")
	// Output:scripts registered
}

// TestRegisterScript is testing the method RegisterScript()
func TestRegisterScript(t *testing.T) {

	t.Run("register script command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase    string
			script      string
			expectedSha string
		}{
			{"valid script", killByDependencyLua, testKillDependencyHash},
			{"another script", `return redis.call("get", KEYS[1])`, "4e6d8fc8bb01276962cce5371fa795a7763657ae"},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				setCmd := conn.Command(ScriptCommand, LoadCommand, test.script).Expect(test.expectedSha)

				val, err := RegisterScriptRaw(client, conn, test.script)
				assert.NoError(t, err)
				assert.Equal(t, true, setCmd.Called)
				assert.Equal(t, test.expectedSha, val)
			})
		}
	})

	t.Run("register script using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		var sha string
		sha, err = RegisterScript(context.Background(), client, killByDependencyLua)
		assert.NoError(t, err)
		assert.Equal(t, testKillDependencyHash, sha)

		// Another script
		sha, err = RegisterScript(context.Background(), client, `return redis.call("get", KEYS[1])`)
		assert.NoError(t, err)
		assert.Equal(t, "a5260dd66ce02462c5b5231c727b3f7772c0bcc5", sha)

		// Another script
		sha, err = RegisterScript(context.Background(), client, lockScript)
		assert.NoError(t, err)
		assert.Equal(t, "e60d96cbb3894dc682fafae2980ad674822f99e1", sha)

		// Another script
		sha, err = RegisterScript(context.Background(), client, releaseLockScript)
		assert.NoError(t, err)
		assert.Equal(t, "3271ffa78c3ca6743c9dc476ff6cae55a9cd3cb4", sha)
	})

	t.Run("register script error", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		var sha string
		sha, err = RegisterScript(context.Background(), client, "invalid script")
		assert.Error(t, err)
		assert.Equal(t, "", sha)
	})
}

// ExampleRegisterScript is an example of the method RegisterScript()
func ExampleRegisterScript() {

	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Register known scripts
	_, _ = RegisterScript(context.Background(), client, killByDependencyLua)

	fmt.Printf("registered: %s", testKillDependencyHash)
	// Output:registered: a648f768f57e73e2497ccaa113d5ad9e731c5cd8
}
