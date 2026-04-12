package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wwvpn/wwvpn/internal/model"
	"github.com/wwvpn/wwvpn/internal/repository"
	"github.com/wwvpn/wwvpn/internal/service/iap"
)

type SubscriptionService struct {
	subRepo          repository.SubscriptionRepository
	connRepo         repository.ConnectionRepository
	appleValidator   *iap.AppleValidator
	googleValidator  *iap.GoogleValidator
}

func NewSubscriptionService(
	subRepo repository.SubscriptionRepository,
	connRepo repository.ConnectionRepository,
	appleValidator *iap.AppleValidator,
	googleValidator *iap.GoogleValidator,
) *SubscriptionService {
	return &SubscriptionService{
		subRepo:         subRepo,
		connRepo:        connRepo,
		appleValidator:  appleValidator,
		googleValidator: googleValidator,
	}
}

func (s *SubscriptionService) GetActive(ctx context.Context, deviceUUID uuid.UUID) (*model.Subscription, error) {
	return s.subRepo.FindActiveByDeviceID(ctx, deviceUUID)
}

func (s *SubscriptionService) CanConnect(ctx context.Context, deviceUUID uuid.UUID) (bool, error) {
	sub, err := s.subRepo.FindActiveByDeviceID(ctx, deviceUUID)
	if err != nil {
		return false, err
	}
	if !sub.IsActive() {
		return false, nil
	}

	activeCount, err := s.connRepo.CountActiveByDevice(ctx, deviceUUID)
	if err != nil {
		return false, err
	}

	return activeCount < sub.MaxConnections(), nil
}

func (s *SubscriptionService) Upsert(ctx context.Context, sub *model.Subscription) error {
	return s.subRepo.Upsert(ctx, sub)
}

func (s *SubscriptionService) VerifyReceipt(
	ctx context.Context,
	deviceUUID uuid.UUID,
	platform string,
	receipt string,
	productID string,
) (*model.SubscriptionStatusResponse, int, error) {
	var (
		result     *iap.VerifyResult
		statusCode int
		err        error
	)

	switch platform {
	case "ios":
		result, statusCode, err = s.appleValidator.Verify(ctx, platform, receipt, productID)
	case "android":
		result, statusCode, err = s.googleValidator.Verify(ctx, platform, receipt, productID)
	default:
		return nil, 400, fmt.Errorf("invalid platform")
	}

	if err != nil {
		if statusCode == 422 {
			return nil, 422, err
		}
		return nil, statusCode, err
	}

	if result == nil || result.Status == "none" || result.StoreTxID == "" {
		return &model.SubscriptionStatusResponse{Status: "none", Plan: ""}, 200, nil
	}

	sub := &model.Subscription{
		DeviceID:        deviceUUID,
		RevenueCatID:    "",
		StoreTxID:       result.StoreTxID,
		Platform:        platform,
		ProductID:       productID,
		Plan:            mapPlanFromProductID(productID),
		Status:          result.Status,
		ExpiresAt:       result.ExpiresAt,
		Environment:     result.Environment,
		RawReceipt:      &receipt,
		CreatedAt:       time.Time{},
		UpdatedAt:       time.Time{},
	}

	if err := s.subRepo.Upsert(ctx, sub); err != nil {
		return nil, 500, err
	}

	return &model.SubscriptionStatusResponse{
		Status:    sub.Status,
		Plan:      sub.ProductID,
		ExpiresAt: sub.ExpiresAt.Format(time.RFC3339),
	}, 200, nil
}

func mapPlanFromProductID(productID string) string {
	if strings.HasSuffix(productID, ".yearly") {
		return model.PlanAnnual
	}
	if strings.HasSuffix(productID, ".monthly") {
		return model.PlanMonthly
	}
	return model.PlanWeekly
}
