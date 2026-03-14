package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/wwvpn/wwvpn/internal/model"
	"github.com/wwvpn/wwvpn/internal/node"
	"github.com/wwvpn/wwvpn/internal/repository"
)

type VPNService struct {
	serverRepo repository.ServerRepository
	connRepo   repository.ConnectionRepository
	mu         sync.Mutex
	clients    map[uuid.UUID]node.NodeClient
	newClient  node.NodeClientFactory
}

func NewVPNService(serverRepo repository.ServerRepository, connRepo repository.ConnectionRepository, factory node.NodeClientFactory) *VPNService {
	return &VPNService{
		serverRepo: serverRepo,
		connRepo:   connRepo,
		clients:    make(map[uuid.UUID]node.NodeClient),
		newClient:  factory,
	}
}

func (s *VPNService) getClient(server *model.Server) (node.NodeClient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.clients[server.ID]; ok {
		return c, nil
	}

	addr := fmt.Sprintf("%s:%d", server.IPAddress, server.GRPCPort)
	c, err := s.newClient(addr)
	if err != nil {
		return nil, fmt.Errorf("node %s unreachable: %w", server.IPAddress, err)
	}

	s.clients[server.ID] = c
	return c, nil
}

type ConnectResult struct {
	ConnectionID    uuid.UUID
	Protocol        string
	Endpoint        string
	ServerPublicKey string
	AssignedIP      string
	AllowedIPs      string
	DNS             string
	Jc, Jmin, Jmax  int
	S1, S2          int
	H1, H2, H3, H4 int
	VlessUUID       string
	VlessPublicKey  string
	VlessShortID    string
	VlessSNI        string
	VlessPort       int
	Fallback        *ConnectResult
}

func (s *VPNService) Connect(ctx context.Context, deviceUUID uuid.UUID, serverID uuid.UUID, protocol string, clientPubKey string) (*ConnectResult, error) {
	if protocol == "auto" {
		return s.connectAuto(ctx, deviceUUID, serverID, clientPubKey)
	}
	return s.connectSingle(ctx, deviceUUID, serverID, protocol, clientPubKey)
}

func (s *VPNService) connectAuto(ctx context.Context, deviceUUID uuid.UUID, serverID uuid.UUID, clientPubKey string) (*ConnectResult, error) {
	primary, err := s.connectSingle(ctx, deviceUUID, serverID, model.ProtocolAmneziaWG, clientPubKey)
	if err != nil {
		return nil, err
	}

	fallback, err := s.connectSingle(ctx, deviceUUID, serverID, model.ProtocolVLESSReality, clientPubKey)
	if err != nil {
		log.Printf("[vpn] fallback VLESS connection failed for device %s: %v", deviceUUID, err)
		return primary, nil
	}

	primary.Fallback = fallback
	return primary, nil
}

func (s *VPNService) connectSingle(ctx context.Context, deviceUUID uuid.UUID, serverID uuid.UUID, protocol string, clientPubKey string) (*ConnectResult, error) {
	server, err := s.serverRepo.FindByID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	if !server.IsActive {
		return nil, fmt.Errorf("server is not active")
	}

	client, err := s.getClient(server)
	if err != nil {
		return nil, err
	}

	resp, err := client.AddPeer(ctx, protocol, clientPubKey, deviceUUID.String())
	if err != nil {
		return nil, fmt.Errorf("node agent error: %w", err)
	}

	conn := &model.Connection{
		DeviceID: deviceUUID,
		ServerID: serverID,
		Protocol: protocol,
	}

	if protocol == model.ProtocolAmneziaWG {
		conn.ClientPublicKey = &clientPubKey
	}
	if protocol == model.ProtocolVLESSReality {
		conn.VlessUUID = &resp.VlessUuid
	}

	if err := s.connRepo.Create(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to save connection: %w", err)
	}

	return &ConnectResult{
		ConnectionID:    conn.ID,
		Protocol:        protocol,
		Endpoint:        resp.Endpoint,
		ServerPublicKey: resp.ServerPublicKey,
		AssignedIP:      resp.AssignedIp,
		AllowedIPs:      resp.AllowedIps,
		DNS:             resp.Dns,
		Jc:              int(resp.Jc),
		Jmin:            int(resp.Jmin),
		Jmax:            int(resp.Jmax),
		S1:              int(resp.S1),
		S2:              int(resp.S2),
		H1:              int(resp.H1),
		H2:              int(resp.H2),
		H3:              int(resp.H3),
		H4:              int(resp.H4),
		VlessUUID:       resp.VlessUuid,
		VlessPublicKey:  resp.VlessPublicKey,
		VlessShortID:    resp.VlessShortId,
		VlessSNI:        resp.VlessSni,
		VlessPort:       int(resp.VlessPort),
	}, nil
}

func (s *VPNService) Disconnect(ctx context.Context, deviceUUID uuid.UUID, connectionID uuid.UUID) error {
	conn, err := s.connRepo.FindByID(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("connection not found: %w", err)
	}

	if conn.DeviceID != deviceUUID {
		return fmt.Errorf("connection does not belong to device")
	}

	if conn.DisconnectedAt != nil {
		return nil
	}

	server, err := s.serverRepo.FindByID(ctx, conn.ServerID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	client, err := s.getClient(server)
	if err != nil {
		return err
	}

	pubKey := ""
	if conn.ClientPublicKey != nil {
		pubKey = *conn.ClientPublicKey
	}
	vlessUUID := ""
	if conn.VlessUUID != nil {
		vlessUUID = *conn.VlessUUID
	}

	if err := client.RemovePeer(ctx, conn.Protocol, pubKey, vlessUUID); err != nil {
		return fmt.Errorf("node agent error: %w", err)
	}

	return s.connRepo.Disconnect(ctx, connectionID)
}

func (s *VPNService) DisconnectAll(ctx context.Context, deviceUUID uuid.UUID) error {
	conns, err := s.connRepo.FindActiveByDevice(ctx, deviceUUID)
	if err != nil {
		return err
	}

	for _, conn := range conns {
		_ = s.Disconnect(ctx, deviceUUID, conn.ID)
	}

	return nil
}
