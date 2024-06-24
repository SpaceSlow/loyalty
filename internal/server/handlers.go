package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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

func (h *Handlers) RegisterUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := &model.User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	exist, err := h.store.CheckUsername(timeoutCtx, user.Username)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if exist {
		w.WriteHeader(http.StatusConflict)
		return
	}

	if user.GenerateHash() != nil {
		http.Error(w, errors.New("error occurred when generating password").Error(), http.StatusInternalServerError)
		return
	}

	timeoutCtx, cancel = context.WithTimeout(ctx, h.timeout)
	defer cancel()
	err = h.store.RegisterUser(timeoutCtx, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) LoginUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := &model.User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	passwordHash, err := h.store.GetPasswordHash(timeoutCtx, user.Username)
	var errNoUser *store.ErrNoUser
	if errors.As(err, &errNoUser) {
		http.Error(w, errNoUser.Error(), http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	check, err := user.CheckPassword(passwordHash)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !check {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	timeoutCtx, cancel = context.WithTimeout(ctx, h.timeout)
	defer cancel()
	userID, err := h.store.GetUserID(timeoutCtx, user.Username)
	if errors.As(err, &errNoUser) {
		http.Error(w, errNoUser.Error(), http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token, err := middleware.BuildJWTString(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) RegisterOrderNumber(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(ctx)
	orderNumber, err := getOrderNumber(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !isValidLuhnAlgorithm(orderNumber) {
		http.Error(w, (&ErrInvalidOrderNumber{orderNumber: orderNumber}).Error(), http.StatusUnprocessableEntity)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	err = h.store.RegisterOrderNumber(timeoutCtx, userID, orderNumber)
	var errOrderAlreadyExist *store.ErrOrderAlreadyExist
	if err != nil && errors.As(err, &errOrderAlreadyExist) && errOrderAlreadyExist.UserID == userID {
		w.WriteHeader(http.StatusOK)
		return
	} else if err != nil && errors.As(err, &errOrderAlreadyExist) && errOrderAlreadyExist.UserID != userID {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go CalculateAccrual(context.Background(), h.store, orderNumber)

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) GetAccrualInfos(ctx context.Context, w http.ResponseWriter, _ *http.Request) {
	userID := middleware.GetUserID(ctx)

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	accruals, err := h.store.GetUserAccruals(timeoutCtx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(accruals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	data, err := json.Marshal(accruals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
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

func (h *Handlers) WithdrawLoyaltyPoints(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(ctx)

	withdrawal := &model.WithdrawalInfo{}
	if err := json.NewDecoder(r.Body).Decode(withdrawal); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if withdrawal.Sum <= 0 {
		http.Error(w, (&ErrIncorrectWithdrawalSum{sum: withdrawal.Sum}).Error(), http.StatusBadRequest)
		return
	}

	orderNumber, err := strconv.Atoi(withdrawal.OrderNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isValidLuhnAlgorithm(orderNumber) {
		http.Error(w, (&ErrInvalidOrderNumber{orderNumber: orderNumber}).Error(), http.StatusUnprocessableEntity)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	balance, err := h.store.GetBalance(timeoutCtx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if withdrawal.Sum > balance.Current {
		http.Error(w, (&ErrNotEnoughLoyaltyPoints{current: balance.Current}).Error(), http.StatusPaymentRequired)
		return
	}

	timeoutCtx, cancel = context.WithTimeout(ctx, h.timeout)
	defer cancel()
	err = h.store.AddWithdrawal(ctx, userID, withdrawal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
