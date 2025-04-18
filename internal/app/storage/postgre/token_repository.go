package sqlstorage

import "github.com/andreyxaxa/rest_auth_svc/internal/app/models"

type TokenRepository struct {
	storage *Storage
}

func (t *TokenRepository) CreateSession(session *models.Session) (*models.Session, error) {
	_, err := t.storage.db.Exec(
		"INSERT INTO sessions (id, user_email, refresh_token_hash, is_revoked, expires_at) VALUES ($1, $2, $3, $4, $5)",
		session.ID,
		session.UserEmail,
		session.RefreshTokenHash,
		session.IsRevoked,
		session.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (t *TokenRepository) GetSession(id string) (*models.Session, error) {
	session := &models.Session{}

	err := t.storage.db.QueryRow(
		"SELECT id, user_email, refresh_token_hash, is_revoked, created_at, expires_at FROM sessions WHERE id = $1",
		id,
	).Scan(
		&session.ID,
		&session.UserEmail,
		&session.RefreshTokenHash,
		&session.IsRevoked,
		&session.CreatedAt,
		&session.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (t *TokenRepository) RevokeSession(id string) error {
	_, err := t.storage.db.Exec(
		"UPDATE sessions SET is_revoked = 1 WHERE id = $1",
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (t *TokenRepository) DeleteSession(id string) error {
	_, err := t.storage.db.Exec(
		"DELETE FROM sessions WHERE id = $1",
		id,
	)
	if err != nil {
		return err
	}

	return nil
}
