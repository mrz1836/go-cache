package cache

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func FuzzHashSet(f *testing.F) {
	f.Add("test-hash", "test-key", "test-value", "dep1")
	f.Add("", "", "", "")
	f.Add("unicode-hash-🏠", "unicode-key-🔑", "unicode-value-💎", "unicode-dep-🎯")
	f.Add("hash with spaces", "key with spaces", "value with spaces", "dep with spaces")
	f.Add("hash\nwith\nnewlines", "key\nwith\nnewlines", "value\nwith\nnewlines", "dep\nwith\nnewlines")
	f.Add("hash\x00\x01\x02", "key\x00\x01\x02", "value\x00\x01\x02", "dep\x00\x01\x02")

	f.Fuzz(func(t *testing.T, hashName, hashKey, value, dependency string) {
		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(HashKeySetCommand, hashName, hashKey, value).Expect("OK")
		conn.Command(AddToSetCommand, DependencyPrefix+dependency, hashName).Expect(int64(1))
		conn.Command(MultiCommand).Expect("QUEUED")
		conn.Command(ExecuteCommand).Expect([]interface{}{int64(1)})

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := HashSet(ctx, client, hashName, hashKey, value, dependency)
			assert.NoError(t, err)
		})
	})
}

func FuzzHashGet(f *testing.F) {
	f.Add("test-hash", "test-key")
	f.Add("", "")
	f.Add("unicode-hash-🏠", "unicode-key-🔑")
	f.Add("very-long-hash-"+strings.Repeat("h", 500), "very-long-key-"+strings.Repeat("k", 500))
	f.Add("hash with spaces", "key with spaces")
	f.Add("hash\nwith\nnewlines", "key\nwith\nnewlines")
	f.Add("hash\x00\x01\x02", "key\x00\x01\x02")

	f.Fuzz(func(t *testing.T, hashName, hashKey string) {
		client, conn := loadMockRedis()
		defer client.Close()

		expectedValue := "test-value-for-" + hashKey
		conn.Command(HashGetCommand, hashName, hashKey).Expect(expectedValue)

		ctx := context.Background()

		assert.NotPanics(t, func() {
			result, err := HashGet(ctx, client, hashName, hashKey)
			if err == nil {
				assert.Equal(t, expectedValue, result)
			}
		})
	})
}

func FuzzHashMapSet(f *testing.F) {
	f.Add("test-hash", "key1:value1,key2:value2")
	f.Add("", "")
	f.Add("unicode-hash-🏠", "unicode-key-🔑:unicode-value-💎")
	f.Add("hash with spaces", "key with spaces:value with spaces")

	f.Fuzz(func(t *testing.T, hashName, keyValuePairsStr string) {
		if keyValuePairsStr == "" {
			return
		}

		pairStrings := strings.Split(keyValuePairsStr, ",")
		pairs := make([][2]interface{}, len(pairStrings))
		args := make([]interface{}, 2*len(pairStrings)+1)
		args[0] = hashName

		for i, pairStr := range pairStrings {
			keyValue := strings.SplitN(pairStr, ":", 2)
			if len(keyValue) < 2 {
				keyValue = append(keyValue, "")
			}
			pairs[i] = [2]interface{}{keyValue[0], keyValue[1]}
			args[2*i+1] = keyValue[0]
			args[2*i+2] = keyValue[1]
		}

		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(HashMapSetCommand, args...).Expect("OK")
		conn.Command(MultiCommand).Expect("QUEUED")
		conn.Command(ExecuteCommand).Expect([]interface{}{})

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := HashMapSet(ctx, client, hashName, pairs)
			assert.NoError(t, err)
		})
	})
}

func FuzzHashMapGet(f *testing.F) {
	f.Add("test-hash", "key1,key2,key3")
	f.Add("", "")
	f.Add("unicode-hash-🏠", "unicode-key-🔑,another-unicode-🚀")
	f.Add("hash with spaces", "key with spaces")
	f.Add("very-long-hash", strings.Repeat("long-key-", 100))

	f.Fuzz(func(t *testing.T, hashName, keysStr string) {
		if keysStr == "" {
			return
		}

		keys := strings.Split(keysStr, ",")

		client, conn := loadMockRedis()
		defer client.Close()

		interfaceKeys := make([]interface{}, len(keys))
		expectedValues := make([]interface{}, len(keys))
		for i, key := range keys {
			interfaceKeys[i] = key
			expectedValues[i] = []byte("value-for-" + key)
		}

		args := append([]interface{}{hashName}, interfaceKeys...)
		conn.Command(HashMapGetCommand, args...).Expect(expectedValues)

		ctx := context.Background()

		assert.NotPanics(t, func() {
			result, err := HashMapGet(ctx, client, hashName, interfaceKeys...)
			if err == nil {
				assert.Len(t, result, len(keys))
			}
		})
	})
}

