package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wwvpn/wwvpn/internal/api/middleware"
	"github.com/wwvpn/wwvpn/internal/service"
)

type SubscriptionHandler struct {
	subService *service.SubscriptionService
}

func NewSubscriptionHandler(subService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subService: subService}
}

// @Summary Get subscription status
// @Tags subscription
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /subscription/status [get]
func (h *SubscriptionHandler) Status(c *gin.Context) {
	deviceUUIDStr, _ := c.Get(middleware.ContextDeviceUUID)
	deviceUUID, err := uuid.Parse(deviceUUIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	sub, err := h.subService.GetActive(c.Request.Context(), deviceUUID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "none", "plan": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     sub.Status,
		"plan":       sub.Plan,
		"expires_at": sub.ExpiresAt,
	})
}

// @Summary RevenueCat webhook
// @Tags subscription
// @Accept json
// @Produce json
// @Success 501 {object} map[string]string
// @Router /webhook/revenuecat [post]
func (h *SubscriptionHandler) RevenueCatWebhook(c *gin.Context) {
	// TODO: implement in Phase 4
	c.JSON(http.StatusNotImplemented, gin.H{"error": "webhook not yet implemented"})
}
