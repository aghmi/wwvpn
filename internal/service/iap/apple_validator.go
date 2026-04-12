package iap

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/devsisters/go-applereceipt"
	"github.com/devsisters/go-applereceipt/applepki"
	appstoreapi "github.com/awa/go-iap/appstore/api"
)

type AppleValidator struct {
	keyContent []byte
	keyID      string
	issuer     string
	bundleID   string
}

func NewAppleValidatorFromEnv() *AppleValidator {
	privateKeyPath := os.Getenv("APPLE_PRIVATE_KEY_PATH")
	keyContent, _ := os.ReadFile(privateKeyPath)

	return &AppleValidator{
		keyContent: keyContent,
		keyID:      os.Getenv("APPLE_KEY_ID"),
		issuer:     os.Getenv("APPLE_ISSUER_ID"),
		bundleID:   os.Getenv("APPLE_BUNDLE_ID"),
	}
}

func (v *AppleValidator) Verify(ctx context.Context, platform string, receipt string, productID string) (*VerifyResult, int, error) {
	if platform != "ios" {
		return &VerifyResult{Status: "none"}, 200, nil
	}

	if len(v.keyContent) == 0 || v.keyID == "" || v.issuer == "" || v.bundleID == "" {
		return nil, 422, fmt.Errorf("apple iap config is not set")
	}

	decoded, err := applereceipt.DecodeBase64(receipt, applepki.CertPool())
	if err != nil {
		return nil, 422, fmt.Errorf("receipt decode failed: %w", err)
	}

	var originalTxID string
	foundProduct := false
	for _, inapp := range decoded.InAppPurchaseReceipts {
		if inapp.ProductIdentifier == productID {
			originalTxID = inapp.OriginalTransactionIdentifier
			foundProduct = true
			break
		}
	}

	if !foundProduct || originalTxID == "" {
		return &VerifyResult{Status: "none"}, 200, nil
	}

	query := &url.Values{}
	query.Set("productType", "AUTO_RENEWABLE")

	result, err := v.verifyWithSandboxFallback(ctx, originalTxID, query, productID)
	if err != nil {
		return nil, 422, err
	}

	if result == nil || result.StoreTxID == "" {
		return &VerifyResult{Status: "none"}, 200, nil
	}

	return result, 200, nil
}

func (v *AppleValidator) verifyWithSandboxFallback(
	ctx context.Context,
	originalTxID string,
	query *url.Values,
	productID string,
) (*VerifyResult, error) {
	prodClient := apiClient(v, false)
	res, err := prodClient.GetALLSubscriptionStatuses(ctx, originalTxID, query)
	if err != nil {
		sandboxClient := apiClient(v, true)
		res2, err2 := sandboxClient.GetALLSubscriptionStatuses(ctx, originalTxID, query)
		if err2 != nil {
			return nil, fmt.Errorf("apple server api error: %w", err2)
		}
		return v.buildResultFromStatus(sandboxClient, res2, productID, originalTxID)
	}

	return v.buildResultFromStatus(prodClient, res, productID, originalTxID)
}

func apiClient(v *AppleValidator, sandbox bool) *appstoreapi.StoreClient {
	cfg := &appstoreapi.StoreConfig{
		KeyContent: v.keyContent,
		KeyID:      v.keyID,
		BundleID:   v.bundleID,
		Issuer:     v.issuer,
		Sandbox:    sandbox,
	}
	return appstoreapi.NewStoreClient(cfg)
}

func (v *AppleValidator) buildResultFromStatus(client *appstoreapi.StoreClient, res *appstoreapi.StatusResponse, productID string, originalTxID string) (*VerifyResult, error) {
	if res == nil {
		return nil, errors.New("nil apple status response")
	}

	for _, group := range res.Data {
		for _, last := range group.LastTransactions {
			renewalPayloadRaw, err := client.ParseJWSEncodeString(last.SignedRenewalInfo)
			if err != nil {
				continue
			}
			renewalPayload, ok := renewalPayloadRaw.(*appstoreapi.JWSRenewalInfoDecodedPayload)
			if !ok || renewalPayload == nil {
				continue
			}

			if renewalPayload.ProductId != productID {
				continue
			}

			status := "none"
			switch last.Status {
			case appstoreapi.SubscriptionActive:
				status = "active"
			case appstoreapi.SubscriptionExpired:
				status = "expired"
			case appstoreapi.SubscriptionRetryPeriod:
				status = "billing_retry"
			case appstoreapi.SubscriptionGracePeriod:
				status = "grace_period"
			case appstoreapi.SubscriptionRevoked:
				status = "cancelled"
			default:
				status = "none"
			}

			expiresAt := time.UnixMilli(renewalPayload.RenewalDate)
			env := strings.ToLower(string(renewalPayload.Environment))
			envPtr := &env

			return &VerifyResult{
				Status:      status,
				ProductID:   productID,
				StoreTxID:   originalTxID,
				Plan:        "",
				ExpiresAt:   expiresAt,
				Environment: envPtr,
				RawReceipt:  nil,
			}, nil
		}
	}

	return &VerifyResult{Status: "none"}, nil
}

