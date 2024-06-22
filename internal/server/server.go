package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/SpaceSlow/loyalty/internal/config"
	"github.com/SpaceSlow/loyalty/internal/store"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func RunServer() error {
	rootCtx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	var err error
	config.ServerConfig, err = config.GetConfigWithFlags()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	g, ctx := errgroup.WithContext(rootCtx)
	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), config.ServerConfig.TimeoutServerShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	})

	db, err := store.NewDB(ctx, config.ServerConfig.DSN)
	if err != nil {
		return fmt.Errorf("failed to initialize a new DB: %w", err)
	}
	defer db.Close()

	g.Go(func() error {
		defer log.Print("closed DB")

		<-ctx.Done()

		db.Close()
		return nil
	})

	g.Go(func() error {
		orderNumbers, err := db.GetUnprocessedOrderAccruals(ctx)
		if err != nil {
			return err
		}
		
		for orderNumber := range orderNumbers {
			go CalculateAccrual(ctx, db, orderNumber)
		}
		return nil
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: Router(db),
	}

	g.Go(func() (err error) {
		defer func() {
			errRec := recover()
			if errRec != nil {
				err = fmt.Errorf("a panic occurred: %v", errRec)
			}
		}()
		if err = srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			return fmt.Errorf("listen and server has failed: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		defer log.Print("server has been shutdown")
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), config.ServerConfig.TimeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()
		if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
			log.Printf("an error occurred during server shutdown: %v", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Print(err)
	}

	return nil
}
