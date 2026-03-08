package cache

import (
	"context"
	"errors"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

// SortedSetMember represents a member of a sorted set with its associated score
type SortedSetMember struct {
	Member interface{}
	Score  float64
}

// parseSortedSetWithScores parses the flat alternating [member, score, ...] response
// returned by ZRANGE WITHSCORES, ZRANGEBYSCORE WITHSCORES, and ZPOPMIN
func parseSortedSetWithScores(values []interface{}) ([]SortedSetMember, error) {
	if len(values)%2 != 0 {
		return nil, redis.ErrNil
	}
	members := make([]SortedSetMember, 0, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		memberBytes, err := redis.Bytes(values[i], nil)
		if err != nil {
			return nil, err
		}
		scoreStr, err := redis.String(values[i+1], nil)
		if err != nil {
			return nil, err
		}
		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			return nil, err
		}
		members = append(members, SortedSetMember{
			Member: string(memberBytes),
			Score:  score,
		})
	}
	return members, nil
}

// SortedSetAdd adds a single member with a score to a sorted set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SortedSetAddRaw()
func SortedSetAdd(ctx context.Context, client *Client, key string, score float64, member interface{}) error {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return err
	}
	defer client.CloseConnection(conn)
	return SortedSetAddRaw(conn, key, score, member)
}

// SortedSetAddRaw adds a single member with a score to a sorted set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/zadd
func SortedSetAddRaw(conn redis.Conn, key string, score float64, member interface{}) (err error) {
	_, err = conn.Do(SortedSetAddCommand, key, score, member)
	return err
}

// SortedSetAddMany adds multiple members with scores to a sorted set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SortedSetAddManyRaw()
func SortedSetAddMany(ctx context.Context, client *Client, key string, members ...SortedSetMember) error {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return err
	}
	defer client.CloseConnection(conn)
	return SortedSetAddManyRaw(conn, key, members...)
}

// SortedSetAddManyRaw adds multiple members with scores to a sorted set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/zadd
func SortedSetAddManyRaw(conn redis.Conn, key string, members ...SortedSetMember) (err error) {
	args := make([]interface{}, 0, 1+2*len(members))
	args = append(args, key)
	for _, m := range members {
		args = append(args, m.Score, m.Member)
	}
	_, err = conn.Do(SortedSetAddCommand, args...)
	return err
}

// SortedSetRemove removes a member from a sorted set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SortedSetRemoveRaw()
func SortedSetRemove(ctx context.Context, client *Client, key string, member interface{}) error {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return err
	}
	defer client.CloseConnection(conn)
	return SortedSetRemoveRaw(conn, key, member)
}

// SortedSetRemoveRaw removes a member from a sorted set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/zrem
func SortedSetRemoveRaw(conn redis.Conn, key string, member interface{}) (err error) {
	_, err = conn.Do(SortedSetRemCommand, key, member)
	return err
}

// SortedSetRange returns the specified range of members in a sorted set (by index)
// Results are ordered from lowest to highest score
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SortedSetRangeRaw()
func SortedSetRange(ctx context.Context, client *Client, key string, start, stop int64) ([]string, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer client.CloseConnection(conn)
	return SortedSetRangeRaw(conn, key, start, stop)
}

// SortedSetRangeRaw returns the specified range of members in a sorted set (by index)
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/zrange
func SortedSetRangeRaw(conn redis.Conn, key string, start, stop int64) ([]string, error) {
	return redis.Strings(conn.Do(SortedSetRangeCommand, key, start, stop))
}

// SortedSetRangeWithScores returns the specified range of members with their scores in a sorted set
// Results are ordered from lowest to highest score
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SortedSetRangeWithScoresRaw()
func SortedSetRangeWithScores(ctx context.Context, client *Client, key string, start, stop int64) ([]SortedSetMember, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer client.CloseConnection(conn)
	return SortedSetRangeWithScoresRaw(conn, key, start, stop)
}

// SortedSetRangeWithScoresRaw returns the specified range of members with their scores in a sorted set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/zrange
func SortedSetRangeWithScoresRaw(conn redis.Conn, key string, start, stop int64) ([]SortedSetMember, error) {
	values, err := redis.Values(conn.Do(SortedSetRangeCommand, key, start, stop, "WITHSCORES"))
	if err != nil {
		return nil, err
	}
	return parseSortedSetWithScores(values)
}

