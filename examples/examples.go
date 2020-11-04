/*
Package main is an example package using the cache package
*/
package main

/*
func main() {

	// Create the pool and first connection
	err := cache.Connect("redis://localhost:6379", 0, 10, 0, 240, true, redis.DialKeepAlive(10*time.Second))
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
	var value string
	if value, err = cache.Get("user-michael"); err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Got value:", value)

	// Remove all keys based on a dependent key getting busted
	// Example: the user updates their profile or record, hence the key "user-23" would get busted
	var keys int
	if keys, err = cache.KillByDependency("user-23"); err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Keys Removed:", keys)

	// Attempting to try and get the value now will not work, as the key has been removed
	if value, _ = cache.Get("user-michael"); value == "" {
		log.Println("No value found for key:", "user-michael")
	}

	// Set a redis lock
	var locked bool
	if locked, err = cache.WriteLock("my-lock-key", "the-lock-secret", int64(10)); err != nil {
		log.Fatal(err.Error())
	} else if locked {
		log.Println("lock succeeded")
	}
}*/
