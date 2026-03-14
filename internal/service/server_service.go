package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/wwvpn/wwvpn/internal/api/dto"
	"github.com/wwvpn/wwvpn/internal/model"
	"github.com/wwvpn/wwvpn/internal/repository"
)

type ServerService struct {
	serverRepo repository.ServerRepository
	connRepo   repository.ConnectionRepository
}

func NewServerService(serverRepo repository.ServerRepository, connRepo repository.ConnectionRepository) *ServerService {
	return &ServerService{serverRepo: serverRepo, connRepo: connRepo}
}

func (s *ServerService) ListActive(ctx context.Context) (*dto.ServerListResponse, error) {
	servers, err := s.serverRepo.FindActive(ctx)
	if err != nil {
		return nil, err
	}

	resp := &dto.ServerListResponse{
		Servers: make([]dto.ServerResponse, 0, len(servers)),
	}

	for _, srv := range servers {
		load, _ := s.serverRepo.CountActiveConnections(ctx, srv.ID)
		resp.Servers = append(resp.Servers, dto.ServerResponse{
			ID:        srv.ID.String(),
			Country:   srv.Country,
			City:      srv.City,
			IsActive:  srv.IsActive,
			IsPremium: srv.IsPremium,
			Load:      load,
			Capacity:  srv.Capacity,
			Protocols: []string{"amnezia_wg", "vless_reality"},
		})
	}

	return resp, nil
}

func (s *ServerService) AddServer(ctx context.Context, req *dto.AddServerRequest) (*model.Server, error) {
	camoDomain := req.CamouflageDomain
	if camoDomain == "" {
		camoDomain = "www.microsoft.com"
	}
	capacity := req.Capacity
	if capacity == 0 {
		capacity = 100
	}

	server := &model.Server{
		Country:          req.Country,
		City:             req.City,
		Hostname:         req.Hostname,
		IPAddress:        req.IPAddress,
		GRPCPort:         req.GRPCPort,
		AWGPort:          req.AWGPort,
		VLESSPort:        req.VLESSPort,
		AWGPublicKey:     req.AWGPublicKey,
		VLESSPublicKey:   req.VLESSPublicKey,
		VLESSShortID:     req.VLESSShortID,
		CamouflageDomain: camoDomain,
		Capacity:         capacity,
		IsActive:         true,
		IsPremium:        req.IsPremium,
	}

	if err := s.serverRepo.Create(ctx, server); err != nil {
		return nil, err
	}
	return server, nil
}

func (s *ServerService) RemoveServer(ctx context.Context, id uuid.UUID) error {
	return s.serverRepo.Delete(ctx, id)
}

func (s *ServerService) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return s.serverRepo.SetActive(ctx, id, active)
}

func (s *ServerService) ListAll(ctx context.Context) ([]dto.AdminServerResponse, error) {
	servers, err := s.serverRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.AdminServerResponse, 0, len(servers))
	for _, srv := range servers {
		load, _ := s.serverRepo.CountActiveConnections(ctx, srv.ID)
		result = append(result, dto.AdminServerResponse{
			ID:                srv.ID.String(),
			Country:           srv.Country,
			City:              srv.City,
			Hostname:          srv.Hostname,
			IPAddress:         srv.IPAddress,
			IsActive:          srv.IsActive,
			IsPremium:         srv.IsPremium,
			ActiveConnections: load,
		})
	}
	return result, nil
}
