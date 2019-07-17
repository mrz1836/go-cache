/*
Package main is an example package using the cache package
*/
package main

import (
	"log"

	"github.com/mrz1836/go-cache"
)

func main() {

	// Create the pool and first connection
	err := cache.Connect("redis://localhost:6379", 0, 10, 0, 240)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Set a key
	err = cache.Set("key-name", "the-value", "dependent-key-1", "dependent-key-2")
	if err != nil {
		log.Fatal(err.Error())
	}

	// Get a key
	value, err := cache.Get("key-name")
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Got value:", value)

	// Kill keys by dependency
	keys, err := cache.KillByDependency("dependent-key-1")
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Keys Removed:", keys)
}
