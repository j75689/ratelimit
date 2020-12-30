package memory

import (
	"errors"
	"fmt"
	"ratelimit/pkg/ratelimit"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	_ ratelimit.Ratelimiter = (*MemoryRateLimiter)(nil)
	_ ratelimit.Token       = (*Token)(nil)
)

func NewMemoryRateLimiter(limit int64) (ratelimit.Ratelimiter, error) {
	return &MemoryRateLimiter{
		cache:     cache.New(time.Second, time.Second),
		frequency: time.Second,
		limit:     limit,
	}, nil
}

type MemoryRateLimiter struct {
	sync.Mutex
	cache     *cache.Cache
	frequency time.Duration
	limit     int64
}

func (ratelimiter *MemoryRateLimiter) SetFrequency(frequency time.Duration) error {
	ratelimiter.Lock()
	defer ratelimiter.Unlock()
	ratelimiter.cache = cache.New(frequency, frequency)
	ratelimiter.frequency = frequency
	return nil
}

func (ratelimiter *MemoryRateLimiter) SetLimit(limit int64) error {
	ratelimiter.Lock()
	defer ratelimiter.Unlock()
	ratelimiter.limit = limit
	return nil
}

func (ratelimiter *MemoryRateLimiter) WithRedis(ratelimit.RedisOption) error {
	return nil
}

func (ratelimiter *MemoryRateLimiter) Acquire(key interface{}) (ratelimit.Token, error) {
	tokens, err := ratelimiter.AcquireN(key, 1)
	if err != nil {
		return nil, err
	}
	return tokens[0], nil
}

func (ratelimiter *MemoryRateLimiter) AcquireN(key interface{}, n int64) ([]ratelimit.Token, error) {
	ratelimiter.Lock()
	defer ratelimiter.Unlock()

	cacheKey := fmt.Sprint(key)
	cacheItem := _CacheItem{
		tokens:    ratelimiter.limit,
		expiredAt: time.Now().Add(ratelimiter.frequency),
	}
	tokensV, ok := ratelimiter.cache.Get(cacheKey)
	if ok {
		if v, ok := tokensV.(_CacheItem); ok {
			if time.Now().Before(v.expiredAt) {
				cacheItem = v
			}
		}
	}

	defer func() {
		ratelimiter.cache.Set(cacheKey, cacheItem, ratelimiter.frequency)
	}()

	if cacheItem.tokens < n {
		return nil, errors.New("not enough tokens")
	}

	cacheItem.tokens--
	ratelimitTokens := make([]ratelimit.Token, n)
	for i := int64(0); i < n; i++ {
		ratelimitTokens[i] = &Token{
			expiredAt: cacheItem.expiredAt,
			number:    ratelimiter.limit - cacheItem.tokens + i,
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
	expiredAt time.Time
	tokens    int64
}
