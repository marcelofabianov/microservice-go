package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDo_SuccessFirstAttempt(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		MaxAttempts: 3,
		Strategy:    NewDefaultExponentialBackoff(),
	}

	callCount := 0
	fn := func(ctx context.Context) error {
		callCount++
		return nil // Success on first attempt
	}

	err := Do(ctx, config, fn)

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount, "should execute only once on first success")
}

func TestDo_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		MaxAttempts: 3,
		Strategy:    NewConstantBackoff(10 * time.Millisecond),
	}

	callCount := 0
	fn := func(ctx context.Context) error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil // Success on 3rd attempt
	}

	err := Do(ctx, config, fn)

	assert.NoError(t, err)
	assert.Equal(t, 3, callCount, "should succeed on 3rd attempt")
}

func TestDo_MaxAttemptsReached(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		MaxAttempts: 3,
		Strategy:    NewConstantBackoff(10 * time.Millisecond),
	}

	callCount := 0
	fn := func(ctx context.Context) error {
		callCount++
		return errors.New("persistent error")
	}

	err := Do(ctx, config, fn)

	assert.ErrorIs(t, err, ErrMaxAttemptsReached, "should return ErrMaxAttemptsReached")
	assert.Equal(t, 4, callCount, "should call 1 initial + 3 retries = 4 times")
}

func TestDo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &Config{
		MaxAttempts: 10,
		Strategy:    NewConstantBackoff(50 * time.Millisecond),
	}

	callCount := 0
	fn := func(ctx context.Context) error {
		callCount++
		if callCount == 2 {
			cancel() // Cancel after second attempt
		}
		return errors.New("error")
	}

	err := Do(ctx, config, fn)

	assert.ErrorIs(t, err, context.Canceled, "should return context.Canceled")
	assert.GreaterOrEqual(t, callCount, 2, "should execute at least 2 attempts before cancellation")
}

func TestDo_NoRetries(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		MaxAttempts: 0, // No retries
		Strategy:    NewDefaultExponentialBackoff(),
	}

	callCount := 0
	testErr := errors.New("test error")
	fn := func(ctx context.Context) error {
		callCount++
		return testErr
	}

	err := Do(ctx, config, fn)

	assert.ErrorIs(t, err, testErr, "should return original error")
	assert.Equal(t, 1, callCount, "should execute only once with no retries")
}

func TestDo_OnRetryCallback(t *testing.T) {
	ctx := context.Background()

	retryCallbacks := []int{}
	retryErrors := []error{}

	config := &Config{
		MaxAttempts: 3,
		Strategy:    NewConstantBackoff(10 * time.Millisecond),
		OnRetry: func(attempt int, err error) {
			retryCallbacks = append(retryCallbacks, attempt)
			retryErrors = append(retryErrors, err)
		},
	}

	callCount := 0
	tempErr := errors.New("temporary error")

	fn := func(ctx context.Context) error {
		callCount++
		if callCount <= 2 {
			return tempErr
		}
		return nil
	}

	err := Do(ctx, config, fn)

	assert.NoError(t, err)
	assert.Equal(t, []int{0, 1}, retryCallbacks, "should call OnRetry with attempts 0 and 1")
	assert.Len(t, retryErrors, 2, "should capture 2 errors")
	for _, err := range retryErrors {
		assert.ErrorIs(t, err, tempErr, "should capture the temporary error")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errType error
	}{
		{
			name: "valid config",
			config: &Config{
				MaxAttempts: 3,
				Strategy:    NewDefaultExponentialBackoff(),
			},
			wantErr: false,
		},
		{
			name: "zero attempts is valid",
			config: &Config{
				MaxAttempts: 0,
				Strategy:    NewDefaultExponentialBackoff(),
			},
			wantErr: false,
		},
		{
			name: "negative attempts invalid",
			config: &Config{
				MaxAttempts: -1,
				Strategy:    NewDefaultExponentialBackoff(),
			},
			wantErr: true,
			errType: ErrInvalidConfig,
		},
		{
			name: "nil strategy invalid",
			config: &Config{
				MaxAttempts: 3,
				Strategy:    nil,
			},
			wantErr: true,
			errType: ErrInvalidConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				require.Error(t, err, "Validate() should return error")
				assert.ErrorIs(t, err, tt.errType, "should return expected error type")
			} else {
				assert.NoError(t, err, "Validate() should not return error")
			}
		})
	}
}

func TestExponentialBackoff_NextDelay(t *testing.T) {
	backoff := NewExponentialBackoff(ExponentialBackoffConfig{
		Min:    1 * time.Second,
		Max:    30 * time.Second,
		Factor: 2.0,
		Jitter: false, // Disable jitter for predictable testing
	})

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 0, want: 1 * time.Second},   // 1 * 2^0 = 1s
		{attempt: 1, want: 2 * time.Second},   // 1 * 2^1 = 2s
		{attempt: 2, want: 4 * time.Second},   // 1 * 2^2 = 4s
		{attempt: 3, want: 8 * time.Second},   // 1 * 2^3 = 8s
		{attempt: 4, want: 16 * time.Second},  // 1 * 2^4 = 16s
		{attempt: 5, want: 30 * time.Second},  // 1 * 2^5 = 32s, capped at 30s
		{attempt: 10, want: 30 * time.Second}, // Capped at max
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			got := backoff.NextDelay(tt.attempt)
			assert.Equal(t, tt.want, got, "NextDelay should return expected duration")
		})
	}
}

