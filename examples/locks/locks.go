package main

import (
	"log"

	"github.com/mrz1836/go-cache"
)

func main() {

	// Create a new client and pool
	client, err := cache.Connect("redis://localhost:6379", 0, 10, 0, 240, true)
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	}

	// Write a lock
	_, err = cache.WriteLock(client, "test-lock", "test-secret", int64(10))
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	}

	log.Println("lock created successfully")

	// Release a lock
	_, err = cache.ReleaseLock(client, "test-lock", "test-secret")
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	}

	log.Println("lock released successfully")
}
