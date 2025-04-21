package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/smtp"
	"time"

	"github.com/andreyxaxa/rest_auth_svc/internal/app/models"
	"github.com/andreyxaxa/rest_auth_svc/internal/app/storage"
	"github.com/andreyxaxa/rest_auth_svc/internal/app/token"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type server struct {
	logger     *logrus.Logger
	router     *mux.Router
	storage    storage.Storage
	tokenMaker *token.JWTMaker
}

func newServer(storage storage.Storage, secretKey string) *server {
	s := &server{
		logger:     logrus.New(),
		router:     mux.NewRouter(),
		storage:    storage,
		tokenMaker: token.NewJWTMaker(secretKey),
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.HandleFunc("/users", s.handleUsersCreate()).Methods("POST")
	s.router.HandleFunc("/login", s.handleUsersLogin()).Methods("POST")

	s.router.HandleFunc("/tokens/{id}", s.handleUsersTokens()).Methods("GET")             // 1 - выдача токенов по GUID
	s.router.HandleFunc("/tokens/refresh", s.handleUsersTokensRefresh()).Methods("POST")  // 2 - рефреш пары токенов
	s.router.HandleFunc("/tokens/renew", s.handleUsersRenewAccessToken()).Methods("POST") // 3 - обновление Access токена
}

// ----- handlers

func (s *server) handleUsersCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &UserCreateReq{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &models.User{
			Email:    req.Email,
			Password: req.Password,
		}

		if err := s.storage.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		res := &UserCreateRes{
			ID:    u.ID,
			Email: u.Email,
		}

		s.respond(w, r, http.StatusCreated, res)
	}
}

func (s *server) handleUsersLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &UserLoginReq{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.storage.User().FindByEmail(req.Email)
		if err != nil || !u.ComparePassword(req.Password) {
			s.error(w, r, http.StatusUnauthorized, errIncorrectEmailOrPassword)
			return
		}

		// creating tokens
		accessToken, accessClaims, err := s.tokenMaker.CreateToken(u.ID, u.Email, r.RemoteAddr, 15*time.Minute)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		refreshToken, refreshClaims, err := s.tokenMaker.CreateToken(u.ID, u.Email, r.RemoteAddr, 24*time.Hour)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		refreshTokenHash, err := hashRefreshToken(refreshToken)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		session, err := s.storage.Token().CreateSession(&models.Session{
			ID:               refreshClaims.RegisteredClaims.ID,
			UserEmail:        u.Email,
			RefreshTokenHash: string(refreshTokenHash),
			IsRevoked:        false,
			ExpiresAt:        refreshClaims.RegisteredClaims.ExpiresAt.Time,
		})
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		res := &UserLoginRes{
			SessionID:             session.ID,
			AccessToken:           accessToken,
			RefreshToken:          refreshToken,
			AccessTokenExpiresAt:  accessClaims.RegisteredClaims.ExpiresAt.Time,
			RefreshTokenExpiresAt: refreshClaims.RegisteredClaims.ExpiresAt.Time,
			User: UserCreateRes{
				ID:    u.ID,
				Email: u.Email,
			},
		}

		s.respond(w, r, http.StatusOK, res)
	}
}

func (s *server) handleUsersTokens() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			s.error(w, r, http.StatusInternalServerError, errors.New("user with this id not found"))
		}

		u, err := s.storage.User().FindByID(id)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errIncorrectEmailOrPassword)
			return
		}

		// creating tokens
		accessToken, accessClaims, err := s.tokenMaker.CreateToken(u.ID, u.Email, r.RemoteAddr, 15*time.Minute)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		refreshToken, refreshClaims, err := s.tokenMaker.CreateToken(u.ID, u.Email, r.RemoteAddr, 24*time.Hour)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		session, err := s.storage.Token().CreateSession(&models.Session{
			ID:               refreshClaims.RegisteredClaims.ID,
			UserEmail:        u.Email,
			RefreshTokenHash: refreshToken,
			IsRevoked:        false,
			ExpiresAt:        refreshClaims.RegisteredClaims.ExpiresAt.Time,
		})
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		res := &UserLoginRes{
			SessionID:             session.ID,
			AccessToken:           accessToken,
			RefreshToken:          refreshToken,
			AccessTokenExpiresAt:  accessClaims.RegisteredClaims.ExpiresAt.Time,
			RefreshTokenExpiresAt: refreshClaims.RegisteredClaims.ExpiresAt.Time,
			User: UserCreateRes{
				ID:    u.ID,
				Email: u.Email,
			},
		}

		s.respond(w, r, http.StatusOK, res)
	}
}

