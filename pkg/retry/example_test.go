package retry_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/marcelofabianov/course/pkg/retry"
)

// Example_basicRetry demonstrates basic retry functionality.
func Example_basicRetry() {
	ctx := context.Background()

	// Configure retry with exponential backoff
	config := &retry.Config{
		MaxAttempts: 3,
		Strategy:    retry.NewDefaultExponentialBackoff(),
	}

	// Function that may fail
	counter := 0
	fn := func(ctx context.Context) error {
		counter++
		if counter < 3 {
			return errors.New("temporary error")
		}
		fmt.Println("Success!")
		return nil
	}

	err := retry.Do(ctx, config, fn)
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// Success!
}

// Example_exponentialBackoff demonstrates exponential backoff strategy.
func Example_exponentialBackoff() {
	ctx := context.Background()

	// Custom exponential backoff: 100ms -> 200ms -> 400ms -> 800ms (max)
	config := &retry.Config{
		MaxAttempts: 5,
		Strategy: retry.NewExponentialBackoff(retry.ExponentialBackoffConfig{
			Min:    100 * time.Millisecond,
			Max:    800 * time.Millisecond,
			Factor: 2.0,
			Jitter: true, // Add randomization to prevent thundering herd
		}),
		OnRetry: func(attempt int, err error) {
			fmt.Printf("Retry attempt %d after error: %v\n", attempt+1, err)
		},
	}

	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("attempt %d failed", attempts)
		}
		fmt.Println("Operation succeeded")
		return nil
	}

	_ = retry.Do(ctx, config, fn)

	// Output will vary due to jitter, but will show:
	// Retry attempt 1 after error: attempt 1 failed
	// Retry attempt 2 after error: attempt 2 failed
	// Operation succeeded
}

// Example_constantBackoff demonstrates constant delay between retries.
func Example_constantBackoff() {
	ctx := context.Background()

	// Retry with constant 1 second delay
	config := &retry.Config{
		MaxAttempts: 3,
		Strategy:    retry.NewConstantBackoff(1 * time.Second),
	}

	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		return fmt.Errorf("failed attempt %d", attempts)
	}

	err := retry.Do(ctx, config, fn)

	// Extract the base error message
	errMsg := err.Error()
	if strings.Contains(errMsg, "all retry attempts failed") {
		fmt.Println("Final error: all retry attempts failed")
	} else {
		fmt.Printf("Final error: %v\n", err)
	}
	fmt.Printf("Total attempts: %d\n", attempts)

	// Output:
	// Final error: all retry attempts failed
	// Total attempts: 4
}

// Example_linearBackoff demonstrates linear backoff strategy.
func Example_linearBackoff() {
	ctx := context.Background()

	// Linear backoff: 1s -> 2s -> 3s -> 4s (capped at 5s)
	config := &retry.Config{
		MaxAttempts: 3,
		Strategy:    retry.NewLinearBackoff(1*time.Second, 5*time.Second),
	}

	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("not ready yet")
		}
		fmt.Println("Ready!")
		return nil
	}

	_ = retry.Do(ctx, config, fn)

	// Output:
	// Ready!
}

// Example_contextCancellation demonstrates context cancellation during retries.
func Example_contextCancellation() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	config := &retry.Config{
		MaxAttempts: 10,
		Strategy:    retry.NewConstantBackoff(200 * time.Millisecond),
	}

	fn := func(ctx context.Context) error {
		return errors.New("persistent error")
	}

	err := retry.Do(ctx, config, fn)

	// Extract the base error type
	if strings.Contains(err.Error(), "context deadline exceeded") {
		fmt.Println("Error type: context deadline exceeded")
	} else {
		fmt.Printf("Error type: %v\n", err)
	}

	// Output:
	// Error type: context deadline exceeded
}

// Example_databaseConnection demonstrates retrying a database connection.
func Example_databaseConnection() {
	ctx := context.Background()

	config := &retry.Config{
		MaxAttempts: 5,
		Strategy: retry.NewExponentialBackoff(retry.ExponentialBackoffConfig{
			Min:    500 * time.Millisecond,
			Max:    10 * time.Second,
			Factor: 2.0,
			Jitter: true,
		}),
		OnRetry: func(attempt int, err error) {
			log.Printf("Database connection attempt %d failed: %v", attempt+1, err)
		},
	}

	var db interface{} // Simulated database connection

	err := retry.Do(ctx, config, func(ctx context.Context) error {
		// Simulate database connection
		// db, err := sql.Open("postgres", dsn)
		// if err != nil {
		//     return err
		// }
		// return db.PingContext(ctx)

		// For this example, simulate success
		db = "connected"
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Printf("Database connected: %v\n", db != nil)

	// Output:
	// Database connected: true
}

// Example_httpRequest demonstrates retrying an HTTP request.
func Example_httpRequest() {
	ctx := context.Background()

	config := &retry.Config{
		MaxAttempts: 3,
		Strategy:    retry.NewDefaultExponentialBackoff(),
		OnRetry: func(attempt int, err error) {
			fmt.Printf("Request retry %d: %v\n", attempt+1, err)
		},
	}

	requestCount := 0
	err := retry.Do(ctx, config, func(ctx context.Context) error {
		requestCount++
		// Simulate HTTP request
		// resp, err := http.Get("https://api.example.com/data")
		// if err != nil {
		//     return err
		// }
		// if resp.StatusCode >= 500 {
		//     return fmt.Errorf("server error: %d", resp.StatusCode)
		// }
		// return nil

		if requestCount < 2 {
			return errors.New("503 Service Unavailable")
		}
		fmt.Println("Request successful")
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Output will show:
	// Request retry 1: 503 Service Unavailable
	// Request successful
}

// Example_noRetries demonstrates executing without retries.
func Example_noRetries() {
	ctx := context.Background()

	// MaxAttempts = 0 means no retries, just execute once
	config := &retry.Config{
		MaxAttempts: 0,
		Strategy:    retry.NewDefaultExponentialBackoff(),
	}

	err := retry.Do(ctx, config, func(ctx context.Context) error {
		return errors.New("immediate failure")
	})

	fmt.Printf("Error: %v\n", err)

	// Output:
	// Error: immediate failure
}
