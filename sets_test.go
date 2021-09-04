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
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

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
				commands = append(commands, conn.Command(AddToSetCommand, test.setName, test.member))

				// Loop for each dependency
				if len(test.dependencies) > 0 {
					commands = append(commands, conn.Command(MultiCommand))
					for _, dep := range test.dependencies {
						commands = append(commands, conn.Command(AddToSetCommand, DependencyPrefix+dep, test.setName))
					}
					commands = append(commands, conn.Command(ExecuteCommand))

					err := SetAddRaw(conn, test.setName, test.member, test.dependencies...)
					assert.NoError(t, err)
				} else {
					err := SetAddRaw(conn, test.setName, test.member, test.dependencies...)
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
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = SetAddRaw(conn, testKey, testStringValue)
		assert.NoError(t, err)

		// Check that the command worked
		var found bool
		found, err = SetIsMemberRaw(conn, testKey, testStringValue)
		assert.NoError(t, err)
		assert.Equal(t, true, found)
	})
}

// ExampleSetAdd is an example of the method SetAdd()
func ExampleSetAdd() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = SetAdd(client, testKey, testStringValue, testDependantKey)

	// Fire the command
	_, _ = SetIsMember(client, testKey, testStringValue)
	fmt.Printf("found member: %v", testStringValue)
	// Output:found member: test-string-value
}

// TestSetAddMany test the method SetAddMany()
func TestSetAddMany(t *testing.T) {

	t.Run("set add many command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

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
				commands = append(commands, conn.Command(AddToSetCommand, args...))

				err := SetAddManyRaw(conn, test.setName, test.members...)
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
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		assert.NoError(t, err)

		// Fire the command
		err = SetAddMany(client, testKey, testStringValue, testStringValue+"2")
		assert.NoError(t, err)

		// Check that the command worked
		var found bool
		found, err = SetIsMember(client, testKey, testStringValue+"2")
		assert.NoError(t, err)
		assert.Equal(t, true, found)
	})
}

// ExampleSetAddMany is an example of the method SetAddMany()
func ExampleSetAddMany() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = SetAddMany(client, testKey, testStringValue, testStringValue+"2")

	// Fire the command
	_, _ = SetIsMember(client, testKey, testStringValue+"2")
	fmt.Printf("found member: %v", testStringValue+"2")
	// Output:found member: test-string-value2
}

// TestSetRemoveMember test the method SetRemoveMember()
func TestSetRemoveMember(t *testing.T) {

	t.Run("set remove member command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase string
			setName  string
			member   interface{}
		}{
			{"set with dep", testKey, testStringValue},
			{"set multiple strings", testKey, testStringValue + "2"},
			{"set multiple integers", testKey, 1},
			{"empty value", testKey, ""},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				var commands []*redigomock.Cmd

				// The main command to test
				commands = append(commands, conn.Command(RemoveMemberCommand, test.setName, test.member))

				err := SetRemoveMember(client, test.setName, test.member)
				assert.NoError(t, err)

				for _, c := range commands {
					assert.Equal(t, true, c.Called)
				}
			})
		}
	})

	t.Run("set remove member command using real redis", func(t *testing.T) {
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
		err = SetAddRaw(conn, testKey, testStringValue)
		assert.NoError(t, err)

		// Check that the command worked
		var found bool
		found, err = SetIsMemberRaw(conn, testKey, testStringValue)
		assert.NoError(t, err)
		assert.Equal(t, true, found)

		// Fire the command
		err = SetRemoveMemberRaw(conn, testKey, testStringValue)
		assert.NoError(t, err)

		// Check that the command worked
		found, err = SetIsMemberRaw(conn, testKey, testStringValue)
		assert.NoError(t, err)
		assert.Equal(t, false, found)
	})
}

// ExampleSetRemoveMember is an example of the method SetRemoveMember()
func ExampleSetRemoveMember() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = SetAddMany(client, testKey, testStringValue, testStringValue+"2")

	// Fire the command
	_ = SetRemoveMember(client, testKey, testStringValue+"2")
	fmt.Printf("removed member: %v", testStringValue+"2")
	// Output:removed member: test-string-value2
}

// TestSetIsMember test the method SetIsMember()
func TestSetIsMember(t *testing.T) {

	t.Run("set is member command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase      string
			setName       interface{}
			member        interface{}
			expectedFound int64
		}{
			{"valid set and member", testKey, testStringValue, 1},
			{"no set name", "", testStringValue, 0},
			{"no member", testKey, "", 0},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				isCmd := conn.Command(IsMemberCommand, test.setName, test.member).Expect(interface{}(test.expectedFound))

				found, err := SetIsMemberRaw(conn, test.setName, test.member)
				assert.NoError(t, err)
				assert.Equal(t, test.expectedFound > 0, found)
				assert.Equal(t, true, isCmd.Called)
			})
		}
	})

	t.Run("set is member command using real redis", func(t *testing.T) {
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
		err = SetAdd(client, testKey, testStringValue)
		assert.NoError(t, err)

		// Check that the command worked
		var found bool
		found, err = SetIsMember(client, testKey, testStringValue)
		assert.NoError(t, err)
		assert.Equal(t, true, found)
	})
}

// ExampleSetIsMember is an example of the method SetIsMember()
func ExampleSetIsMember() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = SetAddMany(client, testKey, testStringValue, testStringValue+"2")

	// Fire the command
	_, _ = SetIsMember(client, testKey, testStringValue+"2")
	fmt.Printf("found member: %v", testStringValue+"2")
	// Output:found member: test-string-value2
}

// TestSetMembers will test the method SetMembers()
func TestSetMembers(t *testing.T) {

	t.Run("get members using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		var tests = []struct {
			testCase      string
			setName       interface{}
			expectedFound []interface{}
		}{
			{"valid set and members", testKey, []interface{}{"one", "two"}},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				cmd := conn.Command(MembersCommand, test.setName).Expect(test.expectedFound)

				found, err := SetMembersRaw(conn, test.setName)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(found))
				assert.Equal(t, "one", found[0])
				assert.Equal(t, "two", found[1])
				assert.Equal(t, true, cmd.Called)
			})
		}
	})

}

// ExampleSetMembers is an example of the method SetMembers()
func ExampleSetMembers() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = SetAddMany(client, testKey, testStringValue, testStringValue)

	// Fire the command
	_, _ = SetMembers(client, testKey)
	fmt.Printf("found members: [%v]", testStringValue)
	// Output:found members: [test-string-value]
}
