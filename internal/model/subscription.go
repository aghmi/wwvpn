package model

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID           uuid.UUID `db:"id" json:"id"`
	DeviceID     uuid.UUID `db:"device_id" json:"device_id"`
	RevenueCatID string    `db:"revenuecat_id" json:"revenuecat_id"`
	Plan         string    `db:"plan" json:"plan"`
	Status       string    `db:"status" json:"status"`
	ExpiresAt    time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

const (
	PlanWeekly  = "weekly"
	PlanMonthly = "monthly"
	PlanAnnual  = "annual"

	StatusActive       = "active"
	StatusExpired      = "expired"
	StatusCancelled    = "cancelled"
	StatusBillingRetry = "billing_retry"
)

func (s *Subscription) IsActive() bool {
	return s.Status == StatusActive && s.ExpiresAt.After(time.Now())
}

func (s *Subscription) MaxConnections() int {
	switch s.Plan {
	case PlanAnnual:
		return 3
	default:
		return 1
	}
}
