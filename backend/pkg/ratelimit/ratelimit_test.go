package ratelimit

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

type fakeStore struct {
	result Result
	err    error
	keys   []string
}

func (s *fakeStore) Increment(_ context.Context, key string, _ time.Duration) (Result, error) {
	s.keys = append(s.keys, key)
	return s.result, s.err
}

func TestMiddlewareAllowsRequestAndSetsHeaders(t *testing.T) {
	store := &fakeStore{result: Result{Count: 2, RetryAfter: 30 * time.Second}}
	app := fiber.New()
	app.Get("/", Middleware(store, Policy{Name: "login", Limit: 5, Window: time.Minute, Key: ByIP}), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	res, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", res.StatusCode, fiber.StatusNoContent)
	}
	if got := res.Header.Get("RateLimit-Remaining"); got != "3" {
		t.Fatalf("remaining = %q, want 3", got)
	}
}

func TestMiddlewareRejectsExceededLimit(t *testing.T) {
	store := &fakeStore{result: Result{Count: 6, RetryAfter: 1500 * time.Millisecond}}
	app := fiber.New()
	app.Get("/", Middleware(store, Policy{Name: "login", Limit: 5, Window: time.Minute, Key: ByIP}), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	res, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", res.StatusCode, fiber.StatusTooManyRequests)
	}
	if got := res.Header.Get("Retry-After"); got != "2" {
		t.Fatalf("retry-after = %q, want 2", got)
	}
}

func TestMiddlewareFailsClosedWhenStoreIsUnavailable(t *testing.T) {
	store := &fakeStore{err: errors.New("Redis unavailable")}
	app := fiber.New()
	app.Get("/", Middleware(store, Policy{Name: "login", Limit: 5, Window: time.Minute, Key: ByIP}), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	res, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", res.StatusCode, fiber.StatusServiceUnavailable)
	}
}

func TestRedisKeyDoesNotExposeIdentifier(t *testing.T) {
	key := redisKey("login", "person@example.com")
	if key == "" || key == "dc-go:ratelimit:login:person@example.com" {
		t.Fatalf("identifier was not hashed: %q", key)
	}
}
