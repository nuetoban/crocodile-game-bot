package main

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// RateLimiter -
type RateLimiter struct {
	pool *redis.Pool
}

// NewRateLimiter returns new instance of RateLimiter
func NewRateLimiter(pool *redis.Pool) *RateLimiter {
	return &RateLimiter{pool: pool}
}

// Limit -
func (r *RateLimiter) Limit(token int64, onsuccess, onFirstFailure, onFailure func() error) error {
	conn := r.pool.Get()
	defer conn.Close()

	minute := time.Now().Minute()
	key := fmt.Sprintf("rate/%d/%d", token, minute)
	resp, err := conn.Do("INCR", key)
	if err != nil {
		panic(err)
	}

	if count, ok := resp.(int64); ok {
		if count == 1 {
			conn.Do("EXPIRE", key, "60")
		}
		if count < 11 {
			err := onsuccess()
			return err
		}
		if count == 11 {
			log.Infof("RateLimiter: Limit: first time limit exceeded for token (%d), tries: %d, minute: %d", token, count, minute)
			err := onFirstFailure()
			return err
		}
		log.Infof("RateLimiter: Limit: limit exceeded for token (%d), tries: %d, minute: %d", token, count, minute)
		onFailure()
		return err
	}

	return fmt.Errorf("RateLimiter: Limit: Redis returned wrong type: %T", resp)
}
