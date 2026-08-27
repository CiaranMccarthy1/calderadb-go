package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	calderadb "github.com/Ciaranmccarthy1/calderadb-go/pkg"
)

func main() {
	pool, err := calderadb.NewPool(calderadb.PoolConfig{
		MaxSize: 10,
		Timeout: 5 * time.Second,
		ClientConfig: calderadb.ClientConfig{
			Address: "localhost:9090",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	var wg sync.WaitGroup
	ops := 100
	wg.Add(ops)

	start := time.Now()

	for i := 0; i < ops; i++ {
		go func(id int) {
			defer wg.Done()

			client, err := pool.Get(ctx)
			if err != nil {
				log.Printf("Failed to get client: %v", err)
				return
			}
			defer pool.Put(client)

			key := fmt.Sprintf("user_%d", id)
			value := map[string]interface{}{
				"id":   id,
				"name": fmt.Sprintf("User %d", id),
			}

			if err := client.Set(ctx, "users", key, value); err != nil {
				log.Printf("Failed to set: %v", err)
			}
		}(i)
	}

	wg.Wait()

	fmt.Printf("Completed %d operations in %v\n", ops, time.Since(start))
	fmt.Printf("Average: %.2f ops/sec\n", float64(ops)/time.Since(start).Seconds())
}
