package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wwvpn/wwvpn/internal/api/dto"
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

	plan := sub.Plan
	if sub.ProductID != "" {
		plan = sub.ProductID
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     sub.Status,
		"plan":       plan,
		"expires_at": sub.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *SubscriptionHandler) VerifyReceipt(c *gin.Context) {
	deviceUUIDStr, _ := c.Get(middleware.ContextDeviceUUID)
	deviceUUID, err := uuid.Parse(deviceUUIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	var req dto.VerifyReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, httpStatus, vErr := h.subService.VerifyReceipt(c.Request.Context(), deviceUUID, req.Platform, req.Receipt, req.ProductID)
	if vErr != nil {
		if httpStatus == 422 {
			c.JSON(httpStatus, gin.H{
				"status": "invalid",
				"plan":   "",
				"error":  vErr.Error(),
			})
			return
		}
		c.JSON(httpStatus, gin.H{"error": vErr.Error()})
		return
	}

	c.JSON(httpStatus, resp)
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
