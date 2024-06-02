package server

import (
	"github.com/SpaceSlow/loyalty/internal/store"
	"net/http"

	"github.com/SpaceSlow/loyalty/internal/config"
)

func RunServer() error {
	cfg, err := config.GetConfigWithFlags()
	if err != nil {
		return err
	}
	db, err := store.NewDB(cfg.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	return http.ListenAndServe(":8080", Router(db))
}
