package ratelimiter

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
	expiry       time.Duration
	mu           sync.RWMutex
}

type TokenBucket struct {
	Tokens     float64
	LastRefill time.Time
	MaxTokens  int
	Mu         sync.Mutex
}

func (tb *TokenBucket) TryAllow(rate int) error {
	tb.Mu.Lock()
	defer tb.Mu.Unlock()

	requestTime := time.Now()
	lastRefill := tb.LastRefill

	elapsedSeconds := requestTime.Sub(lastRefill).Seconds()
	tokensToAdd := elapsedSeconds * float64(rate)

	tb.Tokens = math.Min(float64(tb.MaxTokens), tb.Tokens+tokensToAdd)
	tb.LastRefill = requestTime

	if tb.Tokens < 1 {
		return fmt.Errorf("insufficient tokens")
	}

	tb.Tokens -= 1
	return nil
}

func (r *RateLimiter) Hit(id identifier) error {
	bucket := r.getOrCreateBucket(id)
	err := bucket.TryAllow(r.rate)
	if err != nil {
		return fmt.Errorf("rate limit reached for %s: %v", id, err)
	}
	return nil
}

func (r *RateLimiter) PurgeStale() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, bucket := range r.buckets {
		bucket.Mu.Lock()
		if time.Since(bucket.LastRefill) > r.expiry {
			delete(r.buckets, id)
		}
		bucket.Mu.Unlock()
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
		Tokens:     float64(r.rate * r.burstSeconds),
		LastRefill: time.Now(),
		MaxTokens:  r.rate * r.burstSeconds,
	}

	r.buckets[id] = bucket
	return bucket
}

func NewRateLimiter(rate int, burstSeconds int, expiry time.Duration) *RateLimiter {
	rateLimiter := &RateLimiter{
		buckets:      make(map[identifier]*TokenBucket),
		rate:         rate,
		burstSeconds: burstSeconds,
		expiry:       expiry,
	}
	go func() {
		for {
			time.Sleep(rateLimiter.expiry)
			rateLimiter.PurgeStale()
		}
	}()
	return rateLimiter
}