// SortedSetRangeByScore returns all members in a sorted set with scores between minScore and maxScore
// minScore and maxScore are string representations (e.g. "-inf", "+inf", "1.5")
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SortedSetRangeByScoreRaw()
func SortedSetRangeByScore(ctx context.Context, client *Client, key, minScore, maxScore string) ([]string, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer client.CloseConnection(conn)
	return SortedSetRangeByScoreRaw(conn, key, minScore, maxScore)
}

// SortedSetRangeByScoreRaw returns all members in a sorted set with scores between minScore and maxScore
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/zrangebyscore
func SortedSetRangeByScoreRaw(conn redis.Conn, key, minScore, maxScore string) ([]string, error) {
	return redis.Strings(conn.Do(SortedSetRangeByScoreCmd, key, minScore, maxScore))
}

// SortedSetRangeByScoreWithScores returns all members with scores in a sorted set between minScore and maxScore
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SortedSetRangeByScoreWithScoresRaw()
func SortedSetRangeByScoreWithScores(ctx context.Context, client *Client, key, minScore, maxScore string) ([]SortedSetMember, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer client.CloseConnection(conn)
	return SortedSetRangeByScoreWithScoresRaw(conn, key, minScore, maxScore)
}

// SortedSetRangeByScoreWithScoresRaw returns all members with scores in a sorted set between minScore and maxScore
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/zrangebyscore
func SortedSetRangeByScoreWithScoresRaw(conn redis.Conn, key, minScore, maxScore string) ([]SortedSetMember, error) {
	values, err := redis.Values(conn.Do(SortedSetRangeByScoreCmd, key, minScore, maxScore, "WITHSCORES"))
	if err != nil {
		return nil, err
	}
	return parseSortedSetWithScores(values)
}

// SortedSetPopMin removes and returns the member with the lowest score from a sorted set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SortedSetPopMinRaw()
func SortedSetPopMin(ctx context.Context, client *Client, key string, count int64) ([]SortedSetMember, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer client.CloseConnection(conn)
	return SortedSetPopMinRaw(conn, key, count)
}

// SortedSetPopMinRaw removes and returns the member(s) with the lowest score(s) from a sorted set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/zpopmin
func SortedSetPopMinRaw(conn redis.Conn, key string, count int64) ([]SortedSetMember, error) {
	values, err := redis.Values(conn.Do(SortedSetPopMinCommand, key, count))
	if err != nil {
		return nil, err
	}
	return parseSortedSetWithScores(values)
}

// SortedSetCard returns the number of members in a sorted set
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SortedSetCardRaw()
func SortedSetCard(ctx context.Context, client *Client, key string) (int64, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return 0, err
	}
	defer client.CloseConnection(conn)
	return SortedSetCardRaw(conn, key)
}

// SortedSetCardRaw returns the number of members in a sorted set
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/zcard
func SortedSetCardRaw(conn redis.Conn, key string) (int64, error) {
	return redis.Int64(conn.Do(SortedSetCardCommand, key))
}

// SortedSetScore returns the score of a member in a sorted set
// Returns (score, true, nil) when found, (0, false, nil) when not found
// Creates a new connection and closes connection at end of function call
//
// Custom connections use method: SortedSetScoreRaw()
func SortedSetScore(ctx context.Context, client *Client, key string, member interface{}) (float64, bool, error) {
	conn, err := client.GetConnectionWithContext(ctx)
	if err != nil {
		return 0, false, err
	}
	defer client.CloseConnection(conn)
	return SortedSetScoreRaw(conn, key, member)
}

// SortedSetScoreRaw returns the score of a member in a sorted set
// Returns (score, true, nil) when found, (0, false, nil) when not found
// Uses existing connection (does not close connection)
//
// Spec: https://redis.io/commands/zscore
func SortedSetScoreRaw(conn redis.Conn, key string, member interface{}) (float64, bool, error) {
	score, err := redis.Float64(conn.Do(SortedSetScoreCommand, key, member))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return score, true, nil
}
