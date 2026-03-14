package service

import (
	"context"

	"github.com/wwvpn/wwvpn/internal/crypto"
	"github.com/wwvpn/wwvpn/internal/model"
	"github.com/wwvpn/wwvpn/internal/repository"
)

type AuthService struct {
	deviceRepo repository.DeviceRepository
	jwtManager *crypto.JWTManager
}

func NewAuthService(deviceRepo repository.DeviceRepository, jwtManager *crypto.JWTManager) *AuthService {
	return &AuthService{
		deviceRepo: deviceRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Authenticate(ctx context.Context, deviceID string) (string, int64, *model.Device, error) {
	device, err := s.deviceRepo.FindOrCreate(ctx, deviceID)
	if err != nil {
		return "", 0, nil, err
	}

	token, expiresAt, err := s.jwtManager.Generate(device.ID)
	if err != nil {
		return "", 0, nil, err
	}

	return token, expiresAt, device, nil
}

func (s *AuthService) DeleteDevice(ctx context.Context, deviceUUID string) error {
	uid, err := parseUUID(deviceUUID)
	if err != nil {
		return err
	}
	return s.deviceRepo.SoftDelete(ctx, uid)
}
