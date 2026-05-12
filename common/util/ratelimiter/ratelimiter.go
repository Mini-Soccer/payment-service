package util

import (
	"context"
	"fmt"
	"time"

	"payment-service/domain/dto"

	"github.com/redis/go-redis/v9"
)

type RatelimiterUtil struct {
	redis      *redis.Client
	maxRequest int64
	duration   time.Duration
}

type IRatelimiterUtil interface {
	IsAllowed(ctx context.Context, auth *dto.User) bool
}

func NewRatelimiterUtil(redis *redis.Client, maxRequest int64, duration time.Duration) IRatelimiterUtil {
	return &RatelimiterUtil{
		redis:      redis,
		maxRequest: maxRequest,
		duration:   duration,
	}
}

func (u *RatelimiterUtil) IsAllowed(ctx context.Context, auth *dto.User) bool {
	key := fmt.Sprintf("ratelimit:field-service:%s", auth.UUID)

	increment, err := u.redis.Incr(ctx, key).Result()
	if err != nil {
		fmt.Println("Error incrementing:", err)
		return false
	}

	if increment == 1 {
		err := u.redis.Expire(ctx, key, u.duration).Err()
		if err != nil {
			fmt.Println("Error setting expiration:", err)
			return false
		}
	}

	return increment <= u.maxRequest
}
