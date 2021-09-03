package main

import (
	"log"
	"time"

	"github.com/mrz1836/go-cache"
)

func main() {

	// Create a new client and pool
	client, err := cache.Connect(
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

	// Run command
	_ = cache.SetExp(client, "test-key", "test-value", 10*time.Minute)
}
