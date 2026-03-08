package cache

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSortedSetAdd tests the method SortedSetAdd()
func TestSortedSetAdd(t *testing.T) {
	t.Run("sorted set add command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase string
			key      string
			score    float64
			member   interface{}
		}{
			{"basic add", testKey, 1.0, testStringValue},
			{"zero score", testKey, 0.0, testStringValue},
			{"negative score", testKey, -1.0, testStringValue},
			{"max float score", testKey, math.MaxFloat64, testStringValue},
			{"smallest nonzero score", testKey, math.SmallestNonzeroFloat64, testStringValue},
			{"empty key", "", 1.0, testStringValue},
			{"empty member", testKey, 1.0, ""},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				cmd := conn.Command(SortedSetAddCommand, test.key, test.score, test.member)

				err := SortedSetAddRaw(conn, test.key, test.score, test.member)
				require.NoError(t, err)
				assert.True(t, cmd.Called)
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		conn.Command(SortedSetAddCommand, testKey, 1.0, testStringValue)
		err := SortedSetAdd(context.Background(), client, testKey, 1.0, testStringValue)
		require.NoError(t, err)
	})

	t.Run("sorted set add command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Fire the command
		err = SortedSetAddRaw(conn, testKey, 1.5, testStringValue)
		require.NoError(t, err)

		// Check that the command worked
		score, found, err := SortedSetScoreRaw(conn, testKey, testStringValue)
		require.NoError(t, err)
		assert.True(t, found)
		assert.InDelta(t, 1.5, score, 0.0001)
	})
}

// ExampleSortedSetAdd is an example of the method SortedSetAdd()
func ExampleSortedSetAdd() {
	// Load a mocked redis for testing/examples
	client, _ := loadMockRedis()

	// Close connections at end of request
	defer client.Close()

	// Set the key/value
	_ = SortedSetAdd(context.Background(), client, testKey, 1.0, testStringValue)
	fmt.Printf("added member: %v", testStringValue)
	// Output:added member: test-string-value
}

// TestSortedSetAddMany tests the method SortedSetAddMany()
func TestSortedSetAddMany(t *testing.T) {
	t.Run("sorted set add many command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase string
			key      string
			members  []SortedSetMember
		}{
			{"single member", testKey, []SortedSetMember{{Member: "m1", Score: 1.0}}},
			{
				"multiple members",
				testKey,
				[]SortedSetMember{
					{Member: "m1", Score: 1.0},
					{Member: "m2", Score: 2.0},
					{Member: "m3", Score: 3.0},
				},
			},
			{"zero score", testKey, []SortedSetMember{{Member: "m1", Score: 0.0}}},
			{"negative score", testKey, []SortedSetMember{{Member: "m1", Score: -1.0}}},
			{"empty key", "", []SortedSetMember{{Member: "m1", Score: 1.0}}},
			{"empty members", testKey, []SortedSetMember{}},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// Build args the same way the implementation does
				args := make([]interface{}, 0, 1+2*len(test.members))
				args = append(args, test.key)
				for _, m := range test.members {
					args = append(args, m.Score, m.Member)
				}

				// The main command to test
				cmd := conn.Command(SortedSetAddCommand, args...)

				err := SortedSetAddManyRaw(conn, test.key, test.members...)
				require.NoError(t, err)
				assert.True(t, cmd.Called)
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		managedMembers := []SortedSetMember{{Member: "m1", Score: 1.0}, {Member: "m2", Score: 2.0}}
		conn.Command(SortedSetAddCommand, testKey, 1.0, "m1", 2.0, "m2")
		err := SortedSetAddMany(context.Background(), client, testKey, managedMembers...)
		require.NoError(t, err)
	})

	t.Run("sorted set add many command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Fire the command
		members := []SortedSetMember{
			{Member: "m1", Score: 1.0},
			{Member: "m2", Score: 2.0},
		}
		err = SortedSetAddMany(context.Background(), client, testKey, members...)
		require.NoError(t, err)

		// Check that the command worked
		var card int64
		card, err = SortedSetCard(context.Background(), client, testKey)
		require.NoError(t, err)
		assert.Equal(t, int64(2), card)
	})
}

