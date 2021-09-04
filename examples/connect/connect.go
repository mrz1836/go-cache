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
	)
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	}

	// Pool is ready
	log.Println("client & pool created")

	// Do something with the connection
	_ = cache.Set(ctx, client, "test-key", "test-value")
}