func TestExponentialBackoff_WithJitter(t *testing.T) {
	backoff := NewExponentialBackoff(ExponentialBackoffConfig{
		Min:    1 * time.Second,
		Max:    30 * time.Second,
		Factor: 2.0,
		Jitter: true,
	})

	// With jitter, delays should vary but stay within bounds
	baseDelay := 2 * time.Second // For attempt 1
	minExpected := time.Duration(float64(baseDelay) * 0.5)
	maxExpected := time.Duration(float64(baseDelay) * 1.5)

	for i := 0; i < 10; i++ {
		delay := backoff.NextDelay(1)
		assert.GreaterOrEqual(t, delay, minExpected, "delay should be >= 50% of base")
		assert.LessOrEqual(t, delay, maxExpected, "delay should be <= 150% of base")
	}
}

func TestExponentialBackoff_Defaults(t *testing.T) {
	backoff := NewDefaultExponentialBackoff()

	delay0 := backoff.NextDelay(0)

	// With jitter, should be between 0.5s and 1.5s
	assert.GreaterOrEqual(t, delay0, 500*time.Millisecond, "delay should be >= 500ms")
	assert.LessOrEqual(t, delay0, 2*time.Second, "delay should be <= 2s")
}

func TestExponentialBackoff_NegativeAttempt(t *testing.T) {
	backoff := NewDefaultExponentialBackoff()

	delay := backoff.NextDelay(-5)
	// Should treat negative as 0
	assert.GreaterOrEqual(t, delay, 500*time.Millisecond, "negative attempt should be treated as 0")
}

func TestConstantBackoff_NextDelay(t *testing.T) {
	backoff := NewConstantBackoff(5 * time.Second)

	for i := 0; i < 10; i++ {
		delay := backoff.NextDelay(i)
		assert.Equal(t, 5*time.Second, delay, "should always return constant delay")
	}
}

func TestConstantBackoff_DefaultsToOneSecond(t *testing.T) {
	backoff := NewConstantBackoff(0) // Invalid: 0 duration

	delay := backoff.NextDelay(0)
	assert.Equal(t, 1*time.Second, delay, "should default to 1 second for invalid input")
}

func TestLinearBackoff_NextDelay(t *testing.T) {
	backoff := NewLinearBackoff(2*time.Second, 10*time.Second)

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 0, want: 2 * time.Second},  // 2 * 1 = 2s
		{attempt: 1, want: 4 * time.Second},  // 2 * 2 = 4s
		{attempt: 2, want: 6 * time.Second},  // 2 * 3 = 6s
		{attempt: 3, want: 8 * time.Second},  // 2 * 4 = 8s
		{attempt: 4, want: 10 * time.Second}, // 2 * 5 = 10s (capped)
		{attempt: 5, want: 10 * time.Second}, // Capped at max
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			got := backoff.NextDelay(tt.attempt)
			assert.Equal(t, tt.want, got, "NextDelay should return expected duration")
		})
	}
}

func TestLinearBackoff_NegativeAttempt(t *testing.T) {
	backoff := NewLinearBackoff(2*time.Second, 10*time.Second)

	delay := backoff.NextDelay(-5)
	assert.Equal(t, 2*time.Second, delay, "negative attempt should be treated as 0")
}

func TestLinearBackoff_Defaults(t *testing.T) {
	// Test default validation
	backoff := NewLinearBackoff(0, 5*time.Second) // Invalid increment

	delay := backoff.NextDelay(0)
	assert.Equal(t, 1*time.Second, delay, "should default increment to 1s")

	// Test max < increment
	backoff2 := NewLinearBackoff(10*time.Second, 5*time.Second) // max < increment
	delay2 := backoff2.NextDelay(0)
	assert.Equal(t, 10*time.Second, delay2, "max should be adjusted to increment")
}

func TestThreadSafety(t *testing.T) {
	backoff := NewDefaultExponentialBackoff()

	// Run multiple goroutines calling NextDelay concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(attempt int) {
			for j := 0; j < 100; j++ {
				_ = backoff.NextDelay(attempt)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	// If we reach here without panicking, thread safety is verified
}

func TestDo_InvalidConfig(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "negative max attempts",
			config: &Config{
				MaxAttempts: -1,
				Strategy:    NewDefaultExponentialBackoff(),
			},
		},
		{
			name: "nil strategy",
			config: &Config{
				MaxAttempts: 3,
				Strategy:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Do(ctx, tt.config, func(ctx context.Context) error {
				return nil
			})

			assert.ErrorIs(t, err, ErrInvalidConfig, "should return ErrInvalidConfig")
		})
	}
}

func TestDo_ContextCanceledBeforeRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := &Config{
		MaxAttempts: 3,
		Strategy:    NewConstantBackoff(100 * time.Millisecond),
	}

	callCount := 0
	err := Do(ctx, config, func(ctx context.Context) error {
		callCount++
		return errors.New("test error")
	})

	assert.ErrorIs(t, err, context.Canceled, "should respect pre-canceled context")
	assert.Equal(t, 1, callCount, "should only execute initial attempt")
}

func TestBackoffStrategy_Reset(t *testing.T) {
	// Test that Reset() doesn't panic (it's a no-op for stateless strategies)
	strategies := []Strategy{
		NewDefaultExponentialBackoff(),
		NewConstantBackoff(1 * time.Second),
		NewLinearBackoff(1*time.Second, 10*time.Second),
	}

	for i, strategy := range strategies {
		t.Run(fmt.Sprintf("strategy_%d", i), func(t *testing.T) {
			assert.NotPanics(t, func() {
				strategy.Reset()
			}, "Reset should not panic")
		})
	}
}
