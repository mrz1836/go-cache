package cache

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func FuzzKillByDependency(f *testing.F) {
	f.Add("key1,key2,key3")
	f.Add("")
	f.Add("unicode-key-🔑,unicode-key-🚀")
	f.Add("key with spaces,another key with spaces")
	f.Add("key\nwith\nnewlines,key\nwith\nmore\nnewlines")
	f.Add("key\x00\x01\x02,key\xFF\xFE\xFD")
	f.Add(strings.Repeat("long-key-", 100))

	f.Fuzz(func(t *testing.T, keysStr string) {
		if keysStr == "" {
			return
		}

		keys := strings.Split(keysStr, ",")

		client, conn := loadMockRedis()
		defer client.Close()

		args := make([]interface{}, len(keys)+2)
		deleteArgs := make([]interface{}, len(keys))
		args[0] = killByDependencySha
		args[1] = 0

		for i, key := range keys {
			args[i+2] = DependencyPrefix + key
			deleteArgs[i] = key
		}

		conn.Command(EvalCommand, args...).Expect(int64(len(keys)))
		conn.Command(DeleteCommand, deleteArgs...).Expect(int64(len(keys)))

		ctx := context.Background()

		assert.NotPanics(t, func() {
			total, err := KillByDependency(ctx, client, keys...)
			if err == nil {
				assert.GreaterOrEqual(t, total, 0)
				assert.LessOrEqual(t, total, len(keys)*2)
			}
		})
	})
}

func FuzzDelete(f *testing.F) {
	f.Add("key1,key2,key3")
	f.Add("")
	f.Add("unicode-key-🔑")
	f.Add("key with spaces")
	f.Add("key\nwith\nnewlines")
	f.Add("key\x00\x01\x02")

	f.Fuzz(func(t *testing.T, keysStr string) {
		if keysStr == "" {
			return
		}

		keys := strings.Split(keysStr, ",")

		client, conn := loadMockRedis()
		defer client.Close()

		args := make([]interface{}, len(keys)+2)
		deleteArgs := make([]interface{}, len(keys))
		args[0] = killByDependencySha
		args[1] = 0

		for i, key := range keys {
			args[i+2] = DependencyPrefix + key
			deleteArgs[i] = key
		}

		conn.Command(EvalCommand, args...).Expect(int64(len(keys)))
		conn.Command(DeleteCommand, deleteArgs...).Expect(int64(len(keys)))

		ctx := context.Background()

		assert.NotPanics(t, func() {
			total, err := Delete(ctx, client, keys...)
			if err == nil {
				assert.GreaterOrEqual(t, total, 0)
			}
		})
	})
}

func FuzzLinkDependencies(f *testing.F) {
	f.Add("test-key", "dep1,dep2,dep3")
	f.Add("", "")
	f.Add("unicode-key-🔑", "unicode-dep-🎯,unicode-dep-🚀")
	f.Add("key with spaces", "dep with spaces")
	f.Add("key\nwith\nnewlines", "dep\nwith\nnewlines")
	f.Add("key\x00\x01\x02", "dep\x00\x01\x02")

	f.Fuzz(func(t *testing.T, key, dependenciesStr string) {
		if dependenciesStr == "" {
			return
		}

		dependencies := strings.Split(dependenciesStr, ",")

		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(MultiCommand).Expect("QUEUED")
		for _, dependency := range dependencies {
			conn.Command(AddToSetCommand, DependencyPrefix+dependency, key).Expect("QUEUED")
		}
		conn.Command(ExecuteCommand).Expect(make([]interface{}, len(dependencies)))

		assert.NotPanics(t, func() {
			rawConn := conn
			err := linkDependencies(rawConn, key, dependencies...)
			assert.NoError(t, err)
		})
	})
}

func FuzzKillByDependencyWithVariousInputs(f *testing.F) {
	f.Add("single-key")
	f.Add("")
	f.Add("unicode-key-🔑")
	f.Add(strings.Repeat("very-long-key-", 200))
	f.Add("key\nwith\nnewlines")
	f.Add("key with spaces")
	f.Add("key\x00\x01\x02")
	f.Add("key\t\r\n")

	f.Fuzz(func(t *testing.T, key string) {
		if key == "" {
			return
		}

		client, conn := loadMockRedis()
		defer client.Close()

		args := []interface{}{killByDependencySha, 0, DependencyPrefix + key}
		deleteArgs := []interface{}{key}

		conn.Command(EvalCommand, args...).Expect(int64(1))
		conn.Command(DeleteCommand, deleteArgs...).Expect(int64(1))

		ctx := context.Background()

		assert.NotPanics(t, func() {
			total, err := KillByDependency(ctx, client, key)
			if err == nil {
				assert.GreaterOrEqual(t, total, 0)
			}
		})
	})
}

