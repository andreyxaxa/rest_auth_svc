package sqlstorage

import (
	"database/sql"

	"github.com/andreyxaxa/rest_auth_svc/internal/app/storage"
	_ "github.com/lib/pq"
)

type Storage struct {
	db              *sql.DB
	userRepository  storage.UserRepository
	tokenRepository storage.TokenRepository
}

func New(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) User() storage.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		storage: s,
	}

	return s.userRepository
}

func (s *Storage) Token() storage.TokenRepository {
	if s.tokenRepository != nil {
		return s.tokenRepository
	}

	s.tokenRepository = &TokenRepository{
		storage: s,
	}

	return s.tokenRepository
}
