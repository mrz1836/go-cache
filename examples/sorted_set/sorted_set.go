// Package main provides an example of how to use sorted sets as a priority queue using the go-cache library.
package main

import (
	"context"
	"log"
	"time"

	"github.com/mrz1836/go-cache"
)

func main() {
	ctx := context.Background()

	// Create a new client and pool
	client, err := cache.Connect(
		ctx,
		"redis://localhost:6379",
		0,
		10,
		0,
		240*time.Second,
		true,
		false,
	)
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	}

	const queueKey = "priority-queue"

	// Add items with score = priority (lower score = higher priority)
	if err = cache.SortedSetAdd(ctx, client, queueKey, 10, "low-priority-task"); err != nil {
		log.Fatalf("SortedSetAdd error: %s", err.Error())
	}
	if err = cache.SortedSetAdd(ctx, client, queueKey, 1, "urgent-task"); err != nil {
		log.Fatalf("SortedSetAdd error: %s", err.Error())
	}
	if err = cache.SortedSetAdd(ctx, client, queueKey, 5, "normal-task"); err != nil {
		log.Fatalf("SortedSetAdd error: %s", err.Error())
	}

	// Check queue size
	total, err := cache.SortedSetCard(ctx, client, queueKey)
	if err != nil {
		log.Fatalf("SortedSetCard error: %s", err.Error())
	}
	log.Printf("queue size: %d", total)

	// Peek at all members with scores (ascending by priority)
	members, err := cache.SortedSetRangeWithScores(ctx, client, queueKey, 0, -1)
	if err != nil {
		log.Fatalf("SortedSetRangeWithScores error: %s", err.Error())
	}
	log.Println("queue contents (ascending priority):")
	for _, m := range members {
		log.Printf("  score=%.0f  member=%s", m.Score, m.Member)
	}

	// Dequeue: pop the highest-priority item (lowest score)
	popped, err := cache.SortedSetPopMin(ctx, client, queueKey, 1)
	if err != nil {
		log.Fatalf("SortedSetPopMin error: %s", err.Error())
	}
	if len(popped) > 0 {
		log.Printf("dequeued: %s (priority %.0f)", popped[0].Member, popped[0].Score)
	}

	// Remaining queue size
	remaining, err := cache.SortedSetCard(ctx, client, queueKey)
	if err != nil {
		log.Fatalf("SortedSetCard error: %s", err.Error())
	}
	log.Printf("remaining queue size: %d", remaining)

	// Clean up
	if err = cache.SortedSetRemove(ctx, client, queueKey, "low-priority-task"); err != nil {
		log.Fatalf("SortedSetRemove error: %s", err.Error())
	}
	if err = cache.SortedSetRemove(ctx, client, queueKey, "normal-task"); err != nil {
		log.Fatalf("SortedSetRemove error: %s", err.Error())
	}
	log.Println("cleanup complete")
}