func FuzzKillByDependencyRaw(f *testing.F) {
	f.Add("raw-key1,raw-key2")
	f.Add("unicode-raw-🔑")
	f.Add("")
	f.Add("raw key with spaces")

	f.Fuzz(func(t *testing.T, keysStr string) {
		if keysStr == "" {
			return
		}

		keys := strings.Split(keysStr, ",")

		client, conn := loadMockRedis()
		defer client.Close()

		args := make([]interface{}, len(keys)+2)
		deleteArgs := make([]interface{}, len(keys))
		args[0] = killByDependencySha
		args[1] = 0

		for i, key := range keys {
			args[i+2] = DependencyPrefix + key
			deleteArgs[i] = key
		}

		conn.Command(EvalCommand, args...).Expect(int64(len(keys)))
		conn.Command(DeleteCommand, deleteArgs...).Expect(int64(len(keys)))

		assert.NotPanics(t, func() {
			rawConn := conn
			total, err := KillByDependencyRaw(rawConn, keys...)
			if err == nil {
				assert.GreaterOrEqual(t, total, 0)
			}
		})
	})
}

func FuzzDependencyOperationsWithEdgeCases(f *testing.F) {
	f.Add("normal-key", "normal-dep")
	f.Add("", "")
	f.Add("unicode-key-🔑", "unicode-dep-🎯")
	f.Add("key with spaces", "dep with spaces")
	f.Add("key\nwith\nnewlines", "dep\nwith\nnewlines")
	f.Add("key\x00\x01\x02", "dep\x00\x01\x02")
	f.Add("key\t\r\n", "dep\t\r\n")
	f.Add(strings.Repeat("long-key-", 100), strings.Repeat("long-dep-", 100))

	f.Fuzz(func(t *testing.T, key, dependency string) {
		if key == "" || dependency == "" {
			return
		}

		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(MultiCommand).Expect("QUEUED")
		conn.Command(AddToSetCommand, DependencyPrefix+dependency, key).Expect("QUEUED")
		conn.Command(ExecuteCommand).Expect([]interface{}{int64(1)})

		args := []interface{}{killByDependencySha, 0, DependencyPrefix + dependency}
		deleteArgs := []interface{}{dependency}

		conn.Command(EvalCommand, args...).Expect(int64(1))
		conn.Command(DeleteCommand, deleteArgs...).Expect(int64(1))

		ctx := context.Background()

		assert.NotPanics(t, func() {
			rawConn := conn
			err := linkDependencies(rawConn, key, dependency)
			assert.NoError(t, err)

			total, err := KillByDependency(ctx, client, dependency)
			if err == nil {
				assert.GreaterOrEqual(t, total, 0)
			}
		})
	})
}

func FuzzDependencyPrefixHandling(f *testing.F) {
	f.Add("test-key")
	f.Add("unicode-key-🔑")
	f.Add("key with spaces")
	f.Add("key\nwith\nnewlines")
	f.Add(strings.Repeat("long-key-", 50))

	f.Fuzz(func(t *testing.T, key string) {
		if key == "" {
			return
		}

		prefixedKey := DependencyPrefix + key

		assert.NotPanics(t, func() {
			assert.True(t, strings.HasPrefix(prefixedKey, DependencyPrefix))
			assert.Contains(t, prefixedKey, key)
		})
	})
}

func FuzzEmptyKeyHandling(f *testing.F) {
	f.Add(true)

	f.Fuzz(func(t *testing.T, _ bool) {
		client, conn := loadMockRedis()
		defer client.Close()

		ctx := context.Background()

		assert.NotPanics(t, func() {
			total, err := KillByDependency(ctx, client)
			assert.Equal(t, 0, total)
			assert.NoError(t, err)
		})

		assert.NotPanics(t, func() {
			total, err := Delete(ctx, client)
			assert.Equal(t, 0, total)
			assert.NoError(t, err)
		})

		assert.NotPanics(t, func() {
			rawConn := conn
			total, err := KillByDependencyRaw(rawConn)
			assert.Equal(t, 0, total)
			assert.NoError(t, err)
		})
	})
}
