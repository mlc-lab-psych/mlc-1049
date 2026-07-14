package redis

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/WeatherGod3218/mlc-project-template/internal/logging"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const MILLISECONDS = 1000

var RatemlimitKey = os.Getenv("TRAIL_NAME") + ":ratelimit"

var middlewareScript = redis.NewScript(`
	local key = KEYS[1]
	local now = tonumber(ARGV[1])
	local rate = tonumber(ARGV[2])
	local capacity = tonumber(ARGV[3])
	local ttl = tonumber(ARGV[4])

	local bucketData = redis.call("HMGET", key, "tokens", "last_refill")
	local tokens = tonumber(bucketData[1]) or capacity
	local lastRefill = tonumber(bucketData[2]) or now

	local elapsed = math.max(0, (now - lastRefill) / 1000)
	tokens = math.min(capacity, tokens + (elapsed * rate))

	if tokens < 1 then
		redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
		redis.call("PEXPIRE", key, ttl)
		return {0, tokens}
	end

	tokens = tokens - 1
	redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
	redis.call("PEXPIRE", key, ttl)
	return {1, tokens}
`)

type TokenBucket struct {
	client   *redis.Client
	rate     float64
	capacity float64
}

func NewTokenBucket(rate float64, capacity float64) *TokenBucket {
	return &TokenBucket{client: client, rate: rate, capacity: capacity}
}

func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, int64, error) {
	now := time.Now().UnixMilli()
	ipKey := fmt.Sprintf("%s%s", RatemlimitKey, key)
	ttlMs := int64((tb.capacity / tb.rate) * MILLISECONDS * 2)

	result, err := middlewareScript.Run(ctx, tb.client, []string{ipKey}, now, tb.rate, tb.capacity, ttlMs).Int64Slice()
	if err != nil {
		return false, 0, fmt.Errorf("redis script: %w", err)
	}

	return (result[0] == 1), result[1], nil
}

func RedisRateLimiter(rate float64, capacity float64) gin.HandlerFunc {

	if client == nil {
		logging.Logger.WithFields(logrus.Fields{"module": "api", "method": "RedisRateLimiter"}).Warn("Redis cache was not initalized, proceeding with no rate limit!")

		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := NewTokenBucket(rate, capacity)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		allowed, tokens, err := limiter.Allow(c, ip)

		if err != nil {
			logging.Logger.WithFields(logrus.Fields{"error": err, "module": "api", "method": "RedisRateLimiter"}).Warn("Failure in the redis cache")
		} else if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests!",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", capacity))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%v", tokens))

		c.Next()
	}
}
