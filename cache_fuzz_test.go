package cache

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func FuzzSetAndGet(f *testing.F) {
	f.Add("test-key", "test-value", "dep1", "dep2")
	f.Add("", "", "", "")
	f.Add("unicode-key-🔑", "unicode-value-💎", "unicode-dep-🏠", "")
	f.Add("very-long-key-"+strings.Repeat("x", 1000), "very-long-value-"+strings.Repeat("y", 1000), "dep", "")
	f.Add("key\nwith\nnewlines", "value\nwith\nnewlines", "dep\nwith\nnewlines", "")
	f.Add("key with spaces", "value with spaces", "dep with spaces", "")
	f.Add("key\x00\x01\x02", "value\x00\x01\x02", "dep\x00\x01\x02", "")

	f.Fuzz(func(t *testing.T, key, value, dep1, dep2 string) {
		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(SetCommand, key, value).Expect("OK")
		conn.Command(AddToSetCommand, DependencyPrefix+dep1, key).Expect(int64(1))
		if dep2 != "" {
			conn.Command(AddToSetCommand, DependencyPrefix+dep2, key).Expect(int64(1))
		}
		conn.Command(MultiCommand).Expect("QUEUED")
		conn.Command(ExecuteCommand).Expect([]interface{}{int64(1)})

		conn.Command(GetCommand, key).Expect(value)

		ctx := context.Background()
		dependencies := []string{dep1}
		if dep2 != "" {
			dependencies = append(dependencies, dep2)
		}

		assert.NotPanics(t, func() {
			err := Set(ctx, client, key, value, dependencies...)
			assert.NoError(t, err)
		})

		assert.NotPanics(t, func() {
			result, err := Get(ctx, client, key)
			if err == nil {
				assert.Equal(t, value, result)
			}
		})
	})
}

func FuzzSetList(f *testing.F) {
	f.Add("item1,item2,item3")
	f.Add("")
	f.Add("unicode-item-🎯,another-unicode-🚀")
	f.Add("single-item")
	f.Add(strings.Repeat("long-item-", 100))
	f.Add("item\nwith\nnewlines")
	f.Add("item with spaces")
	f.Add("item\x00\x01\x02")

	f.Fuzz(func(t *testing.T, itemsStr string) {
		var items []string
		if itemsStr != "" {
			items = strings.Split(itemsStr, ",")
		}
		client, conn := loadMockRedis()
		defer client.Close()

		key := "test-list-key"

		args := make([]interface{}, len(items)+1)
		args[0] = key
		for i, item := range items {
			args[i+1] = item
		}

		if len(items) > 0 {
			conn.Command(ListPushCommand, args...).Expect(int64(len(items)))
		}

		expectedValues := make([]interface{}, len(items))
		for i, item := range items {
			expectedValues[i] = []byte(item)
		}
		conn.Command(ListRangeCommand, key, 0, -1).Expect(expectedValues)

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := SetList(ctx, client, key, items)
			if len(items) > 0 {
				assert.NoError(t, err)
			}
		})

		assert.NotPanics(t, func() {
			result, err := GetList(ctx, client, key)
			if err == nil && len(items) > 0 {
				assert.Equal(t, items, result)
			}
		})
	})
}

func FuzzSetToJSON(f *testing.F) {
	type TestStruct struct {
		Name  string   `json:"name"`
		Value int      `json:"value"`
		Tags  []string `json:"tags,omitempty"`
	}

	f.Add("test-key", "TestName", 42)
	f.Add("", "", 0)
	f.Add("unicode-key-🔑", "unicode-name-🎯", -1)
	f.Add("very-long-key-"+strings.Repeat("x", 100), strings.Repeat("y", 100), 999999)
	f.Add("key\nwith\nnewlines", "name\nwith\nnewlines", 123)

	f.Fuzz(func(t *testing.T, key, name string, value int) {
		client, conn := loadMockRedis()
		defer client.Close()

		testData := TestStruct{
			Name:  name,
			Value: value,
			Tags:  []string{"tag1", "tag2"},
		}

		expectedJSON, err := json.Marshal(&testData)
		if err != nil {
			t.Skip("Invalid JSON data")
			return
		}

		conn.Command(SetCommand, key, string(expectedJSON)).Expect("OK")
		conn.Command(MultiCommand).Expect("QUEUED")
		conn.Command(ExecuteCommand).Expect([]interface{}{})

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := SetToJSON(ctx, client, key, testData, 0)
			assert.NoError(t, err)
		})
	})
}

