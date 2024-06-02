package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/SpaceSlow/loyalty/internal/config"
	"github.com/SpaceSlow/loyalty/internal/model"
	"github.com/SpaceSlow/loyalty/internal/store"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func BuildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.ServerConfig.TokenExpiredAt)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(config.ServerConfig.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

type Handlers struct {
	store   *store.DB
	timeout time.Duration
}

func NewHandlers(s *store.DB) *Handlers {
	return &Handlers{
		store:   s,
		timeout: 3 * time.Second,
	}
}

func (h *Handlers) RegisterUser(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	user := &model.User{}
	if err := json.NewDecoder(req.Body).Decode(user); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if err := req.Body.Close(); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	exist, err := h.store.CheckUsername(timeoutCtx, user.Username)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if exist {
		res.WriteHeader(http.StatusConflict)
		return
	}

	if user.GenerateHash() != nil {
		http.Error(res, errors.New("error occured when generating password").Error(), http.StatusInternalServerError)
		return
	}

	timeoutCtx, cancel = context.WithTimeout(ctx, h.timeout)
	defer cancel()
	err = h.store.CreateUser(timeoutCtx, user)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
}

func (h *Handlers) LoginUser(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	user := &model.User{}
	if err := json.NewDecoder(req.Body).Decode(user); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if err := req.Body.Close(); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	passwordHash, err := h.store.GetPasswordHash(timeoutCtx, user.Username)
	var errNoUser *store.ErrNoUser
	if errors.As(err, &errNoUser) {
		res.WriteHeader(http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	check, err := user.CheckPassword(passwordHash)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if !check {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	timeoutCtx, cancel = context.WithTimeout(ctx, h.timeout)
	defer cancel()
	userID, err := h.store.GetUserID(timeoutCtx, user.Username)
	if errors.As(err, &errNoUser) {
		res.WriteHeader(http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	token, err := BuildJWTString(userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Authorization", token)
	res.WriteHeader(http.StatusOK)
}
