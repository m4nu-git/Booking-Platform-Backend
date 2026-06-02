package db

import (
	"AuthServiceInGo/models"
	"database/sql"
	"time"
)

type PasswordResetTokenRepository interface {
	Create(userId int64, tokenHash string, expiresAt time.Time) error
	FindValidToken(tokenHash string) (*models.PasswordResetToken, error)
	MarkUsed(tokenHash string) error
}

type PasswordResetTokenRepositoryImpl struct {
	db *sql.DB
}

func NewPasswordResetTokenRepository(_db *sql.DB) PasswordResetTokenRepository {
	return &PasswordResetTokenRepositoryImpl{db: _db}
}

func (r *PasswordResetTokenRepositoryImpl) Create(userId int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		"INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES (?, ?, ?)",
		userId, tokenHash, expiresAt,
	)
	return err
}

func (r *PasswordResetTokenRepositoryImpl) FindValidToken(tokenHash string) (*models.PasswordResetToken, error) {
	t := &models.PasswordResetToken{}
	err := r.db.QueryRow(
		"SELECT id, user_id, token_hash, expires_at, used FROM password_reset_tokens WHERE token_hash = ? AND used = FALSE AND expires_at > NOW()",
		tokenHash,
	).Scan(&t.Id, &t.UserId, &t.TokenHash, &t.ExpiresAt, &t.Used)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (r *PasswordResetTokenRepositoryImpl) MarkUsed(tokenHash string) error {
	_, err := r.db.Exec("UPDATE password_reset_tokens SET used = TRUE WHERE token_hash = ?", tokenHash)
	return err
}
