package counter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SlidingWindow implements a Redis-based counter for alert rate evaluation.
type SlidingWindow struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *SlidingWindow {
	return &SlidingWindow{rdb: rdb}
}

// Increment increments the counter for a rule and returns the new count.
// TTL is set to windowSeconds on first increment.
func (sw *SlidingWindow) Increment(ctx context.Context, ruleID, projectID string, windowSeconds int) (int64, error) {
	key := fmt.Sprintf("alert_counter:%s:%s", ruleID, projectID)
	count, err := sw.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		sw.rdb.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)
	}
	return count, nil
}

// HasCooldown returns true if a cooldown is active for this rule.
func (sw *SlidingWindow) HasCooldown(ctx context.Context, ruleID, projectID string) (bool, error) {
	key := fmt.Sprintf("cooldown:%s:%s", ruleID, projectID)
	n, err := sw.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// SetCooldown activates a cooldown for the given rule.
func (sw *SlidingWindow) SetCooldown(ctx context.Context, ruleID, projectID string, cooldownSeconds int) error {
	key := fmt.Sprintf("cooldown:%s:%s", ruleID, projectID)
	return sw.rdb.Set(ctx, key, 1, time.Duration(cooldownSeconds)*time.Second).Err()
}
