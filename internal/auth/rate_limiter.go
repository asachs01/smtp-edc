package auth

import (
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	maxAttempts int
	window      time.Duration
	buckets     map[string]*Bucket
	mu          sync.RWMutex
	cleanupTick *time.Ticker
	done        chan bool
}

// Bucket represents a token bucket for rate limiting
type Bucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxAttempts int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		maxAttempts: maxAttempts,
		window:      window,
		buckets:     make(map[string]*Bucket),
		cleanupTick: time.NewTicker(window),
		done:        make(chan bool),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if an attempt is allowed for the given key (e.g., IP address)
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &Bucket{
			tokens:     rl.maxAttempts,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()

	// Refill tokens based on elapsed time
	elapsed := now.Sub(bucket.lastRefill)
	if elapsed >= rl.window {
		bucket.tokens = rl.maxAttempts
		bucket.lastRefill = now
	} else {
		// Gradual refill: add tokens proportional to elapsed time
		tokensToAdd := int(float64(rl.maxAttempts) * elapsed.Seconds() / rl.window.Seconds())
		bucket.tokens = min(rl.maxAttempts, bucket.tokens+tokensToAdd)
		if tokensToAdd > 0 {
			bucket.lastRefill = now
		}
	}

	// Check if we have tokens available
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// GetRemainingAttempts returns the number of remaining attempts for a key
func (rl *RateLimiter) GetRemainingAttempts(key string) int {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		return rl.maxAttempts
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)

	if elapsed >= rl.window {
		return rl.maxAttempts
	}

	// Calculate tokens based on elapsed time
	tokensToAdd := int(float64(rl.maxAttempts) * elapsed.Seconds() / rl.window.Seconds())
	return min(rl.maxAttempts, bucket.tokens+tokensToAdd)
}

// GetResetTime returns when the bucket will be fully reset for a key
func (rl *RateLimiter) GetResetTime(key string) time.Time {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		return time.Now()
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	return bucket.lastRefill.Add(rl.window)
}

// Reset resets the rate limit for a specific key
func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if bucket, exists := rl.buckets[key]; exists {
		bucket.mu.Lock()
		bucket.tokens = rl.maxAttempts
		bucket.lastRefill = time.Now()
		bucket.mu.Unlock()
	}
}

// Clear removes all rate limit data for a specific key
func (rl *RateLimiter) Clear(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.buckets, key)
}

// ClearAll removes all rate limit data
func (rl *RateLimiter) ClearAll() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.buckets = make(map[string]*Bucket)
}

// GetStats returns rate limiting statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["max_attempts"] = rl.maxAttempts
	stats["window_seconds"] = rl.window.Seconds()
	stats["active_buckets"] = len(rl.buckets)

	// Count buckets by status
	fullBuckets := 0
	emptyBuckets := 0
	partialBuckets := 0

	for _, bucket := range rl.buckets {
		bucket.mu.Lock()
		tokens := bucket.tokens
		bucket.mu.Unlock()

		switch {
		case tokens == rl.maxAttempts:
			fullBuckets++
		case tokens == 0:
			emptyBuckets++
		default:
			partialBuckets++
		}
	}

	stats["full_buckets"] = fullBuckets
	stats["empty_buckets"] = emptyBuckets
	stats["partial_buckets"] = partialBuckets

	return stats
}

// SetLimits updates the rate limiting parameters
func (rl *RateLimiter) SetLimits(maxAttempts int, window time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.maxAttempts = maxAttempts
	rl.window = window

	// Reset cleanup ticker
	rl.cleanupTick.Stop()
	rl.cleanupTick = time.NewTicker(window)
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.cleanupTick.C:
			rl.cleanupExpiredBuckets()
		case <-rl.done:
			rl.cleanupTick.Stop()
			return
		}
	}
}

// cleanupExpiredBuckets removes buckets that haven't been used recently
func (rl *RateLimiter) cleanupExpiredBuckets() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, bucket := range rl.buckets {
		bucket.mu.Lock()
		lastRefill := bucket.lastRefill
		tokens := bucket.tokens
		bucket.mu.Unlock()

		// Remove buckets that are full and haven't been used for 2x the window
		if tokens == rl.maxAttempts && now.Sub(lastRefill) > 2*rl.window {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(rl.buckets, key)
	}
}

// Stop stops the rate limiter and cleanup goroutine
func (rl *RateLimiter) Stop() {
	select {
	case rl.done <- true:
	default:
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
