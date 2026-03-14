package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wwvpn/wwvpn/internal/api/dto"
	"github.com/wwvpn/wwvpn/internal/service"
)

type AdminHandler struct {
	serverService *service.ServerService
}

func NewAdminHandler(serverService *service.ServerService) *AdminHandler {
	return &AdminHandler{serverService: serverService}
}

// @Summary List all servers (admin)
// @Tags admin
// @Security AdminKey
// @Produce json
// @Success 200 {object} map[string][]dto.AdminServerResponse
// @Failure 500 {object} map[string]string
// @Router /admin/servers [get]
func (h *AdminHandler) ListServers(c *gin.Context) {
	servers, err := h.serverService.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list servers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"servers": servers})
}

// @Summary Add VPN server (admin)
// @Tags admin
// @Security AdminKey
// @Accept json
// @Produce json
// @Param request body dto.AddServerRequest true "Server details"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/servers [post]
func (h *AdminHandler) AddServer(c *gin.Context) {
	var req dto.AddServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server, err := h.serverService.AddServer(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": server.ID.String(), "status": "created"})
}

// @Summary Remove VPN server (admin)
// @Tags admin
// @Security AdminKey
// @Produce json
// @Param id path string true "Server UUID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/servers/{id} [delete]
func (h *AdminHandler) RemoveServer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server id"})
		return
	}

	if err := h.serverService.RemoveServer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// @Summary Set server active/inactive (admin)
// @Tags admin
// @Security AdminKey
// @Accept json
// @Produce json
// @Param id path string true "Server UUID"
// @Param request body dto.SetActiveRequest true "Active status"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/servers/{id}/active [put]
func (h *AdminHandler) SetActive(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server id"})
		return
	}

	var req dto.SetActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.serverService.SetActive(c.Request.Context(), id, req.Active); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated", "active": req.Active})
}
