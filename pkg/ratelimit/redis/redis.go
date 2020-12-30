package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ratelimit/pkg/ratelimit"
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
	ctx, cancel := context.WithTimeout(context.Background(), ratelimiter.frequency)
	defer cancel()
	data, _ := ratelimiter.client.Get(ctx, cacheKey).Result()
	json.Unmarshal([]byte(data), &cacheItem)
	if time.Now().After(cacheItem.ExpiredAt) {
		cacheItem = _CacheItem{
			Tokens:    ratelimiter.limit,
			ExpiredAt: time.Now().Add(ratelimiter.frequency),
		}
	}
	defer func() {
		data, err := json.Marshal(cacheItem)
		if err != nil {
			ratelimiter.client.Del(ctx, cacheKey)
			return
		}
		err = ratelimiter.client.Set(ctx, cacheKey, string(data), ratelimiter.frequency).Err()
		if err != nil {
			ratelimiter.client.Del(ctx, cacheKey)
		}
	}()

	if cacheItem.Tokens < n {
		return nil, errors.New("not enough tokens")
	}

	cacheItem.Tokens--
	ratelimitTokens := make([]ratelimit.Token, n)
	for i := int64(0); i < n; i++ {
		ratelimitTokens[i] = &Token{
			expiredAt: cacheItem.ExpiredAt,
			number:    ratelimiter.limit - cacheItem.Tokens + i,
		}
	}
	return ratelimitTokens, nil
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
