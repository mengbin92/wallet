package retry

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// RetryFunc is the function to be retried
type RetryFunc func() error

// Config holds retry configuration
type Config struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultConfig returns default retry configuration
func DefaultConfig() *Config {
	return &Config{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
	}
}

// WithExponentialBackoff retries the function with exponential backoff
func WithExponentialBackoff(fn RetryFunc, config *Config) error {
	if config == nil {
		config = DefaultConfig()
	}

	var lastErr error
	for i := 0; i <= config.MaxRetries; i++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't wait after the last attempt
		if i < config.MaxRetries {
			// Calculate delay with exponential backoff
			delay := config.BaseDelay * time.Duration(1<<uint(i))
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}

			// Wait before retry
			time.Sleep(delay)
		}
	}

	return errors.Wrapf(lastErr, "failed after %d retries", config.MaxRetries)
}

// WithContext retries the function with context and exponential backoff
func WithContext(ctx context.Context, fn func(context.Context) error, config *Config) error {
	if config == nil {
		config = DefaultConfig()
	}

	var lastErr error
	for i := 0; i <= config.MaxRetries; i++ {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't wait after the last attempt
		if i < config.MaxRetries {
			// Calculate delay with exponential backoff
			delay := config.BaseDelay * time.Duration(1<<uint(i))
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}

			// Wait with context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue to retry
			}
		}
	}

	return errors.Wrapf(lastErr, "failed after %d retries", config.MaxRetries)
}

// IsRetryable checks if an error is retryable
// You can customize this function based on your needs
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Add specific error types that should be retried
	// For example: network errors, temporary errors, etc.

	return true
}
