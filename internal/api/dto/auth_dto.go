package dto

type DeviceAuthRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

type AuthResponse struct {
	Token    string `json:"token"`
	ExpireAt int64  `json:"expire_at"`
	DeviceID string `json:"device_id"`
}
