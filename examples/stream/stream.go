// Package main provides an example of how to use Redis streams as an append-only log using the go-cache library.
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

	const streamKey = "audit-log"

	// Append entries to the stream (auto-generated IDs)
	id1, err := cache.StreamAdd(ctx, client, streamKey, map[string]string{
		"event":  "user.login",
		"user":   "alice",
		"ip":     "192.168.1.1",
	})
	if err != nil {
		log.Fatalf("StreamAdd error: %s", err.Error())
	}
	log.Printf("appended entry id: %s", id1)

	id2, err := cache.StreamAdd(ctx, client, streamKey, map[string]string{
		"event":  "user.purchase",
		"user":   "alice",
		"amount": "49.99",
	})
	if err != nil {
		log.Fatalf("StreamAdd error: %s", err.Error())
	}
	log.Printf("appended entry id: %s", id2)

	id3, err := cache.StreamAdd(ctx, client, streamKey, map[string]string{
		"event": "user.logout",
		"user":  "alice",
	})
	if err != nil {
		log.Fatalf("StreamAdd error: %s", err.Error())
	}
	log.Printf("appended entry id: %s", id3)

	// Check stream length
	length, err := cache.StreamLen(ctx, client, streamKey)
	if err != nil {
		log.Fatalf("StreamLen error: %s", err.Error())
	}
	log.Printf("stream length: %d", length)

	// Read all entries from the beginning
	entries, err := cache.StreamRead(ctx, client, streamKey, "0", 100)
	if err != nil {
		log.Fatalf("StreamRead error: %s", err.Error())
	}
	log.Println("stream entries:")
	for _, entry := range entries {
		log.Printf("  id=%s fields=%v", entry.ID, entry.Fields)
	}

	// Trim the stream to keep only the latest 2 entries
	trimmed, err := cache.StreamTrim(ctx, client, streamKey, 2)
	if err != nil {
		log.Fatalf("StreamTrim error: %s", err.Error())
	}
	log.Printf("trimmed %d entries", trimmed)

	// Confirm final length
	length, err = cache.StreamLen(ctx, client, streamKey)
	if err != nil {
		log.Fatalf("StreamLen error: %s", err.Error())
	}
	log.Printf("stream length after trim: %d", length)
}
