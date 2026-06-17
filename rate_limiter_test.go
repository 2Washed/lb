package main

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenBucket_tryAllow_should_return_error_when_no_tokens(t *testing.T) {
	tokenBucket := newTestTokenBucket(float64(.5), time.Now(), 10)
	rate := 10

	err := tokenBucket.TryAllow(rate)
	if err == nil {
		t.Errorf("expected error but got nil")
	}

	errMsg, expectedErrMsg := err.Error(), "insufficient tokens"
	if errMsg != expectedErrMsg {
		t.Errorf("expected error to be %s but got %s", expectedErrMsg, errMsg)
	}
}

func TestTokenBucket_tryAllow_should_update_token_count(t *testing.T) {
	lastRefillTime := time.Now().Add(-1 * time.Second)
	tokenBucket := newTestTokenBucket(float64(.5), lastRefillTime, 10)
	rate := 10

	err := tokenBucket.TryAllow(rate)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	expectedTokenCount := float64(rate) - 1
	epsilon := float64(0.0001)

	if math.Abs(tokenBucket.tokens-expectedTokenCount) > epsilon {
		t.Errorf("expected %f to equal %f", tokenBucket.tokens, expectedTokenCount)
	}
}

func TestRateLimiter_Hit_should_return_err_if_rate_limit_reached(t *testing.T) {
	rate, burstSeconds := 10, 1
	expiry, _ := time.ParseDuration("10s")
	rateLimiter := NewRateLimiter(rate, burstSeconds, expiry)
	userId := "user"

	for i := 0; i < rate; i++ {
		rateLimiter.Hit(userId)
	}

	err := rateLimiter.Hit(userId)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	errMsg, expectedErrMsg := err.Error(), fmt.Sprintf("rate limit reached for %s: %s", userId, "insufficient tokens")
	if errMsg != expectedErrMsg {
		t.Errorf("invalid error message, expected: '%s' got '%s'", expectedErrMsg, errMsg)
	}
}

func TestRateLimiter_Hit_concurrency(t *testing.T) {
	rate, burstSeconds := 100, 1
	expiry, _ := time.ParseDuration("10s")
	rateLimiter := NewRateLimiter(rate, burstSeconds, expiry)

	var wg sync.WaitGroup
	var allowed atomic.Int32
	var rejected atomic.Int32

	requestCount := 2000
	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := rateLimiter.Hit("127.0.0.1"); err == nil {
				allowed.Add(1)
			} else {
				rejected.Add(1)
			}
		}()
	}
	wg.Wait()

	if allowed.Load() > 100 {
		t.Errorf("allowed %d requests, expected at most 1000", allowed.Load())
	}
	if allowed.Load()+rejected.Load() != 2000 {
		t.Errorf("total requests mismatch")
	}
}

func newTestTokenBucket(tokens float64, lastRefill time.Time, maxTokens int) *TokenBucket {
	return &TokenBucket{
		tokens:     tokens,
		lastRefill: lastRefill,
		maxTokens:  maxTokens,
	}
}
