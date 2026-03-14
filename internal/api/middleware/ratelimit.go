package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	window   time.Duration
}

type visitor struct {
	count    int
	lastSeen time.Time
}

func RateLimit(requestsPerWindow int, window time.Duration) gin.HandlerFunc {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     requestsPerWindow,
		window:   window,
	}

	go rl.cleanup()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists || time.Since(v.lastSeen) > rl.window {
			rl.visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
			rl.mu.Unlock()
			c.Next()
			return
		}

		v.count++
		v.lastSeen = time.Now()

		if v.count > rl.rate {
			rl.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		rl.mu.Unlock()
		c.Next()
	}
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
