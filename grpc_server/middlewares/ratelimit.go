package middlewares

import (
	"cchoice/internal/ctx"
	"context"
	"fmt"

	"github.com/juju/ratelimit"
)

const (
	rate     = 32
	capacity = 32
)

type RateLimiter struct {
	Bucket *ratelimit.Bucket
}

func (rl *RateLimiter) Limit(_ context.Context) error {
	res := rl.Bucket.TakeAvailable(1)
	if res == 0 {
		return fmt.Errorf("Reached rate limit %d", rl.Bucket.Available())
	}
	return nil
}

func AddRateLimit(ctxGRPC *ctx.GRPCFlags) *RateLimiter {
	if !ctxGRPC.LogPayloadReceived {
		return nil
	}

	return &RateLimiter{
		Bucket: ratelimit.NewBucket(rate, int64(capacity)),
	}
}
