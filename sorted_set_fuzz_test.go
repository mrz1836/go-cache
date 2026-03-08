package cache

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// FuzzSortedSetAdd fuzzes the SortedSetAdd function with key, member, and score inputs
func FuzzSortedSetAdd(f *testing.F) {
	// Seed corpus with interesting values
	f.Add("test-zset", "test-member", 1.0)
	f.Add("", "", 0.0)
	f.Add("unicode-zset-🏠", "unicode-member-🔑", -1.0)
	f.Add("zset with spaces", "member with spaces", 42.5)
	f.Add("zset\nwith\nnewlines", "member\nwith\nnewlines", 0.0)
	f.Add("zset\x00\x01\x02", "member\x00\x01\x02", 0.0)
	f.Add(strings.Repeat("long-zset-", 100), strings.Repeat("long-member-", 100), 9.99e99)
	f.Add("neg-score", "member", -9.99e99)

	f.Fuzz(func(t *testing.T, key, member string, score float64) {
		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(SortedSetAddCommand, key, score, member).Expect(int64(1))

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := SortedSetAdd(ctx, client, key, score, member)
			assert.NoError(t, err)
		})
	})
}

// FuzzSortedSetRange fuzzes the SortedSetRange function with key, start, and stop inputs
func FuzzSortedSetRange(f *testing.F) {
	// Seed corpus with interesting range values
	f.Add("test-zset", int64(0), int64(-1))
	f.Add("", int64(0), int64(0))
	f.Add("unicode-zset-🏠", int64(0), int64(100))
	f.Add("zset with spaces", int64(-1), int64(-1))
	f.Add("zset\nwith\nnewlines", int64(5), int64(10))
	f.Add("zset\x00\x01\x02", int64(0), int64(-1))
	f.Add(strings.Repeat("long-zset-", 50), int64(0), int64(999))

	f.Fuzz(func(t *testing.T, key string, start, stop int64) {
		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(SortedSetRangeCommand, key, start, stop).Expect([]interface{}{})

		ctx := context.Background()

		assert.NotPanics(t, func() {
			result, err := SortedSetRange(ctx, client, key, start, stop)
			if err == nil {
				assert.IsType(t, []string{}, result)
			}
		})
	})
}

// FuzzSortedSetRangeByScore fuzzes the SortedSetRangeByScore function with key, min, and max inputs
func FuzzSortedSetRangeByScore(f *testing.F) {
	// Seed corpus with interesting score range values
	f.Add("test-zset", "-inf", "+inf")
	f.Add("", "0", "100")
	f.Add("unicode-zset-🏠", "-inf", "0")
	f.Add("zset with spaces", "0", "+inf")
	f.Add("zset\nwith\nnewlines", "-1", "1")
	f.Add("zset\x00\x01\x02", "-inf", "+inf")
	f.Add(strings.Repeat("long-zset-", 50), "1.5", "9.99e99")
	f.Add("binary-zset", "0.0", "0.0")

	f.Fuzz(func(t *testing.T, key, minScore, maxScore string) {
		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(SortedSetRangeByScoreCmd, key, minScore, maxScore).Expect([]interface{}{})

		ctx := context.Background()

		assert.NotPanics(t, func() {
			result, err := SortedSetRangeByScore(ctx, client, key, minScore, maxScore)
			if err == nil {
				assert.IsType(t, []string{}, result)
			}
		})
	})
}
