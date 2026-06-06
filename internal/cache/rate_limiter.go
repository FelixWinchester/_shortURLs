package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client  *redis.Client
	limit   int
	window  time.Duration
}

func NewRateLimiter(client *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{client: client, limit: limit, window: window}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixMilli()
	windowStart := now - rl.window.Milliseconds()
	redisKey := fmt.Sprintf("ratelimit:%s", key)

	pipe := rl.client.Pipeline()

	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart))

	count := pipe.ZCard(ctx, redisKey)

	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  float64(now),
		Member: now,
	})

	pipe.Expire(ctx, redisKey, rl.window+time.Second)

	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("rate limiter: %w", err)
	}

	if count.Val() >= int64(rl.limit) {
		return false, nil
	}

	return true, nil
}
