package storage

import "github.com/andreyxaxa/rest_auth_svc/internal/app/models"

type TokenRepository interface {
	CreateSession(*models.Session) (*models.Session, error)
	GetSession(string) (*models.Session, error)
	RevokeSession(string) error
	DeleteSession(string) error
}
