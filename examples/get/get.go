package main

import (
	"log"
	"strings"

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
	_ = cache.Set(conn, "test-key", "test-value")

	// Get
	val, _ := cache.Get(conn, "test-key")
	if !strings.EqualFold(val, "test-value") {
		log.Fatal("error getting value")
	}
	log.Printf("got: %s", val)
}
