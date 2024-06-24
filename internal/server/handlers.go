package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/SpaceSlow/loyalty/internal/config"
	"github.com/SpaceSlow/loyalty/internal/middleware"
	"github.com/SpaceSlow/loyalty/internal/model"
	"github.com/SpaceSlow/loyalty/internal/store"
)

type Handlers struct {
	store   *store.DB
	timeout time.Duration
}

func NewHandlers(s *store.DB) *Handlers {
	return &Handlers{
		store:   s,
		timeout: config.ServerConfig.TimeoutOperation,
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
		http.Error(res, errors.New("error occurred when generating password").Error(), http.StatusInternalServerError)
		return
	}

	timeoutCtx, cancel = context.WithTimeout(ctx, h.timeout)
	defer cancel()
	err = h.store.RegisterUser(timeoutCtx, user)
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
		http.Error(res, errNoUser.Error(), http.StatusUnauthorized)
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
		http.Error(res, errNoUser.Error(), http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	token, err := middleware.BuildJWTString(userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Authorization", token)
	res.WriteHeader(http.StatusOK)
}

func (h *Handlers) RegisterOrderNumber(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	userID := middleware.GetUserID(ctx)
	orderNumber, err := getOrderNumber(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if !isValidLuhnAlgorithm(orderNumber) {
		http.Error(res, (&ErrInvalidOrderNumber{orderNumber: orderNumber}).Error(), http.StatusUnprocessableEntity)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	err = h.store.RegisterOrderNumber(timeoutCtx, userID, orderNumber)
	var errOrderAlreadyExist *store.ErrOrderAlreadyExist
	if err != nil && errors.As(err, &errOrderAlreadyExist) && errOrderAlreadyExist.UserID == userID {
		res.WriteHeader(http.StatusOK)
		return
	} else if err != nil && errors.As(err, &errOrderAlreadyExist) && errOrderAlreadyExist.UserID != userID {
		http.Error(res, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	go CalculateAccrual(context.Background(), h.store, orderNumber)

	res.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) GetAccrualInfos(ctx context.Context, res http.ResponseWriter, _ *http.Request) {
	userID := middleware.GetUserID(ctx)

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	accruals, err := h.store.GetUserAccruals(timeoutCtx, userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(accruals) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}
	data, err := json.Marshal(accruals)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(data)
}

func (h *Handlers) GetBalance(ctx context.Context, w http.ResponseWriter, _ *http.Request) {
	userID := middleware.GetUserID(ctx)

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	balance, err := h.store.GetBalance(timeoutCtx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
