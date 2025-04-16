// Package main shows how to set a value in the cache with an expiration time
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
	_ = cache.SetExp(ctx, client, "test-key", "test-value", 10*time.Minute)
}
