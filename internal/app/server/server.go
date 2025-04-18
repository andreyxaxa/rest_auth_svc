package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andreyxaxa/rest_auth_svc/internal/app/models"
	"github.com/andreyxaxa/rest_auth_svc/internal/app/storage"
	"github.com/andreyxaxa/rest_auth_svc/internal/app/token"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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
	s.router.HandleFunc("/users/login", s.handleUsersLogin()).Methods("POST")
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

		u, err := s.storage.User().Find(req.Email)
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
		// TODO: зашифровать здесь рефреш токен, положить в базу хэш

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
