// Package main provides an example of how to use Redis pub/sub messaging using the go-cache library.
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

	const channel = "notifications"

	// Subscribe to the channel
	sub, err := cache.Subscribe(ctx, client, []string{channel})
	if err != nil {
		log.Fatalf("Subscribe error: %s", err.Error())
	}

	// Receive messages in a background goroutine
	received := make(chan struct{})
	go func() {
		defer close(received)
		for msg := range sub.Messages {
			log.Printf("received on %q: %s", msg.Channel, string(msg.Data))
		}
	}()

	// Publish messages from the main goroutine
	for _, payload := range []string{"hello", "world", "goodbye"} {
		n, pubErr := cache.Publish(ctx, client, channel, payload)
		if pubErr != nil {
			log.Fatalf("Publish error: %s", pubErr.Error())
		}
		log.Printf("published %q to %d subscriber(s)", payload, n)
		time.Sleep(50 * time.Millisecond)
	}

	// Allow final messages to arrive before closing
	time.Sleep(200 * time.Millisecond)

	// Close the subscription cleanly
	if err = sub.Close(); err != nil {
		log.Fatalf("Close error: %s", err.Error())
	}

	// Wait for the receiver goroutine to finish
	<-received
	log.Println("pub/sub example complete")
}
