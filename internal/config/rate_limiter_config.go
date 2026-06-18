package config

type RateLimiterConfig struct {
	Rate         int      `json:"rate"`
	BurstSeconds int      `json:"burstSeconds"`
	Expiry       Duration `json:"expiry"`
}
