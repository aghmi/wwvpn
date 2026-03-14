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
	"github.com/wwvpn/wwvpn/internal/api/middleware"
	"github.com/wwvpn/wwvpn/internal/crypto"
	"github.com/wwvpn/wwvpn/internal/model"
	mockRepo "github.com/wwvpn/wwvpn/internal/repository/mock"
	"github.com/wwvpn/wwvpn/internal/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthHandler_DeviceAuth_Success(t *testing.T) {
	repo := new(mockRepo.DeviceRepo)
	mgr := crypto.NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)
	svc := service.NewAuthService(repo, mgr)
	h := NewAuthHandler(svc)

	device := &model.Device{ID: uuid.New(), DeviceID: "ios-vendor-id-123"}
	repo.On("FindOrCreate", mock.Anything, "ios-vendor-id-123").Return(device, nil)

	r := gin.New()
	r.POST("/auth/device", h.DeviceAuth)

	body, _ := json.Marshal(map[string]string{"device_id": "ios-vendor-id-123"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/device", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["token"])
	assert.Equal(t, "ios-vendor-id-123", resp["device_id"])
}

func TestAuthHandler_DeviceAuth_MissingBody(t *testing.T) {
	repo := new(mockRepo.DeviceRepo)
	mgr := crypto.NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)
	svc := service.NewAuthService(repo, mgr)
	h := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/auth/device", h.DeviceAuth)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/device", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_DeleteAccount_Success(t *testing.T) {
	repo := new(mockRepo.DeviceRepo)
	mgr := crypto.NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)
	svc := service.NewAuthService(repo, mgr)
	h := NewAuthHandler(svc)

	deviceUUID := uuid.New()
	repo.On("SoftDelete", mock.Anything, deviceUUID).Return(nil)

	r := gin.New()
	r.DELETE("/auth/account", func(c *gin.Context) {
		c.Set(middleware.ContextDeviceUUID, deviceUUID.String())
		h.DeleteAccount(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/auth/account", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}
