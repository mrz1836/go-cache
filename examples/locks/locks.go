// Package main shows an example of how to create a lock using the go-cache library.
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

	// Write a lock
	_, err = cache.WriteLock(ctx, client, "test-lock", "test-secret", int64(10))
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	}

	log.Println("lock created successfully")

	// Release a lock
	_, err = cache.ReleaseLock(ctx, client, "test-lock", "test-secret")
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	}

	log.Println("lock released successfully")
}
