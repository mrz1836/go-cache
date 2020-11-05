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

	// Run command
	err = cache.Set(client, "test-key", "test-value", "dependent-key-of-test-key")
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	}

	// Delete
	var total int
	total, err = cache.Delete(client, "test-key")
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	} else if total != 1 {
		log.Fatalf("key was not deleted: %d", total)
	}
	log.Println("key deleted successfully")
}
