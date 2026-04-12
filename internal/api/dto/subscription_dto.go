package dto

type VerifyReceiptRequest struct {
	Receipt   string `json:"receipt" binding:"required"`
	Platform  string `json:"platform" binding:"required,oneof=ios android"`
	ProductID string `json:"product_id" binding:"required"`
}

type VerifyReceiptResponse struct {
	Status    string `json:"status"`
	Plan      string `json:"plan"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

