package cache

import (
	"fmt"
	"testing"

	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
)

// TestSetAdd test the method SetAdd()
func TestSetAdd(t *testing.T) {

	t.Run("set add command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		conn, pool := loadMockRedis()
		assert.NotNil(t, pool)
		defer endTest(pool, conn)

		var tests = []struct {
			testCase     string
			setName      string
			member       interface{}
			dependencies []string
		}{
			{"set with dep", testKey, testStringValue, []string{testDependantKey}},
			{"set multiple strings", testKey, []string{"one", "two", "three"}, []string{testDependantKey}},
			{"set multiple integers", testKey, []int{1, 2, 3}, []string{testDependantKey}},
			{"empty value", testKey, "", []string{testDependantKey}},
			{"no value, no dep", testKey, "", []string{}},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				var commands []*redigomock.Cmd

				// The main command to test
				commands = append(commands, conn.Command(addToSetCommand, test.setName, test.member))

				// Loop for each dependency
				if len(test.dependencies) > 0 {
					commands = append(commands, conn.Command(multiCommand))
					for _, dep := range test.dependencies {
						commands = append(commands, conn.Command(addToSetCommand, dependencyPrefix+dep, test.setName))
					}
					commands = append(commands, conn.Command(executeCommand))

					err := SetAdd(conn, test.setName, test.member, test.dependencies...)
					assert.NoError(t, err)
				} else {
					err := SetAdd(conn, test.setName, test.member, test.dependencies...)
					assert.NoError(t, err)
				}

				for _, c := range commands {
					assert.Equal(t, true, c.Called)
				}
			})
		}
	})

	t.Run("set add command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		conn, pool, err := loadRealRedis()
		assert.NotNil(t, pool)
		assert.NoError(t, err)
		defer endTest(pool, conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = SetAdd(conn, testKey, testStringValue)
		assert.NoError(t, err)

		// Check that the command worked
		var found bool
		found, err = SetIsMember(conn, testKey, testStringValue)
		assert.NoError(t, err)
		assert.Equal(t, true, found)
	})
}

// ExampleSetAdd is an example of the method SetAdd()
func ExampleSetAdd() {
	// Load a mocked redis for testing/examples
	conn, pool := loadMockRedis()

	// Close connections at end of request
	defer CloseAll(pool, conn)

	// Set the key/value
	_ = SetAdd(conn, testKey, testStringValue, testDependantKey)

	// Fire the command
	_, _ = SetIsMember(conn, testKey, testStringValue)
	fmt.Printf("found member: %v", testStringValue)
	// Output:found member: test-string-value
}

// TestSetAddMany test the method SetAddMany()
func TestSetAddMany(t *testing.T) {

	t.Run("set add many command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		conn, pool := loadMockRedis()
		assert.NotNil(t, pool)
		defer endTest(pool, conn)

		var tests = []struct {
			testCase string
			setName  string
			members  []interface{}
		}{
			{"set one", testKey, []interface{}{testStringValue}},
			{"set multiple strings", testKey, []interface{}{"one", "two", "three"}},
			{"empty value", testKey, []interface{}{}},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				var commands []*redigomock.Cmd

				// Create the arguments
				args := make([]interface{}, len(test.members)+1)
				args[0] = test.setName

				// Loop members
				for i, key := range test.members {
					args[i+1] = key
				}

				// The main command to test
				commands = append(commands, conn.Command(addToSetCommand, args...))

				err := SetAddMany(conn, test.setName, test.members...)
				assert.NoError(t, err)

				for _, c := range commands {
					assert.Equal(t, true, c.Called)
				}
			})
		}
	})

	t.Run("set add many command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		conn, pool, err := loadRealRedis()
		assert.NotNil(t, pool)
		assert.NoError(t, err)
		defer endTest(pool, conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = SetAddMany(conn, testKey, testStringValue, testStringValue+"2")
		assert.NoError(t, err)

		// Check that the command worked
		var found bool
		found, err = SetIsMember(conn, testKey, testStringValue+"2")
		assert.NoError(t, err)
		assert.Equal(t, true, found)
	})
}

// ExampleSetAddMany is an example of the method SetAddMany()
func ExampleSetAddMany() {
	// Load a mocked redis for testing/examples
	conn, pool := loadMockRedis()

	// Close connections at end of request
	defer CloseAll(pool, conn)

	// Set the key/value
	_ = SetAddMany(conn, testKey, testStringValue, testStringValue+"2")

	// Fire the command
	_, _ = SetIsMember(conn, testKey, testStringValue+"2")
	fmt.Printf("found member: %v", testStringValue+"2")
	// Output:found member: test-string-value2
}

/*
// TestSetRemoveMember test the SetRemoveMember() method
func TestSetRemoveMember(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	var empty []interface{}
	empty = append(empty, "one", "two", "three")

	err := SetAddMany("test-set", empty...)
	if err != nil {
		t.Fatal("error setting members", err.Error())
	}

	// Try to delete
	err = SetRemoveMember("test-set", "two")
	if err != nil {
		t.Fatal("error SetRemoveMember", err.Error())
	}

	// Test for not finding a value
	var found bool
	found, err = SetIsMember("test-set", "two")
	if err != nil {
		t.Fatal("error in SetIsMember", err.Error())
	} else if found {
		t.Fatal("found a member that should NOT exist")
	}
}


*/
