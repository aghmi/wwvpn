package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wwvpn/wwvpn/internal/crypto"
)

const ContextDeviceUUID = "device_uuid"

func Auth(jwtManager *crypto.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		claims, err := jwtManager.Validate(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextDeviceUUID, claims.DeviceUUID.String())
		c.Next()
	}
}

func AdminAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Admin-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing admin key"})
			return
		}

		if subtle.ConstantTimeCompare([]byte(key), []byte(apiKey)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid admin key"})
			return
		}

		c.Next()
	}
}
