package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wwvpn/wwvpn/internal/crypto"
	"github.com/wwvpn/wwvpn/internal/model"
	mockRepo "github.com/wwvpn/wwvpn/internal/repository/mock"
)

func TestAuthService_Authenticate_NewDevice(t *testing.T) {
	repo := new(mockRepo.DeviceRepo)
	mgr := crypto.NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)
	svc := NewAuthService(repo, mgr)

	device := &model.Device{ID: uuid.New(), DeviceID: "test-device-123"}
	repo.On("FindOrCreate", context.Background(), "test-device-123").Return(device, nil)

	token, exp, d, err := svc.Authenticate(context.Background(), "test-device-123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, exp > 0)
	assert.Equal(t, "test-device-123", d.DeviceID)
	repo.AssertExpectations(t)
}

func TestAuthService_Authenticate_RepoError(t *testing.T) {
	repo := new(mockRepo.DeviceRepo)
	mgr := crypto.NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)
	svc := NewAuthService(repo, mgr)

	repo.On("FindOrCreate", context.Background(), "bad-device").Return(nil, errors.New("db error"))

	_, _, _, err := svc.Authenticate(context.Background(), "bad-device")
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestAuthService_DeleteDevice(t *testing.T) {
	repo := new(mockRepo.DeviceRepo)
	mgr := crypto.NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)
	svc := NewAuthService(repo, mgr)

	id := uuid.New()
	repo.On("SoftDelete", context.Background(), id).Return(nil)

	err := svc.DeleteDevice(context.Background(), id.String())
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAuthService_DeleteDevice_InvalidUUID(t *testing.T) {
	repo := new(mockRepo.DeviceRepo)
	mgr := crypto.NewJWTManager("test-secret-32-chars-long!!!!!!!!", 60)
	svc := NewAuthService(repo, mgr)

	err := svc.DeleteDevice(context.Background(), "not-a-uuid")
	assert.Error(t, err)
}
