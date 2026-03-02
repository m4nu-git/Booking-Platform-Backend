package db

import (
	"AuthServiceInGo/models"
	"database/sql"
	"time"
)

type RefreshTokenRepository interface {
	Create(userId int64, token string, expiresAt time.Time) error
	GetByToken(token string) (*models.RefreshToken, error)
	DeleteByToken(token string) error
}

type RefreshTokenRepositoryImpl struct {
	db *sql.DB
}

func NewRefreshTokenRepository(_db *sql.DB) RefreshTokenRepository {
	return &RefreshTokenRepositoryImpl{db: _db}
}

func (r *RefreshTokenRepositoryImpl) Create(userId int64, token string, expiresAt time.Time) error {
	query := "INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES (?, ?, ?)"
	_, err := r.db.Exec(query, userId, token, expiresAt.Format("2006-01-02 15:04:05"))
	return err
}

func (r *RefreshTokenRepositoryImpl) GetByToken(token string) (*models.RefreshToken, error) {
	query := "SELECT id, user_id, token, expires_at, created_at FROM refresh_tokens WHERE token = ?"
	row := r.db.QueryRow(query, token)

	rt := &models.RefreshToken{}
	if err := row.Scan(&rt.Id, &rt.UserId, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt); err != nil {
		return nil, err
	}
	return rt, nil
}

func (r *RefreshTokenRepositoryImpl) DeleteByToken(token string) error {
	_, err := r.db.Exec("DELETE FROM refresh_tokens WHERE token = ?", token)
	return err
}
