package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ratelimit/pkg/ratelimit"
	ratelimiterErrs "ratelimit/pkg/ratelimit/errors"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	_ ratelimit.Ratelimiter = (*RedisRateLimiter)(nil)
	_ ratelimit.Token       = (*Token)(nil)
)

func NewRedisRateLimiter(limit int64) (ratelimit.Ratelimiter, error) {
	return &RedisRateLimiter{
		frequency: time.Second,
		limit:     limit,
	}, nil
}

type RedisRateLimiter struct {
	sync.Mutex
	client    *redis.Client
	frequency time.Duration
	limit     int64
	maxRetry  int
}

func (ratelimiter *RedisRateLimiter) SetFrequency(frequency time.Duration) error {
	ratelimiter.Lock()
	defer ratelimiter.Unlock()
	ratelimiter.frequency = frequency
	return nil
}
func (ratelimiter *RedisRateLimiter) SetLimit(limit int64) error {
	ratelimiter.Lock()
	defer ratelimiter.Unlock()
	ratelimiter.limit = limit
	return nil
}

func (ratelimiter *RedisRateLimiter) WithRedis(redisOption ratelimit.RedisOption) error {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", redisOption.Host, redisOption.Port),
		Password:     redisOption.Password,
		DB:           redisOption.DB,
		PoolSize:     redisOption.MaxPoolSize,
		MinIdleConns: redisOption.MinIdleConns,
	})
	ctx, cancel := context.WithTimeout(context.Background(), redisOption.DialTimeout)
	defer cancel()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return err
	}
	ratelimiter.client = client
	ratelimiter.maxRetry = redisOption.MaxRetry
	return nil
}

func (ratelimiter *RedisRateLimiter) Acquire(key interface{}) (ratelimit.Token, error) {
	tokens, err := ratelimiter.AcquireN(key, 1)
	if err != nil {
		return nil, err
	}
	return tokens[0], nil
}
func (ratelimiter *RedisRateLimiter) AcquireN(key interface{}, n int64) ([]ratelimit.Token, error) {
	ratelimiter.Lock()
	defer ratelimiter.Unlock()
	cacheKey := fmt.Sprint(key)
	cacheItem := _CacheItem{}
	ratelimitTokens := make([]ratelimit.Token, n)
	ctx, cancel := context.WithTimeout(context.Background(), ratelimiter.frequency)
	defer cancel()

	txf := func(tx *redis.Tx) error {
		defer tx.Close(ctx)
		data, _ := tx.Get(ctx, cacheKey).Result()
		json.Unmarshal([]byte(data), &cacheItem)
		if time.Now().After(cacheItem.ExpiredAt) {
			cacheItem = _CacheItem{
				Tokens:    ratelimiter.limit,
				ExpiredAt: time.Now().Add(ratelimiter.frequency),
			}
		}

		if cacheItem.Tokens < n {
			return ratelimiterErrs.ErrNotEnoughToken
		}

		cacheItem.Tokens--
		for i := int64(0); i < n; i++ {
			ratelimitTokens[i] = &Token{
				expiredAt: cacheItem.ExpiredAt,
				number:    ratelimiter.limit - cacheItem.Tokens + i,
			}
		}

		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			data, err := json.Marshal(cacheItem)
			if err != nil {
				return pipe.Del(ctx, cacheKey).Err()
			}
			err = pipe.Set(ctx, cacheKey, string(data), ratelimiter.frequency).Err()
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	}

	for i := 0; i < ratelimiter.maxRetry; i++ {
		err := ratelimiter.client.Watch(ctx, txf, cacheKey)
		if errors.Is(err, redis.TxFailedErr) {
			continue
		}
		return ratelimitTokens, err
	}

	return nil, ratelimiterErrs.ErrNotEnoughToken
}

type Token struct {
	number    int64
	expiredAt time.Time
}

func (token *Token) Expire() (time.Time, error) {
	return token.expiredAt, nil
}

func (token *Token) Number() int64 {
	return token.number
}

type _CacheItem struct {
	ExpiredAt time.Time `json:"expired_at"`
	Tokens    int64     `json:"tokens"`
}
