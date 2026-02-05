package cache_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/marcelofabianov/course/config"
	"github.com/marcelofabianov/course/pkg/cache"
)

func Example_basicConnection() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	c, err := cache.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	fmt.Println("Connected to Redis successfully")
}

func Example_setAndGet() {
	cfg, _ := config.Load()
	c, _ := cache.New(cfg)

	ctx := context.Background()
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if err := c.Set(ctx, "user:123", "John Doe", 5*time.Minute); err != nil {
		log.Fatal(err)
	}

	value, err := c.Get(ctx, "user:123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User: %s\n", value)
}

func Example_withExpiration() {
	cfg, _ := config.Load()
	c, _ := cache.New(cfg)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	c.Set(ctx, "session:abc", "active", 1*time.Hour)

	ttl, _ := c.TTL(ctx, "session:abc")
	fmt.Printf("TTL: %v\n", ttl)

	c.Expire(ctx, "session:abc", 30*time.Minute)
}

func Example_incrementCounter() {
	cfg, _ := config.Load()
	c, _ := cache.New(cfg)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	count, _ := c.Increment(ctx, "page:views")
	fmt.Printf("Page views: %d\n", count)

	count, _ = c.Increment(ctx, "page:views")
	fmt.Printf("Page views: %d\n", count)
}

func Example_checkExists() {
	cfg, _ := config.Load()
	c, _ := cache.New(cfg)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	c.Set(ctx, "key1", "value1", time.Minute)
	c.Set(ctx, "key2", "value2", time.Minute)

	count, _ := c.Exists(ctx, "key1", "key2", "key3")
	fmt.Printf("Existing keys: %d\n", count)
}

func Example_deleteKeys() {
	cfg, _ := config.Load()
	c, _ := cache.New(cfg)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	c.Set(ctx, "temp1", "value1", time.Minute)
	c.Set(ctx, "temp2", "value2", time.Minute)

	if err := c.Delete(ctx, "temp1", "temp2"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Keys deleted successfully")
}

func Example_healthCheck() {
	cfg, _ := config.Load()
	c, _ := cache.New(cfg)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	if err := c.HealthCheck(ctx); err != nil {
		log.Printf("Health check failed: %v", err)
		return
	}

	stats := c.Stats()
	fmt.Printf("Pool stats: Hits=%d, Misses=%d, Total=%d\n",
		stats.Hits, stats.Misses, stats.TotalConns)
}

func Example_withLogger() {
	cfg, _ := config.Load()
	c, _ := cache.New(cfg)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	c.Set(ctx, "user:456", "Jane Doe", 10*time.Minute)

	value, _ := c.Get(ctx, "user:456")
	fmt.Printf("User: %s\n", value)
}

func Example_keyNotFound() {
	cfg, _ := config.Load()
	c, _ := cache.New(cfg)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	_, err := c.Get(ctx, "nonexistent:key")
	if err != nil {
		fmt.Println("Key not found error handled")
	}
}
