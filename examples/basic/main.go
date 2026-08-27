package main

import (
	"context"
	"fmt"
	"log"

	calderadb "github.com/Ciaranmccarthy1/calderadb-go/pkg"
)

func main() {
	client, err := calderadb.NewClient(calderadb.ClientConfig{
		Address: "localhost:9090",
	})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	if err := client.CreateCollection(ctx, "users"); err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}
	fmt.Println("Created collection: users")

	user := map[string]interface{}{
		"name":  "Alice",
		"age":   30,
		"email": "alice@example.com",
	}
	if err := client.Set(ctx, "users", "alice", user); err != nil {
		log.Fatalf("Failed to set document: %v", err)
	}
	fmt.Println("Added document: alice")

	result, err := client.Get(ctx, "users", "alice")
	if err != nil {
		log.Fatalf("Failed to get document: %v", err)
	}
	fmt.Printf("Retrieved: %+v\n", result)

	stats, err := client.Stats(ctx)
	if err != nil {
		log.Fatalf("Failed to get stats: %v", err)
	}
	fmt.Printf("Stats: %+v\n", stats)

	docs, err := client.FindByPrefix(ctx, "a")
	if err != nil {
		log.Fatalf("Failed to find by prefix: %v", err)
	}
	fmt.Printf("Documents starting with 'a': %d\n", len(docs))
}
