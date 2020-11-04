package cache

/*
// ExampleDelete is an example of Delete() method
func ExampleDelete() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-destroy-cache", testStringValue, "another-key")

	// Delete keys
	total, _ := Delete("another-key", "another-key-2")
	fmt.Print(total, " deleted keys")
	// Output: 2 deleted keys
}


// TestKillByDependency will test KillByDependency()
func TestKillByDependency(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Test no keys
	_, err := KillByDependency()
	if err != nil {
		t.Errorf("error occurred %s", err.Error())
	}

	// Test  keys
	if _, err = KillByDependency("key1"); err != nil {
		t.Errorf("error occurred %s", err.Error())
	}

	// Test  keys
	if _, err = KillByDependency("key1", "key two; another"); err != nil {
		t.Errorf("error occurred %s", err.Error())
	}
}

// ExampleKillByDependency is an example of KillByDependency() method
func ExampleKillByDependency() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Set the key/value
	_ = Set("example-destroy-cache", testStringValue, "another-key")

	// Delete keys
	_, _ = KillByDependency("another-key", "another-key-2")
	fmt.Print("deleted keys")
	// Output: deleted keys
}

// TestDependencyManagement tests basic dependency functionality
// Tests a myriad of methods
func TestDependencyManagement(t *testing.T) {

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Set a key with two dependent keys
	err := Set("test-set-dep", testStringValue, "dependent-1", "dependent-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Test for dependent key 1
	var ok bool
	if ok, err = SetIsMember("depend:dependent-1", "test-set-dep"); err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Test for dependent key 2
	if ok, err = SetIsMember("depend:dependent-2", "test-set-dep"); err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Kill a dependent key
	var total int
	if total, err = Delete("dependent-1"); err != nil {
		t.Fatal(err.Error())
	} else if total != 2 {
		t.Fatal("expected 2 keys to be removed", total)
	}

	// Test for main key
	var found bool
	if found, err = Exists("test-set-dep"); err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}

	// Test for dependency relation
	if found, err = Exists("depend:dependent-1"); err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}

	// Test for dependent key 2
	if ok, err = SetIsMember("depend:dependent-2", "test-set-dep"); err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Kill all dependent keys
	if total, err = KillByDependency("dependent-1", "dependent-2"); err != nil {
		t.Fatal(err.Error())
	} else if total != 1 {
		t.Fatal("expected 1 key to be removed", total)
	}

	// Test for dependency relation
	if found, err = Exists("depend:dependent-2"); err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}

	// Test for main key
	if found, err = Exists("test-set-dep"); err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}
}

// TestHashMapDependencyManagement tests HASH map dependency functionality
// Tests a myriad of methods
func TestHashMapDependencyManagement(t *testing.T) {

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Create pairs
	pairs := [][2]interface{}{
		{"pair-1", testPairValue},
		{"pair-2", "pair-2-value"},
		{"pair-3", "pair-3-value"},
	}

	// Set the hash map
	err := HashMapSet("test-hash-map-dependency", pairs, "test-hash-map-depend-1", "test-hash-map-depend-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	var val string
	val, err = HashGet("test-hash-map-dependency", "pair-1")
	if err != nil {
		t.Fatal(err.Error())
	} else if val != testPairValue {
		t.Fatal("expected value was wrong")
	}

	// Get a key in the map
	var values []string
	values, err = HashMapGet("test-hash-map-dependency", "pair-1", "pair-2")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Got two values?
	if len(values) != 2 {
		t.Fatal("expected 2 values", values, len(values))
	}

	// Test for dependent key 1
	var ok bool
	ok, err = SetIsMember("depend:test-hash-map-depend-1", "test-hash-map-dependency")
	if err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Test for dependent key 2
	ok, err = SetIsMember("depend:test-hash-map-depend-2", "test-hash-map-dependency")
	if err != nil {
		t.Fatal(err.Error())
	} else if !ok {
		t.Fatal("expected to be true")
	}

	// Kill a dependent key
	var total int
	total, err = Delete("test-hash-map-depend-2")
	if err != nil {
		t.Fatal(err.Error())
	} else if total != 2 {
		t.Fatal("expected 2 keys to be removed", total)
	}

	// Test for main key
	var found bool
	found, err = Exists("test-hash-map-dependency")
	if err != nil {
		t.Fatal(err.Error())
	} else if found {
		t.Fatal("expected found to be false")
	}
}


*/
