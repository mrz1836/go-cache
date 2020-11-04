package cache

/*// TestRegisterScript tests registering a script
func TestRegisterScript(t *testing.T) {

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Register the script
	sha, err := RegisterScript(killByDependencyLua)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Test the returned sha
	if sha != killDependencyHash {
		t.Fatalf("expected: %s, got: %s", killDependencyHash, sha)
	}

	// Is it set
	if !DidRegisterKillByDependencyScript() {
		t.Fatal("Failed to register script")
	}
}

// ExampleRegisterScript is an example of RegisterScript() method
func ExampleRegisterScript() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Fire the method
	sha, _ := RegisterScript(killByDependencyLua)
	fmt.Print(sha)
	// Output: a648f768f57e73e2497ccaa113d5ad9e731c5cd8
}

// BenchmarkRegisterScript benchmarks the RegisterScript() method
func BenchmarkRegisterScript(b *testing.B) {
	_ = startTest()
	for i := 0; i < b.N; i++ {
		_, _ = RegisterScript(killByDependencyLua)
	}
}

// TestRegisterScripts tests registering all scripts
func TestRegisterScripts(t *testing.T) {

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Register the script
	err := RegisterScripts()
	if err != nil {
		t.Fatal(err.Error())
	}

	// Test our only script
	if !DidRegisterKillByDependencyScript() {
		t.Fatal("Did not register the script")
	}
}

// ExampleRegisterScripts is an example of RegisterScripts() method
func ExampleRegisterScripts() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Fire
	_ = RegisterScripts()
	fmt.Print("registered")
	// Output: registered
}

// BenchmarkRegisterScripts benchmarks the RegisterScripts() method
func BenchmarkRegisterScripts(b *testing.B) {
	_ = startTest()
	for i := 0; i < b.N; i++ {
		_ = RegisterScripts()
	}
}

// TestDidRegisterKillByDependencyScript tests the check method
func TestDidRegisterKillByDependencyScript(t *testing.T) {

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Test our only script
	if !DidRegisterKillByDependencyScript() {
		t.Fatal("Did not register the script")
	}
}

// ExampleRegisterScripts is an example of RegisterScripts() method
func ExampleDidRegisterKillByDependencyScript() {
	// Create a local connection
	_ = Connect(connectionURL, maxActiveConnections, maxIdleConnections, maxConnLifetime, idleTimeout, true)

	// Disconnect at end
	defer Disconnect()

	// Fire
	_ = DidRegisterKillByDependencyScript()
	fmt.Print("registered")
	// Output: registered
}

// BenchmarkDidRegisterKillByDependencyScript benchmarks the DidRegisterKillByDependencyScript() method
func BenchmarkDidRegisterKillByDependencyScript(b *testing.B) {
	_ = startTest()
	for i := 0; i < b.N; i++ {
		_ = DidRegisterKillByDependencyScript()
	}
}
*/
