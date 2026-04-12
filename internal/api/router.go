package api

import (
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/wwvpn/wwvpn/internal/api/handler"
	"github.com/wwvpn/wwvpn/internal/api/middleware"
	"github.com/wwvpn/wwvpn/internal/crypto"
)

type Router struct {
	engine         *gin.Engine
	authHandler    *handler.AuthHandler
	serverHandler  *handler.ServerHandler
	connectHandler *handler.ConnectHandler
	subHandler     *handler.SubscriptionHandler
	adminHandler   *handler.AdminHandler
	jwtManager     *crypto.JWTManager
	adminAPIKey    string
}

func NewRouter(
	authHandler *handler.AuthHandler,
	serverHandler *handler.ServerHandler,
	connectHandler *handler.ConnectHandler,
	subHandler *handler.SubscriptionHandler,
	adminHandler *handler.AdminHandler,
	jwtManager *crypto.JWTManager,
	adminAPIKey string,
) *Router {
	return &Router{
		authHandler:    authHandler,
		serverHandler:  serverHandler,
		connectHandler: connectHandler,
		subHandler:     subHandler,
		adminHandler:   adminHandler,
		jwtManager:     jwtManager,
		adminAPIKey:    adminAPIKey,
	}
}

func (r *Router) Setup(mode string) *gin.Engine {
	gin.SetMode(mode)
	r.engine = gin.New()
	r.engine.Use(gin.Recovery())
	r.engine.Use(middleware.RateLimit(100, time.Minute))

	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.engine.Group("/api/v1")

	v1.POST("/auth/device", r.authHandler.DeviceAuth)

	protected := v1.Group("")
	protected.Use(middleware.Auth(r.jwtManager))
	{
		protected.DELETE("/auth/account", r.authHandler.DeleteAccount)
		protected.GET("/servers", r.serverHandler.ListServers)
		protected.GET("/subscription/status", r.subHandler.Status)
		protected.POST("/subscription/verify-receipt", r.subHandler.VerifyReceipt)

		protected.POST("/connect/disconnect", r.connectHandler.Disconnect)
		protected.POST("/connect", r.connectHandler.Connect)
	}

	v1.POST("/webhook/revenuecat", r.subHandler.RevenueCatWebhook)

	admin := v1.Group("/admin")
	admin.Use(middleware.AdminAuth(r.adminAPIKey))
	{
		admin.GET("/servers", r.adminHandler.ListServers)
		admin.POST("/servers", r.adminHandler.AddServer)
		admin.DELETE("/servers/:id", r.adminHandler.RemoveServer)
		admin.PUT("/servers/:id/active", r.adminHandler.SetActive)
	}

	return r.engine
}