func FuzzHashMapSetExp(f *testing.F) {
	f.Add("test-hash", "key1:value1,key2:value2", int64(60))
	f.Add("unicode-hash-🏠", "unicode-key-🔑:unicode-value-💎", int64(3600))
	f.Add("hash with spaces", "key with spaces:value with spaces", int64(1))

	f.Fuzz(func(t *testing.T, hashName, keyValuePairsStr string, ttlSeconds int64) {
		if keyValuePairsStr == "" {
			return
		}

		pairStrings := strings.Split(keyValuePairsStr, ",")
		pairs := make([][2]interface{}, len(pairStrings))
		args := make([]interface{}, 2*len(pairStrings)+1)
		args[0] = hashName

		for i, pairStr := range pairStrings {
			keyValue := strings.SplitN(pairStr, ":", 2)
			if len(keyValue) < 2 {
				keyValue = append(keyValue, "")
			}
			pairs[i] = [2]interface{}{keyValue[0], keyValue[1]}
			args[2*i+1] = keyValue[0]
			args[2*i+2] = keyValue[1]
		}

		if ttlSeconds < 0 {
			ttlSeconds = 0
		}
		if ttlSeconds > 86400 {
			ttlSeconds = 86400
		}

		client, conn := loadMockRedis()
		defer client.Close()

		ttl := time.Duration(ttlSeconds) * time.Second

		conn.Command(HashMapSetCommand, args...).Expect("OK")
		if ttlSeconds > 0 {
			conn.Command(ExpireCommand, hashName, ttlSeconds).Expect(int64(1))
		}
		conn.Command(MultiCommand).Expect("QUEUED")
		conn.Command(ExecuteCommand).Expect([]interface{}{})

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := HashMapSetExp(ctx, client, hashName, pairs, ttl)
			if ttlSeconds > 0 {
				assert.NoError(t, err)
			}
		})
	})
}

func FuzzHashMapOperationsRoundTrip(f *testing.F) {
	f.Add("test-hash", "key1", "value1", "key2", "value2")
	f.Add("", "", "", "", "")
	f.Add("unicode-hash-🏠", "unicode-key1-🔑", "unicode-value1-💎", "unicode-key2-🚀", "unicode-value2-⭐")

	f.Fuzz(func(t *testing.T, hashName, key1, value1, key2, value2 string) {
		if key1 == "" && key2 == "" {
			return
		}

		client, conn := loadMockRedis()
		defer client.Close()

		var pairs [][2]interface{}
		var setArgs []interface{}
		var getKeys []interface{}
		var expectedValues []interface{}

		setArgs = append(setArgs, hashName)

		if key1 != "" {
			pairs = append(pairs, [2]interface{}{key1, value1})
			setArgs = append(setArgs, key1, value1)
			getKeys = append(getKeys, key1)
			expectedValues = append(expectedValues, []byte(value1))
		}

		if key2 != "" && key2 != key1 {
			pairs = append(pairs, [2]interface{}{key2, value2})
			setArgs = append(setArgs, key2, value2)
			getKeys = append(getKeys, key2)
			expectedValues = append(expectedValues, []byte(value2))
		}

		if len(pairs) == 0 {
			return
		}

		conn.Command(HashMapSetCommand, setArgs...).Expect("OK")
		conn.Command(MultiCommand).Expect("QUEUED")
		conn.Command(ExecuteCommand).Expect([]interface{}{})

		getArgs := append([]interface{}{hashName}, getKeys...)
		conn.Command(HashMapGetCommand, getArgs...).Expect(expectedValues)

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := HashMapSet(ctx, client, hashName, pairs)
			assert.NoError(t, err)

			result, err := HashMapGet(ctx, client, hashName, getKeys...)
			if err == nil {
				assert.Len(t, result, len(getKeys))
				for i, key := range getKeys {
					expectedValue := ""
					switch key {
					case key1:
						expectedValue = value1
					case key2:
						expectedValue = value2
					}
					if i < len(result) {
						assert.Equal(t, expectedValue, result[i])
					}
				}
			}
		})
	})
}
