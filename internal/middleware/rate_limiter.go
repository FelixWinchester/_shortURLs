package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/winchester/shorturls/internal/cache"
)

const rateLimit = 20
const rateWindow = time.Minute

func RateLimiter(rl *cache.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		allowed, err := rl.Allow(c.Request.Context(), ip)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rate limiter error"})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}

		c.Next()
	}
}
