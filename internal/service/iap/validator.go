package iap

import (
	"context"
	"time"
)

type VerifyResult struct {
	Status      string
	ProductID   string
	StoreTxID   string
	Plan        string
	ExpiresAt   time.Time
	Environment *string
	RawReceipt  *string
}

type Validator interface {
	Verify(ctx context.Context, platform string, receipt string, productID string) (*VerifyResult, int, error)
}

