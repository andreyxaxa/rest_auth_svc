package storage

import "github.com/andreyxaxa/rest_auth_svc/internal/app/models"

type UserRepository interface {
	Create(*models.User) error
	FindByEmail(string) (*models.User, error)
	FindByID(string) (*models.User, error)
}
