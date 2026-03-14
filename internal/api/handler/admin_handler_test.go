package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wwvpn/wwvpn/internal/model"
	mockRepo "github.com/wwvpn/wwvpn/internal/repository/mock"
	"github.com/wwvpn/wwvpn/internal/service"
)

func setupAdminRouter(h *AdminHandler) *gin.Engine {
	r := gin.New()
	r.GET("/admin/servers", h.ListServers)
	r.POST("/admin/servers", h.AddServer)
	r.DELETE("/admin/servers/:id", h.RemoveServer)
	r.PUT("/admin/servers/:id/active", h.SetActive)
	return r
}

func TestAdminHandler_ListServers(t *testing.T) {
	serverRepo := new(mockRepo.ServerRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := service.NewServerService(serverRepo, connRepo)
	h := NewAdminHandler(svc)

	id := uuid.New()
	serverRepo.On("FindAll", mock.Anything).Return([]model.Server{
		{ID: id, Country: "NL", City: "Amsterdam", Hostname: "nl-01", IPAddress: "1.1.1.1", IsActive: true},
	}, nil)
	serverRepo.On("CountActiveConnections", mock.Anything, id).Return(3, nil)

	r := setupAdminRouter(h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/servers", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	servers := resp["servers"].([]interface{})
	assert.Len(t, servers, 1)
}

func TestAdminHandler_AddServer(t *testing.T) {
	serverRepo := new(mockRepo.ServerRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := service.NewServerService(serverRepo, connRepo)
	h := NewAdminHandler(svc)

	serverRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Server")).Return(nil)

	body, _ := json.Marshal(map[string]interface{}{
		"country":          "NL",
		"city":             "Amsterdam",
		"hostname":         "nl-01",
		"ip_address":       "1.1.1.1",
		"grpc_port":        50051,
		"awg_port":         51820,
		"vless_port":       443,
		"awg_public_key":   "key1",
		"vless_public_key": "key2",
		"vless_short_id":   "abc",
	})

	r := setupAdminRouter(h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/servers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAdminHandler_RemoveServer(t *testing.T) {
	serverRepo := new(mockRepo.ServerRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := service.NewServerService(serverRepo, connRepo)
	h := NewAdminHandler(svc)

	id := uuid.New()
	serverRepo.On("Delete", mock.Anything, id).Return(nil)

	r := setupAdminRouter(h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/servers/"+id.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_SetActive(t *testing.T) {
	serverRepo := new(mockRepo.ServerRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := service.NewServerService(serverRepo, connRepo)
	h := NewAdminHandler(svc)

	id := uuid.New()
	serverRepo.On("SetActive", mock.Anything, id, false).Return(nil)

	body, _ := json.Marshal(map[string]bool{"active": false})

	r := setupAdminRouter(h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/servers/"+id.String()+"/active", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_RemoveServer_InvalidID(t *testing.T) {
	serverRepo := new(mockRepo.ServerRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := service.NewServerService(serverRepo, connRepo)
	h := NewAdminHandler(svc)

	r := setupAdminRouter(h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/servers/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
