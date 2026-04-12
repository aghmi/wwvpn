package iap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

type GoogleValidator struct {
	serviceAccountPath string
	packageName        string
}

func NewGoogleValidatorFromEnv() *GoogleValidator {
	return &GoogleValidator{
		serviceAccountPath: os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON"),
		packageName:        os.Getenv("GOOGLE_PACKAGE_NAME"),
	}
}

func (v *GoogleValidator) Verify(ctx context.Context, platform string, receipt string, productID string) (*VerifyResult, int, error) {
	if platform != "android" {
		return &VerifyResult{Status: "none"}, 200, nil
	}

	if v.serviceAccountPath == "" || v.packageName == "" {
		return nil, 422, fmt.Errorf("google iap config is not set")
	}

	srv, err := androidpublisher.NewService(ctx, option.WithCredentialsFile(v.serviceAccountPath))
	if err != nil {
		return nil, 422, fmt.Errorf("google api init failed: %w", err)
	}

	resp, err := srv.Purchases.Subscriptionsv2.Get(v.packageName, receipt).Do()
	if err != nil {
		return nil, 422, fmt.Errorf("google api error: %w", err)
	}

	lineItem := pickGoogleLineItem(resp, productID)
	if lineItem == nil {
		return &VerifyResult{Status: "none"}, 200, nil
	}

	expiresAt, err := parseGoogleTime(lineItem.ExpiryTime)
	if err != nil {
		return nil, 422, fmt.Errorf("google expiry time parse failed: %w", err)
	}

	status := mapGoogleSubscriptionState(resp.SubscriptionState)

	env := "production"
	if resp.TestPurchase != nil {
		env = "sandbox"
	}

	return &VerifyResult{
		Status:      status,
		ProductID:   productID,
		StoreTxID:   pickGoogleStoreTxID(lineItem),
		Plan:        "",
		ExpiresAt:   expiresAt,
		Environment: &env,
		RawReceipt:  nil,
	}, 200, nil
}

func pickGoogleLineItem(resp *androidpublisher.SubscriptionPurchaseV2, productID string) *androidpublisher.SubscriptionPurchaseLineItem {
	if resp == nil {
		return nil
	}
	for _, li := range resp.LineItems {
		if li == nil {
			continue
		}
		if strings.EqualFold(li.ProductId, productID) {
			return li
		}
	}
	if len(resp.LineItems) == 0 {
		return nil
	}
	return resp.LineItems[0]
}

func pickGoogleStoreTxID(li *androidpublisher.SubscriptionPurchaseLineItem) string {
	if li == nil {
		return ""
	}
	if li.LatestSuccessfulOrderId != "" {
		return li.LatestSuccessfulOrderId
	}
	return ""
}

func parseGoogleTime(v string) (time.Time, error) {
	if v == "" {
		return time.Time{}, errors.New("empty expiryTime")
	}
	return time.Parse(time.RFC3339, v)
}

func mapGoogleSubscriptionState(state string) string {
	switch state {
	case "SUBSCRIPTION_STATE_ACTIVE":
		return "active"
	case "SUBSCRIPTION_STATE_EXPIRED":
		return "expired"
	case "SUBSCRIPTION_STATE_CANCELED":
		return "cancelled"
	case "SUBSCRIPTION_STATE_IN_GRACE_PERIOD":
		return "grace_period"
	case "SUBSCRIPTION_STATE_ON_HOLD":
		return "billing_retry"
	case "SUBSCRIPTION_STATE_PAUSED":
		return "expired"
	default:
		return "none"
	}
}

