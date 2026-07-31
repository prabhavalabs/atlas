package adminapi

import (
	"sync"
	"time"
)

type attemptWindow struct {
	count       int
	windowStart time.Time
}

type attemptLimiter struct {
	mutex  sync.Mutex
	items  map[string]attemptWindow
	limit  int
	window time.Duration
}

func newAttemptLimiter(limit int, window time.Duration) *attemptLimiter {
	return &attemptLimiter{items: make(map[string]attemptWindow), limit: limit, window: window}
}

func (limiter *attemptLimiter) Allow(key string, now time.Time) bool {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()
	item := limiter.items[key]
	if item.windowStart.IsZero() || now.Sub(item.windowStart) >= limiter.window {
		limiter.items[key] = attemptWindow{count: 1, windowStart: now}
		return true
	}
	if item.count >= limiter.limit {
		return false
	}
	item.count++
	limiter.items[key] = item
	return true
}

func (limiter *attemptLimiter) Reset(key string) {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()
	delete(limiter.items, key)
}
