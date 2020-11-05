package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestWriteLock tests the method WriteLock()
func TestWriteLock(t *testing.T) {

	// todo: mock redis write

	t.Run("write lock error - real redis", func(t *testing.T) {
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

		// Write a lock
		var locked bool
		locked, err = WriteLock(client, "d  `!$-()my-key", "d d d", int64(0))
		assert.Error(t, err)
		assert.Equal(t, false, locked)
	})

	t.Run("write lock - real redis", func(t *testing.T) {
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

		// Write a lock
		var locked bool
		locked, err = WriteLockRaw(conn, "my-key", "the-secret", int64(10))
		assert.NoError(t, err)
		assert.Equal(t, true, locked)
	})

	t.Run("re-lock - real redis", func(t *testing.T) {
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

		// Write a lock
		var locked bool
		locked, err = WriteLockRaw(conn, "my-key", "the-secret", int64(10))
		assert.NoError(t, err)
		assert.Equal(t, true, locked)

		// Attempt to re-lock (should succeed)
		locked, err = WriteLockRaw(conn, "my-key", "the-secret", int64(5))
		assert.NoError(t, err)
		assert.Equal(t, true, locked)
	})

	t.Run("re-lock different secret - real redis", func(t *testing.T) {
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

		// Write a lock
		var locked bool
		locked, err = WriteLockRaw(conn, "my-key", "the-secret", int64(10))
		assert.NoError(t, err)
		assert.Equal(t, true, locked)

		// Attempt to re-lock (should succeed)
		locked, err = WriteLockRaw(conn, "my-key", "different-secret", int64(5))
		assert.Error(t, err)
		assert.Equal(t, false, locked)
	})

	t.Run("lock expired - real redis", func(t *testing.T) {
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

		// Write a lock
		var locked bool
		locked, err = WriteLockRaw(conn, "my-key", "the-secret", int64(1))
		assert.NoError(t, err)
		assert.Equal(t, true, locked)

		time.Sleep(2 * time.Second)

		// Write new lock
		locked, err = WriteLockRaw(conn, "my-key", "new-secret", int64(2))
		assert.NoError(t, err)
		assert.Equal(t, true, locked)
	})
}

// ExampleWriteLock is an example of the method WriteLock()
func ExampleWriteLock() {

	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Write a lock
	_, _ = WriteLock(client, "test-lock", "test-secret", int64(10))

	fmt.Printf("lock created")
	// Output:lock created
}

// TestReleaseLock tests the method ReleaseLock()
func TestReleaseLock(t *testing.T) {

	// todo: mock redis unlock

	t.Run("release lock - real redis", func(t *testing.T) {
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

		// Write a lock
		var locked bool
		locked, err = WriteLockRaw(conn, "my-key", "the-secret", int64(10))
		assert.NoError(t, err)
		assert.Equal(t, true, locked)

		// Release a lock
		locked, err = ReleaseLockRaw(conn, "my-key", "the-secret")
		assert.NoError(t, err)
		assert.Equal(t, true, locked)
	})

	t.Run("release lock twice - real redis", func(t *testing.T) {
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

		// Write a lock
		var locked bool
		locked, err = WriteLockRaw(conn, "my-key", "the-secret", int64(10))
		assert.NoError(t, err)
		assert.Equal(t, true, locked)

		// Release a lock
		locked, err = ReleaseLockRaw(conn, "my-key", "the-secret")
		assert.NoError(t, err)
		assert.Equal(t, true, locked)

		// Release a lock (again)
		locked, err = ReleaseLockRaw(conn, "my-key", "the-secret")
		assert.NoError(t, err)
		assert.Equal(t, true, locked)
	})

	t.Run("release lock - wrong secret - real redis", func(t *testing.T) {
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

		// Write a lock
		var locked bool
		locked, err = WriteLockRaw(conn, "my-key", "the-secret", int64(10))
		assert.NoError(t, err)
		assert.Equal(t, true, locked)

		// Release a lock
		locked, err = ReleaseLockRaw(conn, "my-key", "wrong-secret")
		assert.Error(t, err)
		assert.Equal(t, false, locked)
	})
}

// ExampleReleaseLock is an example of the method ReleaseLock()
func ExampleReleaseLock() {

	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Release a lock
	_, _ = ReleaseLock(client, "test-lock", "test-secret")

	fmt.Printf("lock released")
	// Output:lock released
}