func FuzzSetExp(f *testing.F) {
	f.Add("test-key", "test-value", int64(60))
	f.Add("", "", int64(0))
	f.Add("unicode-key-🔑", "unicode-value-💎", int64(3600))
	f.Add("key with spaces", "value with spaces", int64(1))

	f.Fuzz(func(t *testing.T, key, value string, ttlSeconds int64) {
		if ttlSeconds < 0 {
			ttlSeconds = 0
		}
		if ttlSeconds > 86400 {
			ttlSeconds = 86400
		}

		client, conn := loadMockRedis()
		defer client.Close()

		ttl := time.Duration(ttlSeconds) * time.Second

		conn.Command(SetExpirationCommand, key, ttlSeconds, value).Expect("OK")
		conn.Command(MultiCommand).Expect("QUEUED")
		conn.Command(ExecuteCommand).Expect([]interface{}{})

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := SetExp(ctx, client, key, value, ttl)
			if ttlSeconds > 0 {
				assert.NoError(t, err)
			}
		})
	})
}

func FuzzExists(f *testing.F) {
	f.Add("test-key")
	f.Add("")
	f.Add("unicode-key-🔑")
	f.Add("very-long-key-" + strings.Repeat("x", 1000))
	f.Add("key\nwith\nnewlines")
	f.Add("key with spaces")
	f.Add("key\x00\x01\x02")

	f.Fuzz(func(t *testing.T, key string) {
		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(ExistsCommand, key).Expect(int64(1))

		ctx := context.Background()

		assert.NotPanics(t, func() {
			result, err := Exists(ctx, client, key)
			if err == nil {
				assert.IsType(t, bool(true), result)
			}
		})
	})
}

func FuzzExpire(f *testing.F) {
	f.Add("test-key", int64(60))
	f.Add("", int64(0))
	f.Add("unicode-key-🔑", int64(3600))
	f.Add("key with spaces", int64(1))

	f.Fuzz(func(t *testing.T, key string, ttlSeconds int64) {
		if ttlSeconds < 0 {
			ttlSeconds = 0
		}
		if ttlSeconds > 86400 {
			ttlSeconds = 86400
		}

		client, conn := loadMockRedis()
		defer client.Close()

		ttl := time.Duration(ttlSeconds) * time.Second

		conn.Command(ExpireCommand, key, ttlSeconds).Expect("OK")

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := Expire(ctx, client, key, ttl)
			assert.NoError(t, err)
		})
	})
}

func FuzzGetBytes(f *testing.F) {
	f.Add("test-key", []byte("test-value"))
	f.Add("", []byte(""))
	f.Add("unicode-key-🔑", []byte("unicode-value-💎"))
	f.Add("binary-key", []byte{0x00, 0x01, 0x02, 0xFF})

	f.Fuzz(func(t *testing.T, key string, value []byte) {
		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(GetCommand, key).Expect(value)

		ctx := context.Background()

		assert.NotPanics(t, func() {
			result, err := GetBytes(ctx, client, key)
			if err == nil {
				assert.Equal(t, value, result)
			}
		})
	})
}

func FuzzDeleteWithoutDependency(f *testing.F) {
	f.Add("key1,key2,key3")
	f.Add("")
	f.Add("unicode-key-🔑,another-unicode-🚀")
	f.Add("single-key")
	f.Add(strings.Repeat("long-key-", 100))

	f.Fuzz(func(t *testing.T, keysStr string) {
		if keysStr == "" {
			return
		}

		keys := strings.Split(keysStr, ",")

		client, conn := loadMockRedis()
		defer client.Close()

		for _, key := range keys {
			conn.Command(DeleteCommand, key).Expect(int64(1))
		}

		ctx := context.Background()

		assert.NotPanics(t, func() {
			total, err := DeleteWithoutDependency(ctx, client, keys...)
			if err == nil {
				assert.GreaterOrEqual(t, total, 0)
				assert.LessOrEqual(t, total, len(keys))
			}
		})
	})
}

func FuzzPing(f *testing.F) {
	f.Add(true)

	f.Fuzz(func(t *testing.T, _ bool) {
		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(PingCommand).Expect("PONG")

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := Ping(ctx, client)
			assert.NoError(t, err)
		})
	})
}