// TestSortedSetRemove tests the method SortedSetRemove()
func TestSortedSetRemove(t *testing.T) {
	t.Run("sorted set remove command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase string
			key      string
			member   interface{}
		}{
			{"basic remove", testKey, testStringValue},
			{"empty key", "", testStringValue},
			{"empty member", testKey, ""},
			{"integer member", testKey, 42},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				cmd := conn.Command(SortedSetRemCommand, test.key, test.member)

				err := SortedSetRemoveRaw(conn, test.key, test.member)
				require.NoError(t, err)
				assert.True(t, cmd.Called)
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		conn.Command(SortedSetRemCommand, testKey, testStringValue)
		err := SortedSetRemove(context.Background(), client, testKey, testStringValue)
		require.NoError(t, err)
	})

	t.Run("sorted set remove command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Add a member first
		err = SortedSetAddRaw(conn, testKey, 1.0, testStringValue)
		require.NoError(t, err)

		// Confirm it's there
		_, found, err := SortedSetScoreRaw(conn, testKey, testStringValue)
		require.NoError(t, err)
		assert.True(t, found)

		// Remove it
		err = SortedSetRemove(context.Background(), client, testKey, testStringValue)
		require.NoError(t, err)

		// Confirm it's gone
		_, found, err = SortedSetScoreRaw(conn, testKey, testStringValue)
		require.NoError(t, err)
		assert.False(t, found)
	})
}

// TestSortedSetRange tests the method SortedSetRange()
func TestSortedSetRange(t *testing.T) {
	t.Run("sorted set range command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase        string
			key             string
			start           int64
			stop            int64
			mockResponse    []interface{}
			expectedMembers int
		}{
			{"full range start=0 stop=-1", testKey, 0, -1, []interface{}{"m1", "m2", "m3"}, 3},
			{"first element only", testKey, 0, 0, []interface{}{"m1"}, 1},
			{"out of range", testKey, 5, 10, []interface{}{}, 0},
			{"empty key", "", 0, -1, []interface{}{}, 0},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				cmd := conn.Command(SortedSetRangeCommand, test.key, test.start, test.stop).Expect(test.mockResponse)

				result, err := SortedSetRangeRaw(conn, test.key, test.start, test.stop)
				require.NoError(t, err)
				assert.Len(t, result, test.expectedMembers)
				assert.True(t, cmd.Called)
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		conn.Command(SortedSetRangeCommand, testKey, int64(0), int64(-1)).Expect([]interface{}{"m1", "m2"})
		rangeResult, err := SortedSetRange(context.Background(), client, testKey, 0, -1)
		require.NoError(t, err)
		assert.Len(t, rangeResult, 2)
	})

	t.Run("sorted set range command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Add members
		members := []SortedSetMember{
			{Member: "a", Score: 1.0},
			{Member: "b", Score: 2.0},
			{Member: "c", Score: 3.0},
		}
		err = SortedSetAddManyRaw(conn, testKey, members...)
		require.NoError(t, err)

		// Get full range
		var result []string
		result, err = SortedSetRange(context.Background(), client, testKey, 0, -1)
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})
}

