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

	// Set a key that is dependent to other keys
	// Example: this key has data that is related to "user-23" and "user-profile-23" keys
	// If those keys are removed/changed, this "user-michael" key should be removed
	err = cache.Set("user-michael", "my name is Michael and my ID is 23", "user-23", "user-profile-23")
	if err != nil {
		log.Fatal(err.Error())
	}

	// Get the value for the key we set that has a value
	value, err := cache.Get("user-michael")
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Got value:", value)

	// Remove all keys based on a dependent key getting busted
	// Example: the user updates their profile or record, hence the key "user-23" would get busted
	keys, err := cache.KillByDependency("user-23")
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Keys Removed:", keys)

	// Attempting to try and get the value now will not work, as the key has been removed
	value, _ = cache.Get("user-michael")
	if value == "" {
		log.Println("No value found for key:", "user-michael")
	}
}
