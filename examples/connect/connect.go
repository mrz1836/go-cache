package main

import (
	"log"

	"github.com/mrz1836/go-cache"
)

func main() {

	// Create a new client and pool
	client, err := cache.Connect(
		"redis://localhost:6379",
		0,
		10,
		0,
		240,
		true,
	)
	if err != nil {
		log.Fatalf("error occurred: %s", err.Error())
	}

	// Pool is ready
	log.Println("client & pool created")

	// Get a connection (close pool and connection after)
	conn := client.GetConnection()
	defer client.CloseAll(conn)

	// Do something with the connection
	_ = cache.Set(conn, "test-key", "test-value")
}