// TestSortedSetRangeWithScores tests the method SortedSetRangeWithScores()
func TestSortedSetRangeWithScores(t *testing.T) {
	t.Run("sorted set range with scores command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase        string
			key             string
			start           int64
			stop            int64
			mockResponse    []interface{}
			expectedMembers int
			firstMember     string
			firstScore      float64
		}{
			{
				"full range with scores",
				testKey, 0, -1,
				[]interface{}{[]byte("m1"), []byte("1.5"), []byte("m2"), []byte("2.5")},
				2, "m1", 1.5,
			},
			{
				"empty result",
				testKey, 5, 10,
				[]interface{}{},
				0, "", 0,
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test (WITHSCORES variant)
				cmd := conn.Command(SortedSetRangeCommand, test.key, test.start, test.stop, "WITHSCORES").Expect(test.mockResponse)

				result, err := SortedSetRangeWithScoresRaw(conn, test.key, test.start, test.stop)
				require.NoError(t, err)
				assert.Len(t, result, test.expectedMembers)
				assert.True(t, cmd.Called)
				if test.expectedMembers > 0 {
					assert.Equal(t, test.firstMember, result[0].Member)
					assert.InDelta(t, test.firstScore, result[0].Score, 0.0001)
				}
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		conn.Command(SortedSetRangeCommand, testKey, int64(0), int64(-1), "WITHSCORES").Expect(
			[]interface{}{[]byte("m1"), []byte("1.0")},
		)
		wsResult, err := SortedSetRangeWithScores(context.Background(), client, testKey, 0, -1)
		require.NoError(t, err)
		assert.Len(t, wsResult, 1)
	})

	t.Run("sorted set range with scores command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Add members
		members := []SortedSetMember{
			{Member: "a", Score: 1.0},
			{Member: "b", Score: 2.0},
		}
		err = SortedSetAddManyRaw(conn, testKey, members...)
		require.NoError(t, err)

		// Get range with scores
		var result []SortedSetMember
		result, err = SortedSetRangeWithScores(context.Background(), client, testKey, 0, -1)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "a", result[0].Member)
		assert.InDelta(t, 1.0, result[0].Score, 0.0001)
	})
}

// TestSortedSetRangeByScore tests the method SortedSetRangeByScore()
func TestSortedSetRangeByScore(t *testing.T) {
	t.Run("sorted set range by score command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase        string
			key             string
			minScore        string
			maxScore        string
			mockResponse    []interface{}
			expectedMembers int
		}{
			{"neg-inf to pos-inf", testKey, "-inf", "+inf", []interface{}{"m1", "m2"}, 2},
			{"specific range", testKey, "1", "3", []interface{}{"m1"}, 1},
			{"empty result", testKey, "100", "200", []interface{}{}, 0},
			{"empty key", "", "-inf", "+inf", []interface{}{}, 0},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				cmd := conn.Command(SortedSetRangeByScoreCmd, test.key, test.minScore, test.maxScore).Expect(test.mockResponse)

				result, err := SortedSetRangeByScoreRaw(conn, test.key, test.minScore, test.maxScore)
				require.NoError(t, err)
				assert.Len(t, result, test.expectedMembers)
				assert.True(t, cmd.Called)
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		conn.Command(SortedSetRangeByScoreCmd, testKey, "-inf", "+inf").Expect([]interface{}{"m1", "m2"})
		rbsResult, err := SortedSetRangeByScore(context.Background(), client, testKey, "-inf", "+inf")
		require.NoError(t, err)
		assert.Len(t, rbsResult, 2)
	})

	t.Run("sorted set range by score command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Add members
		members := []SortedSetMember{
			{Member: "a", Score: 1.0},
			{Member: "b", Score: 5.0},
			{Member: "c", Score: 10.0},
		}
		err = SortedSetAddManyRaw(conn, testKey, members...)
		require.NoError(t, err)

		// Get by full score range
		var result []string
		result, err = SortedSetRangeByScore(context.Background(), client, testKey, "-inf", "+inf")
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})
}

