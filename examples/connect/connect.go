// Package main provides an example of how to connect to a Redis server using the go-cache library.
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

	// Pool is ready
	log.Println("client & pool created")

	// Test the connection
	if err = cache.Ping(ctx, client); err != nil {
		log.Fatalf("ping failed - connection error: %s", err.Error())
	}

	// Do something with the connection
	_ = cache.Set(ctx, client, "test-key", "test-value")
}
