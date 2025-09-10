// Package main provides an example of how to delete a key from a Redis server using the go-cache library.
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

	// Run command
	err = cache.Set(ctx, client, "test-key", "test-value", "dependent-key-of-test-key")
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	}

	// Delete
	var total int
	total, err = cache.Delete(ctx, client, "test-key")
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	} else if total != 1 {
		log.Fatalf("key was not deleted: %d", total)
	}
	log.Println("key deleted successfully")
}