// TestSortedSetRangeByScoreWithScores tests the method SortedSetRangeByScoreWithScores()
func TestSortedSetRangeByScoreWithScores(t *testing.T) {
	t.Run("sorted set range by score with scores command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase        string
			key             string
			minScore        string
			maxScore        string
			mockResponse    []interface{}
			expectedMembers int
			firstMember     string
			firstScore      float64
		}{
			{
				"neg-inf to pos-inf with scores",
				testKey, "-inf", "+inf",
				[]interface{}{[]byte("m1"), []byte("1.5"), []byte("m2"), []byte("2.5")},
				2, "m1", 1.5,
			},
			{
				"empty result",
				testKey, "100", "200",
				[]interface{}{},
				0, "", 0,
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test (WITHSCORES variant)
				cmd := conn.Command(SortedSetRangeByScoreCmd, test.key, test.minScore, test.maxScore, "WITHSCORES").Expect(test.mockResponse)

				result, err := SortedSetRangeByScoreWithScoresRaw(conn, test.key, test.minScore, test.maxScore)
				require.NoError(t, err)
				assert.Len(t, result, test.expectedMembers)
				assert.True(t, cmd.Called)
				if test.expectedMembers > 0 {
					assert.Equal(t, test.firstMember, result[0].Member)
					assert.InDelta(t, test.firstScore, result[0].Score, 0.0001)
				}
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		conn.Command(SortedSetRangeByScoreCmd, testKey, "-inf", "+inf", "WITHSCORES").Expect(
			[]interface{}{[]byte("m1"), []byte("1.5")},
		)
		rbswsResult, err := SortedSetRangeByScoreWithScores(context.Background(), client, testKey, "-inf", "+inf")
		require.NoError(t, err)
		assert.Len(t, rbswsResult, 1)
	})

	t.Run("sorted set range by score with scores command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Add members
		members := []SortedSetMember{
			{Member: "a", Score: 1.0},
			{Member: "b", Score: 2.0},
		}
		err = SortedSetAddManyRaw(conn, testKey, members...)
		require.NoError(t, err)

		// Get by score range with scores
		var result []SortedSetMember
		result, err = SortedSetRangeByScoreWithScores(context.Background(), client, testKey, "-inf", "+inf")
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "a", result[0].Member)
		assert.InDelta(t, 1.0, result[0].Score, 0.0001)
	})
}

// TestSortedSetPopMin tests the method SortedSetPopMin()
func TestSortedSetPopMin(t *testing.T) {
	t.Run("sorted set pop min command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase        string
			key             string
			count           int64
			mockResponse    []interface{}
			expectedMembers int
			firstMember     string
			firstScore      float64
		}{
			{
				"pop one member",
				testKey, 1,
				[]interface{}{[]byte("m1"), []byte("1.0")},
				1, "m1", 1.0,
			},
			{
				"pop multiple members",
				testKey, 2,
				[]interface{}{[]byte("m1"), []byte("1.0"), []byte("m2"), []byte("2.0")},
				2, "m1", 1.0,
			},
			{
				"pop from empty set",
				testKey, 1,
				[]interface{}{},
				0, "", 0,
			},
			{
				"empty key",
				"", 1,
				[]interface{}{},
				0, "", 0,
			},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				cmd := conn.Command(SortedSetPopMinCommand, test.key, test.count).Expect(test.mockResponse)

				result, err := SortedSetPopMinRaw(conn, test.key, test.count)
				require.NoError(t, err)
				assert.Len(t, result, test.expectedMembers)
				assert.True(t, cmd.Called)
				if test.expectedMembers > 0 {
					assert.Equal(t, test.firstMember, result[0].Member)
					assert.InDelta(t, test.firstScore, result[0].Score, 0.0001)
				}
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		conn.Command(SortedSetPopMinCommand, testKey, int64(1)).Expect(
			[]interface{}{[]byte("m1"), []byte("1.0")},
		)
		popResult, err := SortedSetPopMin(context.Background(), client, testKey, 1)
		require.NoError(t, err)
		assert.Len(t, popResult, 1)
		assert.Equal(t, "m1", popResult[0].Member)
	})

	t.Run("sorted set pop min command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Add members
		members := []SortedSetMember{
			{Member: "a", Score: 1.0},
			{Member: "b", Score: 2.0},
		}
		err = SortedSetAddManyRaw(conn, testKey, members...)
		require.NoError(t, err)

		// Pop the minimum member
		var result []SortedSetMember
		result, err = SortedSetPopMin(context.Background(), client, testKey, 1)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "a", result[0].Member)
		assert.InDelta(t, 1.0, result[0].Score, 0.0001)
	})
}

