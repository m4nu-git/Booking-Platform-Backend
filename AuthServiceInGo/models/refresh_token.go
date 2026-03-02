package models

type RefreshToken struct {
	Id        int64  `json:"id"`
	UserId    int64  `json:"user_id"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}
