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

	// Get a connection (close pool and connection after)
	conn := client.GetConnection()
	defer client.CloseAll(conn)

	// Run command
	err = cache.Set(conn, "test-key", "test-value", "dependent-key-of-test-key")
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	}

	// Delete
	var total int
	total, err = cache.Delete(client, conn, "test-key")
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	} else if total != 1 {
		log.Fatalf("key was not deleted: %d", total)
	}
	log.Println("key deleted successfully")
}
