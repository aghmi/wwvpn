package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wwvpn/wwvpn/internal/service"
)

func RequireSubscription(subService *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceUUIDStr, exists := c.Get(ContextDeviceUUID)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		deviceUUID, err := uuid.Parse(deviceUUIDStr.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
			return
		}

		canConnect, err := subService.CanConnect(c.Request.Context(), deviceUUID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no active subscription"})
			return
		}

		if !canConnect {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "subscription limit reached"})
			return
		}

		c.Next()
	}
}
