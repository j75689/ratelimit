package wireset

import (
	"ratelimit/internal/config"
	"ratelimit/pkg/ratelimit"
	"ratelimit/pkg/ratelimit/supported"
)

func InitRateLimiter(config config.Config) (ratelimit.Ratelimiter, error) {
	return supported.New(
		config.RateLimit.Driver,
		config.RateLimit.Limit,
		ratelimit.SetFrequency(config.RateLimit.Frequency),
		ratelimit.WithRedis(ratelimit.RedisOption{
			Host:         config.RateLimit.RedisOption.Host,
			Port:         config.RateLimit.RedisOption.Port,
			DB:           config.RateLimit.RedisOption.DB,
			Password:     config.RateLimit.RedisOption.Password,
			MinIdleConns: config.RateLimit.RedisOption.MinIdleConns,
			MaxPoolSize:  config.RateLimit.RedisOption.MaxPoolSize,
			DialTimeout:  config.RateLimit.RedisOption.DialTimeout,
		}),
	)
}
