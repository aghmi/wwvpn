package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wwvpn/wwvpn/internal/api/dto"
	"github.com/wwvpn/wwvpn/internal/api/middleware"
	"github.com/wwvpn/wwvpn/internal/service"
)

type ConnectHandler struct {
	vpnService *service.VPNService
}

func NewConnectHandler(vpnService *service.VPNService) *ConnectHandler {
	return &ConnectHandler{vpnService: vpnService}
}

// @Summary Connect to VPN server
// @Tags connect
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ConnectRequest true "Connection params"
// @Success 200 {object} dto.ConnectResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /connect [post]
func (h *ConnectHandler) Connect(c *gin.Context) {
	var req dto.ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceUUIDStr, _ := c.Get(middleware.ContextDeviceUUID)
	deviceUUID, err := uuid.Parse(deviceUUIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	serverID, err := uuid.Parse(req.ServerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server id"})
		return
	}

	result, err := h.vpnService.Connect(c.Request.Context(), deviceUUID, serverID, req.Protocol, req.ClientPublicKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, buildConnectResponse(result))
}

// @Summary Disconnect from VPN server
// @Tags connect
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.DisconnectRequest true "Connection ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /connect/disconnect [post]
func (h *ConnectHandler) Disconnect(c *gin.Context) {
	var req dto.DisconnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceUUIDStr, _ := c.Get(middleware.ContextDeviceUUID)
	deviceUUID, err := uuid.Parse(deviceUUIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	connectionID, err := uuid.Parse(req.ConnectionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection id"})
		return
	}

	if err := h.vpnService.Disconnect(c.Request.Context(), deviceUUID, connectionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "disconnected"})
}

func buildConnectResponse(r *service.ConnectResult) dto.ConnectResponse {
	resp := dto.ConnectResponse{
		Protocol:        r.Protocol,
		Endpoint:        r.Endpoint,
		ServerPublicKey: r.ServerPublicKey,
		AssignedIP:      r.AssignedIP,
		AllowedIPs:      r.AllowedIPs,
		DNS:             r.DNS,
		MTU:             1420,
		ConnectionID:    r.ConnectionID.String(),
	}

	if r.Protocol == "amnezia_wg" {
		resp.Obfuscation = &dto.ObfuscationParams{
			Jc: r.Jc, Jmin: r.Jmin, Jmax: r.Jmax,
			S1: r.S1, S2: r.S2,
			H1: r.H1, H2: r.H2, H3: r.H3, H4: r.H4,
		}
	}

	if r.Protocol == "vless_reality" {
		resp.VLESS = &dto.VLESSParams{
			UUID: r.VlessUUID, PublicKey: r.VlessPublicKey,
			ShortID: r.VlessShortID, SNI: r.VlessSNI, Port: r.VlessPort,
		}
	}

	if r.Fallback != nil {
		fb := &dto.FallbackConfig{
			Protocol:     r.Fallback.Protocol,
			ConnectionID: r.Fallback.ConnectionID.String(),
		}
		if r.Fallback.Protocol == "vless_reality" {
			fb.VLESS = &dto.VLESSParams{
				UUID: r.Fallback.VlessUUID, PublicKey: r.Fallback.VlessPublicKey,
				ShortID: r.Fallback.VlessShortID, SNI: r.Fallback.VlessSNI, Port: r.Fallback.VlessPort,
			}
		}
		if r.Fallback.Protocol == "amnezia_wg" {
			fb.Endpoint = r.Fallback.Endpoint
			fb.ServerPublicKey = r.Fallback.ServerPublicKey
			fb.AssignedIP = r.Fallback.AssignedIP
			fb.Obfuscation = &dto.ObfuscationParams{
				Jc: r.Fallback.Jc, Jmin: r.Fallback.Jmin, Jmax: r.Fallback.Jmax,
				S1: r.Fallback.S1, S2: r.Fallback.S2,
				H1: r.Fallback.H1, H2: r.Fallback.H2, H3: r.Fallback.H3, H4: r.Fallback.H4,
			}
		}
		resp.Fallback = fb
	}

	return resp
}
