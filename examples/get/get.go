// Package main provides an example of how to get a value from a Redis cache using the go-cache library.
package main

import (
	"context"
	"log"
	"strings"
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
	_ = cache.Set(ctx, client, "test-key", "test-value")

	// Get
	val, _ := cache.Get(ctx, client, "test-key")
	if !strings.EqualFold(val, "test-value") {
		log.Fatal("error getting value")
	}
	log.Printf("got: %s", val)
}
