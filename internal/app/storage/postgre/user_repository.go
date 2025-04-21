package sqlstorage

import (
	"database/sql"
	"errors"

	"github.com/andreyxaxa/rest_auth_svc/internal/app/models"
	"github.com/andreyxaxa/rest_auth_svc/internal/app/storage"
	"github.com/google/uuid"
)

type UserRepository struct {
	storage *Storage
}

func (r *UserRepository) Create(u *models.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	u.ID = uuid.New().String()

	_, err := r.storage.db.Exec(
		"INSERT INTO users (id, email, encrypted_password) VALUES ($1, $2, $3)",
		u.ID,
		u.Email,
		u.EncryptedPassword,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	u := &models.User{}
	if err := r.storage.db.QueryRow(
		"SELECT id, email, encrypted_password FROM users WHERE email = $1",
		email,
	).Scan(
		&u.ID,
		&u.Email,
		&u.EncryptedPassword,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	u := &models.User{}
	if err := r.storage.db.QueryRow(
		"SELECT id, email, encrypted_password FROM users WHERE id = $1",
		id,
	).Scan(
		&u.ID,
		&u.Email,
		&u.EncryptedPassword,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}
