package model

type SubscriptionStatusResponse struct {
	Status    string `json:"status"`
	Plan      string `json:"plan"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

