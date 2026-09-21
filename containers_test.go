//go:build docker

// Package cache container tests exercise the client against real Redis and
// Valkey servers started via the local Docker CLI (no third-party container
// libraries, so no extra dependencies are added to the module). They require a
// running Docker daemon and are excluded from normal builds/tests by the
// "docker" build tag.
//
// Run them with:
//
//	go test -tags=docker ./...
package cache

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// redisImage and valkeyImage are the engines exercised for parity. Redis
	// and Valkey are wire-compatible, so the same client works against both.
	redisImage  = "redis:7-alpine"
	valkeyImage = "valkey/valkey:8-alpine"
)

// skipIfNoDocker skips the test when the Docker CLI or daemon is unavailable,
// so environments without Docker are unaffected rather than failing.
func skipIfNoDocker(tb testing.TB) {
	tb.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		tb.Skip("docker CLI not found; skipping container test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(ctx, "docker", "info", "--format", "{{.ServerVersion}}").CombinedOutput(); err != nil {
		tb.Skipf("docker daemon not available; skipping container test: %v: %s", err, out)
	}
}

// startEngineContainer starts a Redis or Valkey container, waits until it is
// ready, and returns its connection URL. The container is force-removed via
// t.Cleanup.
func startEngineContainer(tb testing.TB, image string) string {
	tb.Helper()
	skipIfNoDocker(tb)

	// Publish the container's 6379 to a random loopback host port
	runCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	out, err := exec.CommandContext(runCtx, "docker", "run", "-d", "--rm",
		"-p", "127.0.0.1::6379", image).CombinedOutput()
	require.NoError(tb, err, "docker run failed: %s", out)

	id := strings.TrimSpace(string(out))
	tb.Cleanup(func() {
		_ = exec.Command("docker", "rm", "-f", id).Run()
	})

	// Resolve the mapped host port
	portOut, err := exec.Command("docker", "port", id, "6379/tcp").CombinedOutput()
	require.NoError(tb, err, "docker port failed: %s", portOut)

	hostPort := firstHostPort(string(portOut))
	require.NotEmpty(tb, hostPort, "could not resolve mapped port from %q", portOut)

	address := hostPort
	waitForRedis(tb, address)
	return "redis://" + address
}

// firstHostPort returns the first "host:port" line from `docker port` output.
func firstHostPort(portOutput string) string {
	for _, line := range strings.Split(strings.TrimSpace(portOutput), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}

// waitForRedis polls the server with PING until it responds or the deadline
// passes.
func waitForRedis(tb testing.TB, address string) {
	tb.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := redis.Dial("tcp", address, redis.DialConnectTimeout(2*time.Second))
		if err == nil {
			_, pingErr := conn.Do(PingCommand)
			_ = conn.Close()
			if pingErr == nil {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	tb.Fatalf("redis at %s did not become ready in time", address)
}

// TestContainersEngineParity verifies the core cache and lock behavior — which
// relies on real server-side Lua (the kill-by-dependency script and the lock
// scripts) — against both Redis and Valkey, proving the two engines are
// interchangeable through the same client and URL/config.
func TestContainersEngineParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping container tests")
	}

	tests := []struct {
		name  string
		image string
	}{
		{name: "redis", image: redisImage},
		{name: "valkey", image: valkeyImage},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			url := startEngineContainer(t, tc.image)

			// Dependency mode loads the kill-by-dependency Lua script on connect
			// (real SCRIPT LOAD), so success proves scripting works on the engine.
			client, err := ConnectWithOptions(ctx, PoolOptions{
				URL:             url,
				IdleConnections: 10,
				DependencyMode:  true,
			})
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			assert.Equal(t, testKillDependencyHash, client.DependencyScriptSha)

			// Set + Get round-trip
			const key, value = "parity:key", "parity-value"
			require.NoError(t, Set(ctx, client, key, value))

			got, err := Get(ctx, client, key)
			require.NoError(t, err)
			assert.Equal(t, value, got)

			// Lock round-trip exercises the lock Lua scripts (EVALSHA)
			const lockName, secret = "parity:lock", "secret-123"
			locked, err := WriteLock(ctx, client, lockName, secret, 10)
			require.NoError(t, err)
			assert.True(t, locked)

			released, err := ReleaseLock(ctx, client, lockName, secret)
			require.NoError(t, err)
			assert.True(t, released)
		})
	}
}
