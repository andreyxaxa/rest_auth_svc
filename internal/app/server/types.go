package server

import "time"

type UserCreateReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserCreateRes struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type UserLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginRes struct {
	SessionID             string        `json:"session_id"`
	AccessToken           string        `json:"access_token"`
	RefreshToken          string        `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time     `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time     `json:"refresh_token_expires_at"`
	User                  UserCreateRes `json:"user"`
}

type RenewAccessTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}

type RenewAccessTokenRes struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}
