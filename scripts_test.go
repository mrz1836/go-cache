package cache

import (
	"testing"
)

// Testing variables
var (
	connectionURL        = "redis://localhost:6379"
	idleTimeout          = 240
	killDependencyHash   = "a648f768f57e73e2497ccaa113d5ad9e731c5cd8"
	maxActiveConnections = 0
	maxConnLifetime      = 0
	maxIdleConnections   = 10
)

// TestRegisterScript tests registering a script
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
