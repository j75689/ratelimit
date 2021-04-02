package ratelimit

import (
	"time"
)

// Ratelimiter ...
type Ratelimiter interface {
	SetFrequency(time.Duration) error
	SetLimit(int64) error
	WithRedis(RedisOption) error
	Acquire(interface{}) (Token, error)
	AcquireN(interface{}, int64) ([]Token, error)
}

// Token ...
type Token interface {
	Expire() (time.Time, error)
	Number() int64
}

// An Option configures a Ratelimiter
type Option interface {
	Apply(Ratelimiter) error
}

// OptionFunc is a function that configures a Ratelimiter
type OptionFunc func(Ratelimiter) error

// Apply is a function that set value to Ratelimiter
func (f OptionFunc) Apply(ratelimiter Ratelimiter) error {
	return f(ratelimiter)
}

func SetFrequency(frequency time.Duration) Option {
	return OptionFunc(func(r Ratelimiter) error {
		return r.SetFrequency(frequency)
	})
}

type RedisOption struct {
	Host         string
	Port         uint
	DB           int
	Password     string
	MinIdleConns int
	MaxPoolSize  int
	DialTimeout  time.Duration
	MaxRetry     int
}

func WithRedis(redis RedisOption) Option {
	return OptionFunc(func(r Ratelimiter) error {
		return r.WithRedis(redis)
	})
}
