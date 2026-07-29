package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

// InternalSecretGuard protects server-to-server internal API endpoints.
// Requires X-Internal-Secret header matching the configured shared secret.
// Used by TAI Protocol backend for usage reconciliation (Phase 0)
// and wallet-auth compute access (Phase 1).
func InternalSecretGuard(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "internal API not configured",
			})
			return
		}

		provided := c.GetHeader("X-Internal-Secret")
		if provided == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing X-Internal-Secret header",
			})
			return
		}

		if subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "invalid internal secret",
			})
			return
		}

		c.Next()
	}
}
