// Package middleware: basic in-memory rate limiting middleware.
//
// This is a pragmatic production baseline for API abuse prevention:
// - token bucket by identifier (user or source IP)
// - per-method configuration through env vars
// - X-RateLimit headers (remaining/reset)
//
// Env configuration:
//
//	BP_RATE_LIMIT_ENABLED=1 (default off)
//	BP_RATE_LIMIT_RPS=30
//	BP_RATE_LIMIT_BURST=60
package middleware

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimit configures global settings.
type RateLimit struct {
	rate      rate.Limit
	burst     int
	enabled   bool
	discarded map[string]struct{}
}

var (
	rateLimitCfg = func() *RateLimit {
		rps := 30
		burst := 60
		if v := strings.TrimSpace(os.Getenv("BP_RATE_LIMIT_RPS")); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				rps = parsed
			}
		}
		if v := strings.TrimSpace(os.Getenv("BP_RATE_LIMIT_BURST")); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				burst = parsed
			}
		}
		enabled := os.Getenv("BP_RATE_LIMIT_ENABLED") == "1"
		return &RateLimit{rate.Limit(rps), burst, enabled, make(map[string]struct{})}
	}()

	limiterMux sync.Mutex
	limiters   = make(map[string]*rate.Limiter)
)

// RateLimitProtect enforces token bucket limits for each identifier.
// The middleware is opt-in via BP_RATE_LIMIT_ENABLED=1 for safer staged rollout.
func RateLimitProtect() gin.HandlerFunc {
	cfg := rateLimitCfg
	if !cfg.enabled {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		id := rateLimitIdentifier(c)
		limiter := getLimiter(id, cfg)
		if limiter.Allow() {
			c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.burst))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(limiter.Burst()-1))
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.burst))
		c.Header("Retry-After", "1")
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": "rate limit exceeded",
			"code":  "RATE_LIMIT",
		})
	}
}

func getLimiter(id string, cfg *RateLimit) *rate.Limiter {
	limiterMux.Lock()
	defer limiterMux.Unlock()
	if limiter, ok := limiters[id]; ok {
		return limiter
	}
	limiter := rate.NewLimiter(cfg.rate, cfg.burst)
	limiters[id] = limiter
	return limiter
}

func rateLimitIdentifier(c *gin.Context) string {
	if id, ok := c.Get("user_uuid"); ok {
		if s, ok := id.(string); ok && s != "" {
			return "u:" + s
		}
	}
	return "ip:" + c.ClientIP()
}