func (s *server) handleUsersTokensRefresh() http.HandlerFunc {
	type request struct {
		ID      string `json:"id"`
		Sess_id string `json:"session_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			s.respond(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.storage.User().FindByID(req.ID)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errIncorrectEmailOrPassword)
			return
		}

		session, err := s.storage.Token().GetSession(req.Sess_id)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		err = s.storage.Token().DeleteSession(session.ID)
		if err != nil {
			s.respond(w, r, http.StatusInternalServerError, err)
			return
		}

		// creating tokens
		accessToken, accessClaims, err := s.tokenMaker.CreateToken(u.ID, u.Email, r.RemoteAddr, 15*time.Minute)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		refreshToken, refreshClaims, err := s.tokenMaker.CreateToken(u.ID, u.Email, r.RemoteAddr, 24*time.Hour)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		refreshTokenHash, err := hashRefreshToken(refreshToken)
		if err != nil {
			s.respond(w, r, http.StatusInternalServerError, err)
			return
		}

		session, err = s.storage.Token().CreateSession(&models.Session{
			ID:               refreshClaims.RegisteredClaims.ID,
			UserEmail:        u.Email,
			RefreshTokenHash: refreshTokenHash,
			IsRevoked:        false,
			ExpiresAt:        refreshClaims.RegisteredClaims.ExpiresAt.Time,
		})
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		res := &UserLoginRes{
			SessionID:             session.ID,
			AccessToken:           accessToken,
			RefreshToken:          refreshToken,
			AccessTokenExpiresAt:  accessClaims.RegisteredClaims.ExpiresAt.Time,
			RefreshTokenExpiresAt: refreshClaims.RegisteredClaims.ExpiresAt.Time,
			User: UserCreateRes{
				ID:    u.ID,
				Email: u.Email,
			},
		}

		s.respond(w, r, http.StatusOK, res)
	}
}

func (s *server) handleUsersRenewAccessToken() http.HandlerFunc {
	type request struct {
		RefreshToken string `json:"refresh_token"`
	}

	type response struct {
		AccessToken          string    `json:"access_token"`
		AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		refreshClaims, err := s.tokenMaker.VerifyToken(req.RefreshToken)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}

		session, err := s.storage.Token().GetSession(refreshClaims.RegisteredClaims.ID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if refreshClaims.IP != r.RemoteAddr {
			err = s.sendEmailWarning(session.UserEmail, r.RemoteAddr) // email-warning
			if err != nil {
				s.logger.Debug(err)
			}
		}

		if session.IsRevoked {
			s.error(w, r, http.StatusUnauthorized, errors.New("session revoked"))
			return
		}

		if session.UserEmail != refreshClaims.Email {
			s.error(w, r, http.StatusUnauthorized, errors.New("invalid session"))
			return
		}

		accessToken, accessClaims, err := s.tokenMaker.CreateToken(
			refreshClaims.ID,
			refreshClaims.Email,
			refreshClaims.IP,
			15*time.Minute,
		)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
		}

		res := &response{
			AccessToken:          accessToken,
			AccessTokenExpiresAt: accessClaims.RegisteredClaims.ExpiresAt.Time,
		}

		s.respond(w, r, http.StatusOK, res)
	}

}

// ----- helpers

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func hashRefreshToken(token string) (string, error) {
	const maxLen = 72
	var hashes []string

	for i := 0; i < len(token); i += maxLen {
		end := i + maxLen
		if end > len(token) {
			end = len(token)
		}
		part := token[i:end]

		hash, err := bcrypt.GenerateFromPassword([]byte(part), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		hashes = append(hashes, string(hash))
	}

	var buffer bytes.Buffer
	for _, hash := range hashes {
		buffer.WriteString(hash)
	}

	return buffer.String(), nil
}

func (s *server) sendEmailWarning(email string, ip string) error {
	auth := smtp.PlainAuth("",
		"workauthml@gmail.com",
		"z0of123laopL3rv",
		"smtp.gmail.com",
	)

	msg := "Warning: Another IP\nNew IP - " + ip

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		"workauthml@gmail.com",
		[]string{email},
		[]byte(msg),
	)
	if err != nil {
		return err
	}

	return nil
}
