package supported

import (
	"fmt"
	"ratelimit/pkg/ratelimit"
	"ratelimit/pkg/ratelimit/memory"
	"ratelimit/pkg/ratelimit/redis"
)

// SupportedDriver ...
type SupportedDriver string

type _InitFunc func(limit int64) (ratelimit.Ratelimiter, error)

// ...
const (
	MEMORY SupportedDriver = "memory"
	REDIS  SupportedDriver = "redis"
)

var _SuppotredDrivers = map[SupportedDriver]_InitFunc{
	MEMORY: memory.NewMemoryRateLimiter,
	REDIS:  redis.NewRedisRateLimiter,
}

// New returns a Ratelimiter
func New(driver SupportedDriver, limit int64, opts ...ratelimit.Option) (ratelimit.Ratelimiter, error) {
	var ratelimiter ratelimit.Ratelimiter
	f, ok := _SuppotredDrivers[driver]
	if !ok {
		return nil, fmt.Errorf("not support driver [%s]", driver)
	}

	ratelimiter, err := f(limit)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		if err := opt.Apply(ratelimiter); err != nil {
			return nil, err
		}
	}

	return ratelimiter, nil
}
