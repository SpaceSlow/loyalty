package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SpaceSlow/loyalty/internal/model"
	"net/http"
	"time"

	"github.com/SpaceSlow/loyalty/internal/store"
)

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

	if timeoutCtx.Err() != nil {
		http.Error(res, errors.New("timeout exceeded").Error(), http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if exist {
		res.WriteHeader(http.StatusConflict)
		return
	}

	//TODO register and return OK-200
}

func (h *Handlers) LoginUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}
