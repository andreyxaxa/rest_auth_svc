package models

import "time"

type Session struct {
	ID               string    `json:"id"`
	UserEmail        string    `json:"user_email"`
	RefreshTokenHash string    `json:"refresh_token_hash"`
	IsRevoked        bool      `json:"is_revoked"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
}
