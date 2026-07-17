package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/eci4ever/dc-go/pkg/response"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

var incrementScript = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
return {count, ttl}
`)

type Result struct {
	Count      int64
	RetryAfter time.Duration
}

type Store interface {
	Increment(context.Context, string, time.Duration) (Result, error)
}

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

func (s *RedisStore) Increment(ctx context.Context, key string, window time.Duration) (Result, error) {
	values, err := incrementScript.Run(ctx, s.client, []string{key}, window.Milliseconds()).Int64Slice()
	if err != nil {
		return Result{}, err
	}
	if len(values) != 2 {
		return Result{}, errors.New("invalid rate limit response")
	}
	retryAfter := time.Duration(values[1]) * time.Millisecond
	if retryAfter < 0 {
		retryAfter = window
	}
	return Result{Count: values[0], RetryAfter: retryAfter}, nil
}

type KeyFunc func(*fiber.Ctx) string

type Policy struct {
	Name   string
	Limit  int64
	Window time.Duration
	Key    KeyFunc
}

func Middleware(store Store, policy Policy) fiber.Handler {
	return func(c *fiber.Ctx) error {
		identifier := policy.Key(c)
		result, err := store.Increment(c.UserContext(), redisKey(policy.Name, identifier), policy.Window)
		if err != nil {
			slog.Error("rate limiter unavailable", "policy", policy.Name, "error", err)
			return c.Status(fiber.StatusServiceUnavailable).JSON(response.Error("service temporarily unavailable"))
		}

		remaining := max(policy.Limit-result.Count, 0)
		c.Set("RateLimit-Limit", strconv.FormatInt(policy.Limit, 10))
		c.Set("RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Set("RateLimit-Reset", strconv.FormatInt(resetSeconds(result.RetryAfter), 10))
		if result.Count <= policy.Limit {
			return c.Next()
		}

		retryAfter := resetSeconds(result.RetryAfter)
		c.Set("Retry-After", strconv.FormatInt(retryAfter, 10))
		return c.Status(fiber.StatusTooManyRequests).JSON(response.Error("too many requests"))
	}
}

func ByIP(c *fiber.Ctx) string {
	return c.IP()
}

func ByUser(c *fiber.Ctx) string {
	if userID, ok := c.Locals("user_id").(string); ok && userID != "" {
		return userID
	}
	return c.IP()
}

func ByParam(name string) KeyFunc {
	return func(c *fiber.Ctx) string {
		if value := c.Params(name); value != "" {
			return value
		}
		return c.IP()
	}
}

func redisKey(policy, identifier string) string {
	digest := sha256.Sum256([]byte(identifier))
	return "dc-go:ratelimit:" + policy + ":" + hex.EncodeToString(digest[:16])
}

func resetSeconds(duration time.Duration) int64 {
	seconds := int64((duration + time.Second - 1) / time.Second)
	return max(seconds, 1)
}
