package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type identifier = string

type RateLimiter struct {
	buckets      map[identifier]*TokenBucket
	rate         int
	burstSeconds int
	ttlSeconds   int
	mu           sync.RWMutex
}

type TokenBucket struct {
	tokens     float64
	lastRefill time.Time
	maxTokens  int
	mu         sync.Mutex
}

func (tb *TokenBucket) Allow(rate int) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	requestTime := time.Now()
	lastRefill := tb.lastRefill

	elapsedSeconds := requestTime.Sub(lastRefill).Seconds()
	tokensToAdd := elapsedSeconds * float64(rate)

	tb.tokens = math.Min(float64(tb.maxTokens), tb.tokens+tokensToAdd)
	tb.lastRefill = requestTime

	if tb.tokens < 1 {
		return fmt.Errorf("insufficient tokens")
	}

	tb.tokens -= 1
	return nil
}

func (r *RateLimiter) Hit(id identifier) error {
	bucket := r.getOrCreateBucket(id)
	err := bucket.Allow(r.rate)
	if err != nil {
		return fmt.Errorf("rate limit reached for %s: %w", id, err)
	}
	return nil
}

func (r *RateLimiter) PurgeStale() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, bucket := range r.buckets {
		bucket.mu.Lock()
		if time.Since(bucket.lastRefill).Seconds() > 99 {
			delete(r.buckets, id)
		}
		bucket.mu.Unlock()
	}
}

func (r *RateLimiter) getOrCreateBucket(id identifier) *TokenBucket {
	r.mu.RLock()
	bucket, ok := r.buckets[id]
	r.mu.RUnlock()

	if ok {
		return bucket
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if bucket, ok = r.buckets[id]; ok { //just incase another goroutine created it between runlock and lock
		return bucket
	}

	bucket = &TokenBucket{
		tokens:     float64(r.rate * r.burstSeconds),
		lastRefill: time.Now(),
		maxTokens:  r.rate * r.burstSeconds,
	}

	r.buckets[id] = bucket
	return bucket
}

func NewRateLimiter(rate int, burstSeconds int, ttlSeconds int) *RateLimiter {
	return &RateLimiter{
		buckets:      make(map[identifier]*TokenBucket),
		rate:         rate,
		burstSeconds: burstSeconds,
		ttlSeconds:   ttlSeconds,
	}
}
