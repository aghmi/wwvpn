package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wwvpn/wwvpn/internal/api/dto"
	"github.com/wwvpn/wwvpn/internal/api/middleware"
	"github.com/wwvpn/wwvpn/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// @Summary Authenticate device
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.DeviceAuthRequest true "Device ID"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/device [post]
func (h *AuthHandler) DeviceAuth(c *gin.Context) {
	var req dto.DeviceAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, expiresAt, device, err := h.authService.Authenticate(c.Request.Context(), req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		Token:    token,
		ExpireAt: expiresAt,
		DeviceID: device.DeviceID,
	})
}

// @Summary Delete account
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/account [delete]
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	deviceUUID, _ := c.Get(middleware.ContextDeviceUUID)

	if err := h.authService.DeleteDevice(c.Request.Context(), deviceUUID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
