package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wwvpn/wwvpn/internal/service"
)

type ServerHandler struct {
	serverService *service.ServerService
}

func NewServerHandler(serverService *service.ServerService) *ServerHandler {
	return &ServerHandler{serverService: serverService}
}

// @Summary List active VPN servers
// @Tags servers
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.ServerListResponse
// @Failure 500 {object} map[string]string
// @Router /servers [get]
func (h *ServerHandler) ListServers(c *gin.Context) {
	resp, err := h.serverService.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch servers"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
