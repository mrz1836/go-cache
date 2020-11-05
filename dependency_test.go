package cache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDelete tests the method Delete()
func TestDelete(t *testing.T) {

	// todo: mock delete

	t.Run("no keys - real redis", func(t *testing.T) {
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

		// Test no keys
		_, err = Delete(client)
		assert.NoError(t, err)
	})

	t.Run("single key - real redis", func(t *testing.T) {
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

		// Test one key
		var total int
		total, err = Delete(client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, 0, total)
	})

	t.Run("multiple keys - real redis", func(t *testing.T) {
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

		// Test multiple keys
		var total int
		total, err = Delete(client, testKey, "key2", "key3")
		assert.NoError(t, err)
		assert.Equal(t, 0, total)
	})
}

// ExampleDelete is an example of the method Delete()
func ExampleDelete() {

	// Load a mocked redis for testing/examples
	client, conn := loadMockRedis()

	// Close connections at end of request
	defer client.CloseAll(conn)

	// Run command
	_, _ = Delete(client, testDependantKey)
	if conn != nil {
		fmt.Printf("all dependencies deleted")
	}
	// Output:all dependencies deleted
}

// TestKillByDependency tests the method KillByDependency()
func TestKillByDependency(t *testing.T) {

	// todo: mock kill by dependency

	t.Run("no keys - real redis", func(t *testing.T) {
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

		// Test no keys
		_, err = KillByDependency(client)
		assert.NoError(t, err)
	})

	t.Run("single key - real redis", func(t *testing.T) {
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

		// Test one key
		var total int
		total, err = KillByDependency(client, testKey)
		assert.NoError(t, err)
		assert.Equal(t, 0, total)
	})

	t.Run("multiple keys - real redis", func(t *testing.T) {
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

		// Test multiple keys
		var total int
		total, err = KillByDependency(client, testKey, "key2", "key3")
		assert.NoError(t, err)
		assert.Equal(t, 0, total)
	})
}

// ExampleKillByDependency is an example of the method KillByDependency()
func ExampleKillByDependency() {

	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Run command
	_, _ = KillByDependency(client, testDependantKey)
	fmt.Printf("all dependencies removed")
	// Output:all dependencies removed
}

// TestDependencyManagement tests basic dependency functionality
func TestDependencyManagement(t *testing.T) {

	// todo: mock all scenarios

	t.Run("set with dependencies - real redis", func(t *testing.T) {
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

		// Set a key with two dependent keys
		err = SetRaw(conn, "test-set-dep", testStringValue, "dependent-1", "dependent-2")
		assert.NoError(t, err)

		// Test for dependent key 1
		var ok bool
		ok, err = SetIsMemberRaw(conn, "depend:dependent-1", "test-set-dep")
		assert.NoError(t, err)
		assert.Equal(t, true, ok)

		// Test for dependent key 2
		ok, err = SetIsMemberRaw(conn, "depend:dependent-2", "test-set-dep")
		assert.NoError(t, err)
		assert.Equal(t, true, ok)

		// Kill a dependent key
		var total int
		total, err = DeleteRaw(conn, "dependent-1")
		assert.NoError(t, err)
		assert.Equal(t, 2, total)

		// Test for main key
		var found bool
		found, err = ExistsRaw(conn, "test-set-dep")
		assert.NoError(t, err)
		assert.Equal(t, false, found)

		// Test for dependency relation
		found, err = ExistsRaw(conn, "depend:dependent-1")
		assert.NoError(t, err)
		assert.Equal(t, false, found)

		// Test for dependency relation 2
		found, err = SetIsMemberRaw(conn, "depend:dependent-2", "test-set-dep")
		assert.NoError(t, err)
		assert.Equal(t, true, found)

		// Kill all dependent keys
		total, err = KillByDependencyRaw(conn, "dependent-1", "dependent-2")
		assert.NoError(t, err)
		assert.Equal(t, 1, total)

		// Test for dependency relation
		found, err = ExistsRaw(conn, "depend:dependent-2")
		assert.NoError(t, err)
		assert.Equal(t, false, found)

		// Test for main key
		found, err = ExistsRaw(conn, "test-set-dep")
		assert.NoError(t, err)
		assert.Equal(t, false, found)
	})

}

// TestHashMapDependencyManagement tests HASH map dependency functionality
func TestHashMapDependencyManagement(t *testing.T) {

	// todo: mock all scenarios

	t.Run("set with dependencies - real redis", func(t *testing.T) {
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

		// Create pairs
		pairs := [][2]interface{}{
			{"pair-1", "pair-1-value"},
			{"pair-2", "pair-2-value"},
			{"pair-3", "pair-3-value"},
		}

		// Set a key with two dependent keys
		err = HashMapSetRaw(conn, "test-hash-map-dependency", pairs, "test-hash-map-depend-1", "test-hash-map-depend-2")
		assert.NoError(t, err)

		// Test get
		var val string
		val, err = HashGetRaw(conn, "test-hash-map-dependency", "pair-1")
		assert.NoError(t, err)
		assert.Equal(t, "pair-1-value", val)

		// Test get values
		var values []string
		values, err = HashMapGetRaw(conn, "test-hash-map-dependency", "pair-1", "pair-2")
		assert.NoError(t, err)
		assert.Equal(t, 2, len(values))

		// Test for dependent key 1
		var ok bool
		ok, err = SetIsMemberRaw(conn, "depend:test-hash-map-depend-1", "test-hash-map-dependency")
		assert.NoError(t, err)
		assert.Equal(t, true, ok)

		// Test for dependent key 2
		ok, err = SetIsMemberRaw(conn, "depend:test-hash-map-depend-2", "test-hash-map-dependency")
		assert.NoError(t, err)
		assert.Equal(t, true, ok)

		// Kill a dependent key
		var total int
		total, err = DeleteRaw(conn, "test-hash-map-depend-2")
		assert.NoError(t, err)
		assert.Equal(t, 2, total)

		// Test for main key
		var found bool
		found, err = ExistsRaw(conn, "test-hash-map-dependency")
		assert.NoError(t, err)
		assert.Equal(t, false, found)
	})
}
