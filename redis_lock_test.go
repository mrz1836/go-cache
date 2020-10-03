package cache

import (
	"testing"
	"time"
)

// TestWriteLock will run basic tests for lock/release
func TestWriteLock(t *testing.T) {

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// attempt to lock
	locked, err := WriteLock("my-key", "the-secret", int64(10))
	if err != nil {
		t.Fatalf("error acquiring lock: %q", err.Error())
	}
	if !locked {
		t.Fatal("expected WriteLock to return true")
	}

	// attempt to re-lock (should succeed)
	locked, err = WriteLock("my-key", "the-secret", int64(5))
	if !locked || err != nil {
		t.Fatalf("expected re-lock attempt to succeed, got locked %t error %q", locked, err)
	}

	// attempt to re-lock with different secret (should return error)
	locked, err = WriteLock("my-key", "the-different-secret", int64(5))
	if locked || err != ErrLockMismatch {
		t.Fatalf("expected re-lock attempt to fail, got locked %t error %q", locked, err)
	}

	// attempt to release lock w/ bad secret
	var unlocked bool
	unlocked, err = ReleaseLock("my-key", "the-wrong-secret")
	if unlocked || err != ErrLockMismatch {
		t.Fatalf("expected release lock w/ bad secret to fail, got unlocked %t error %q", unlocked, err)
	}

	// attempt to release lock w/ correct secret
	unlocked, err = ReleaseLock("my-key", "the-secret")
	if !unlocked || err != nil {
		t.Fatalf("expected release lock to succeed, got unlocked %t error %q", unlocked, err)
	}

	// attempt to release lock again (should return true, nil)
	unlocked, err = ReleaseLock("myKey", "the-secret")
	if !unlocked || err != nil {
		t.Fatalf("expected repeat release lock to succeed, got unlocked %t error %q", unlocked, err)
	}
}

// TestReleaseLock will run basic tests for lock/release
func TestReleaseLock(t *testing.T) {

	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// attempt to lock
	locked, err := WriteLock("my-key", "the-secret", int64(5))
	if err != nil {
		t.Fatalf("error acquiring lock: %q", err.Error())
	}
	if !locked {
		t.Fatal("expected WriteLock to return true")
	}

	time.Sleep(50 * time.Millisecond)

	// test if lock is there
	locked, err = WriteLock("my-key", "the-different-secret", int64(5))
	if locked || err != ErrLockMismatch {
		t.Fatalf("expected lock attempt to fail, got locked %t error %q", locked, err)
	}

	time.Sleep(50 * time.Millisecond)

	// test if lock is there
	locked, err = WriteLock("my-key", "the-different-secret", int64(5))
	if locked || err != ErrLockMismatch {
		t.Fatalf("expected lock attempt to fail, got locked %t error %q", locked, err)
	}

	// attempt to release lock w/ correct secret
	var unlocked bool
	unlocked, err = ReleaseLock("my-key", "the-secret")
	if !unlocked || err != nil {
		t.Fatalf("expected release lock to succeed, got unlocked %t error %q", unlocked, err)
	}
}

// TestWriteLockError will run basic error test for WriteLock()
func TestWriteLockError(t *testing.T) {
	// Create a local connection
	if err := startTest(); err != nil {
		t.Fatal(err.Error())
	}

	// Disconnect at end
	defer endTest()

	// Test error case
	_, err := WriteLock("d  `!$-()my-key", "d d d", int64(0))
	if err == nil {
		t.Fatalf("expected error to occur")
	}
}
