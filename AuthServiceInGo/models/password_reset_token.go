package models

import "time"

type PasswordResetToken struct {
	Id        int64
	UserId    int64
	TokenHash string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}