// TestSortedSetCard tests the method SortedSetCard()
func TestSortedSetCard(t *testing.T) {
	t.Run("sorted set card command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase     string
			key          string
			expectedCard int64
		}{
			{"non-empty set", testKey, 3},
			{"empty set", testKey, 0},
			{"empty key", "", 0},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				// The main command to test
				cmd := conn.Command(SortedSetCardCommand, test.key).Expect(test.expectedCard)

				result, err := SortedSetCardRaw(conn, test.key)
				require.NoError(t, err)
				assert.Equal(t, test.expectedCard, result)
				assert.True(t, cmd.Called)
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		conn.Command(SortedSetCardCommand, testKey).Expect(int64(5))
		cardResult, err := SortedSetCard(context.Background(), client, testKey)
		require.NoError(t, err)
		assert.Equal(t, int64(5), cardResult)
	})

	t.Run("sorted set card command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Empty set should have cardinality 0
		card, err := SortedSetCard(context.Background(), client, testKey)
		require.NoError(t, err)
		assert.Equal(t, int64(0), card)

		// Add members
		members := []SortedSetMember{
			{Member: "a", Score: 1.0},
			{Member: "b", Score: 2.0},
		}
		err = SortedSetAddManyRaw(conn, testKey, members...)
		require.NoError(t, err)

		// Non-empty set should have cardinality 2
		card, err = SortedSetCard(context.Background(), client, testKey)
		require.NoError(t, err)
		assert.Equal(t, int64(2), card)
	})
}

// TestSortedSetScore tests the method SortedSetScore()
func TestSortedSetScore(t *testing.T) {
	t.Run("sorted set score command using mocked redis", func(t *testing.T) {
		t.Parallel()

		// Load redis
		client, conn := loadMockRedis()
		assert.NotNil(t, client)
		defer client.CloseAll(conn)

		tests := []struct {
			testCase      string
			key           string
			member        interface{}
			mockResponse  interface{}
			expectError   bool
			expectedFound bool
			expectedScore float64
		}{
			{"member exists", testKey, "m1", []byte("1.5"), false, true, 1.5},
			{"member not found", testKey, "missing", redis.ErrNil, true, false, 0},
			{"zero score", testKey, "m1", []byte("0"), false, true, 0},
			{"negative score", testKey, "m1", []byte("-1"), false, true, -1.0},
			{"empty key", "", "m1", []byte("2.0"), false, true, 2.0},
			{"empty member", testKey, "", []byte("0.5"), false, true, 0.5},
		}
		for _, test := range tests {
			t.Run(test.testCase, func(t *testing.T) {
				conn.Clear()

				var cmd *redigomock.Cmd
				if test.expectError {
					cmd = conn.Command(SortedSetScoreCommand, test.key, test.member).ExpectError(test.mockResponse.(error))
				} else {
					cmd = conn.Command(SortedSetScoreCommand, test.key, test.member).Expect(test.mockResponse)
				}

				score, found, err := SortedSetScoreRaw(conn, test.key, test.member)
				require.NoError(t, err)
				assert.Equal(t, test.expectedFound, found)
				assert.InDelta(t, test.expectedScore, score, 0.0001)
				assert.True(t, cmd.Called)
			})
		}

		// Test managed variant via mock client pool
		conn.Clear()
		conn.Command(SortedSetScoreCommand, testKey, testStringValue).Expect([]byte("3.14"))
		scoreResult, scoreFound, err := SortedSetScore(context.Background(), client, testKey, testStringValue)
		require.NoError(t, err)
		assert.True(t, scoreFound)
		assert.InDelta(t, 3.14, scoreResult, 0.001)
	})

	t.Run("sorted set score command using real redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		// Load redis
		client, conn, err := loadRealRedis()
		assert.NotNil(t, client)
		require.NoError(t, err)
		defer client.CloseAll(conn)

		// Start with a fresh db
		err = clearRealRedis(conn)
		require.NoError(t, err)

		// Member not found returns false, no error
		score, found, err := SortedSetScore(context.Background(), client, testKey, "missing")
		require.NoError(t, err)
		assert.False(t, found)
		assert.InDelta(t, 0.0, score, 0.0001)

		// Add a member
		err = SortedSetAddRaw(conn, testKey, 1.5, testStringValue)
		require.NoError(t, err)

		// Member found returns true with correct score
		score, found, err = SortedSetScore(context.Background(), client, testKey, testStringValue)
		require.NoError(t, err)
		assert.True(t, found)
		assert.InDelta(t, 1.5, score, 0.0001)
	})
}
