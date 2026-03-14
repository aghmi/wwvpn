package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wwvpn/wwvpn/internal/api/dto"
	"github.com/wwvpn/wwvpn/internal/model"
	mockRepo "github.com/wwvpn/wwvpn/internal/repository/mock"
)

func TestServerService_ListActive(t *testing.T) {
	serverRepo := new(mockRepo.ServerRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := NewServerService(serverRepo, connRepo)
	ctx := context.Background()

	serverID := uuid.New()
	servers := []model.Server{
		{ID: serverID, Country: "NL", City: "Amsterdam", IsActive: true, Capacity: 100},
	}
	serverRepo.On("FindActive", ctx).Return(servers, nil)
	serverRepo.On("CountActiveConnections", ctx, serverID).Return(5, nil)

	resp, err := svc.ListActive(ctx)
	require.NoError(t, err)
	assert.Len(t, resp.Servers, 1)
	assert.Equal(t, "NL", resp.Servers[0].Country)
	assert.Equal(t, 5, resp.Servers[0].Load)
	assert.Equal(t, 100, resp.Servers[0].Capacity)
}

func TestServerService_AddServer(t *testing.T) {
	serverRepo := new(mockRepo.ServerRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := NewServerService(serverRepo, connRepo)
	ctx := context.Background()

	req := &dto.AddServerRequest{
		Country:        "NL",
		City:           "Amsterdam",
		Hostname:       "nl-ams-01",
		IPAddress:      "1.2.3.4",
		GRPCPort:       50051,
		AWGPort:        51820,
		VLESSPort:      443,
		AWGPublicKey:   "key1",
		VLESSPublicKey: "key2",
		VLESSShortID:   "abc123",
	}

	serverRepo.On("Create", ctx, mock.AnythingOfType("*model.Server")).Return(nil)

	server, err := svc.AddServer(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "NL", server.Country)
	assert.Equal(t, "www.microsoft.com", server.CamouflageDomain)
	assert.Equal(t, 100, server.Capacity)
	assert.True(t, server.IsActive)
}

func TestServerService_RemoveServer(t *testing.T) {
	serverRepo := new(mockRepo.ServerRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := NewServerService(serverRepo, connRepo)
	ctx := context.Background()

	id := uuid.New()
	serverRepo.On("Delete", ctx, id).Return(nil)

	err := svc.RemoveServer(ctx, id)
	require.NoError(t, err)
	serverRepo.AssertExpectations(t)
}

func TestServerService_SetActive(t *testing.T) {
	serverRepo := new(mockRepo.ServerRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := NewServerService(serverRepo, connRepo)
	ctx := context.Background()

	id := uuid.New()
	serverRepo.On("SetActive", ctx, id, false).Return(nil)

	err := svc.SetActive(ctx, id, false)
	require.NoError(t, err)
	serverRepo.AssertExpectations(t)
}

func TestServerService_ListAll(t *testing.T) {
	serverRepo := new(mockRepo.ServerRepo)
	connRepo := new(mockRepo.ConnectionRepo)
	svc := NewServerService(serverRepo, connRepo)
	ctx := context.Background()

	id1 := uuid.New()
	id2 := uuid.New()
	servers := []model.Server{
		{ID: id1, Country: "NL", City: "Amsterdam", Hostname: "nl-01", IPAddress: "1.1.1.1", IsActive: true},
		{ID: id2, Country: "DE", City: "Frankfurt", Hostname: "de-01", IPAddress: "2.2.2.2", IsActive: false},
	}
	serverRepo.On("FindAll", ctx).Return(servers, nil)
	serverRepo.On("CountActiveConnections", ctx, id1).Return(10, nil)
	serverRepo.On("CountActiveConnections", ctx, id2).Return(0, nil)

	result, err := svc.ListAll(ctx)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 10, result[0].ActiveConnections)
	assert.Equal(t, 0, result[1].ActiveConnections)
}
