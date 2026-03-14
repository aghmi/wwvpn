package main

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wwvpn/wwvpn/internal/api"
	"github.com/wwvpn/wwvpn/internal/api/handler"
	"github.com/wwvpn/wwvpn/internal/config"
	"github.com/wwvpn/wwvpn/internal/crypto"
	"github.com/wwvpn/wwvpn/internal/node"
	"github.com/wwvpn/wwvpn/internal/repository"
	"github.com/wwvpn/wwvpn/internal/service"

	_ "github.com/wwvpn/wwvpn/docs"
)

// @title wwvpn API
// @version 1.0
// @description VPN backend API with AmneziaWG and VLESS+Reality protocols
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey AdminKey
// @in header
// @name X-Admin-Key
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := sqlx.Connect("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	jwtManager := crypto.NewJWTManager(cfg.JWT.Secret, cfg.JWT.TTL)

	deviceRepo := repository.NewDeviceRepo(db)
	serverRepo := repository.NewServerRepo(db)
	subRepo := repository.NewSubscriptionRepo(db)
	connRepo := repository.NewConnectionRepo(db)

	authService := service.NewAuthService(deviceRepo, jwtManager)
	serverService := service.NewServerService(serverRepo, connRepo)
	subService := service.NewSubscriptionService(subRepo, connRepo)
	vpnService := service.NewVPNService(serverRepo, connRepo, node.DefaultClientFactory)

	healthMonitor := service.NewHealthMonitor(serverRepo, node.DefaultClientFactory)
	go healthMonitor.Start(context.Background())

	authHandler := handler.NewAuthHandler(authService)
	serverHandler := handler.NewServerHandler(serverService)
	connectHandler := handler.NewConnectHandler(vpnService)
	subHandler := handler.NewSubscriptionHandler(subService)
	adminHandler := handler.NewAdminHandler(serverService)

	router := api.NewRouter(
		authHandler,
		serverHandler,
		connectHandler,
		subHandler,
		adminHandler,
		jwtManager,
		subService,
		cfg.Admin.APIKey,
	)

	engine := router.Setup(cfg.Server.Mode)

	log.Printf("starting wwvpn API on :%s", cfg.Server.Port)
	if err := engine.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
