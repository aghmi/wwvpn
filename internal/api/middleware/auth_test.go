package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wwvpn/wwvpn/internal/crypto"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupAuthRouter(jwtManager *crypto.JWTManager) *gin.Engine {
	r := gin.New()
	r.Use(Auth(jwtManager))
	r.GET("/test", func(c *gin.Context) {
		deviceUUID, _ := c.Get(ContextDeviceUUID)
		c.JSON(200, gin.H{"device_uuid": deviceUUID})
	})
	return r
}

func TestAuth_ValidToken(t *testing.T) {
	mgr := crypto.NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)
	deviceID := uuid.New()
	token, _, err := mgr.Generate(deviceID)
	require.NoError(t, err)

	r := setupAuthRouter(mgr)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_MissingHeader(t *testing.T) {
	mgr := crypto.NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)
	r := setupAuthRouter(mgr)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidFormat(t *testing.T) {
	mgr := crypto.NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)
	r := setupAuthRouter(mgr)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	mgr := crypto.NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)
	r := setupAuthRouter(mgr)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.value")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminAuth_Valid(t *testing.T) {
	r := gin.New()
	r.Use(AdminAuth("my-secret-key"))
	r.GET("/admin", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("X-Admin-Key", "my-secret-key")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminAuth_Missing(t *testing.T) {
	r := gin.New()
	r.Use(AdminAuth("my-secret-key"))
	r.GET("/admin", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminAuth_Wrong(t *testing.T) {
	r := gin.New()
	r.Use(AdminAuth("my-secret-key"))
	r.GET("/admin", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("X-Admin-Key", "wrong-key")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
