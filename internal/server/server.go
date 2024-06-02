package server

import (
	"net/http"

	"github.com/SpaceSlow/loyalty/internal/config"
	"github.com/SpaceSlow/loyalty/internal/store"
)

func RunServer() error {
	var err error
	config.ServerConfig, err = config.GetConfigWithFlags()
	if err != nil {
		return err
	}
	db, err := store.NewDB(config.ServerConfig.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	return http.ListenAndServe(":8080", Router(db))
}
