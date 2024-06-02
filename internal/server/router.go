package server

import (
	"github.com/SpaceSlow/loyalty/internal/store"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func Router(storage *store.DB) chi.Router {
	r := chi.NewRouter()

	h := NewHandlers(storage)

	r.Route("/", func(r chi.Router) {
		r.Post("/api/user/register", func(w http.ResponseWriter, r *http.Request) {
			h.RegisterUser(r.Context(), w, r)
		})
		r.Post("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
			h.LoginUser(r.Context(), w, r)
		})
	})

	return r
}
