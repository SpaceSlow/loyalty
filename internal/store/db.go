package store

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/SpaceSlow/loyalty/internal/model"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(ctx context.Context, dsn string) (*DB, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}
	return &DB{
		pool: pool,
	}, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func (db *DB) CheckUsername(ctx context.Context, username string) (bool, error) {
	row := db.pool.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT id FROM users WHERE username=$1)`,
		username,
	)
	var existUsername bool
	if err := row.Scan(&existUsername); err != nil {
		return false, fmt.Errorf("failed to check existing username: %w", err)
	}
	return existUsername, nil
}

func (db *DB) RegisterUser(ctx context.Context, u *model.User) error {
	_, err := db.pool.Exec(
		ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2)`,
		u.Username, u.PasswordHash,
	)
	return err
}

func (db *DB) GetPasswordHash(ctx context.Context, username string) (string, error) {
	row := db.pool.QueryRow(
		ctx,
		"SELECT password_hash FROM users WHERE username=$1",
		username,
	)

	var hash string
	err := row.Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", &ErrNoUser{username: username}
	} else if err != nil {
		return "", err
	}
	return hash, nil
}

func (db *DB) GetUserID(ctx context.Context, username string) (int, error) {
	row := db.pool.QueryRow(
		ctx,
		"SELECT id FROM users WHERE username=$1",
		username,
	)

	var userID int
	err := row.Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return -1, &ErrNoUser{username: username}
	} else if err != nil {
		return -1, err
	}
	return userID, nil
}

func (db *DB) RegisterOrderNumber(ctx context.Context, userID int, orderNumber int) error {
	row := db.pool.QueryRow(
		ctx,
		"SELECT user_id FROM accruals WHERE order_number=$1",
		orderNumber,
	)

	var storedUserID int
	err := row.Scan(&storedUserID)
	if !errors.Is(err, pgx.ErrNoRows) {
		return &ErrOrderAlreadyExist{UserID: storedUserID}
	}

	_, err = db.pool.Exec(
		ctx,
		`INSERT INTO accruals (user_id, order_number) VALUES ($1, $2)`,
		userID, orderNumber,
	)
	return err
}

func (db *DB) GetUnprocessedOrderAccruals(ctx context.Context) ([]int, error) {
	rows, err := db.pool.Query(ctx, "SELECT order_number FROM unprocessed_orders_view")

	if err != nil {
		return nil, err
	}

	orders := make([]int, 0)
	for rows.Next() {
		var orderNumber int
		err := rows.Scan(&orderNumber)
		if err != nil {
			return nil, err
		}
		orders = append(orders, orderNumber)
	}
	return orders, nil
}

func (db *DB) UpdateAccrualInfo(ctx context.Context, accrualInfo model.AccrualInfo) error {
	_, err := db.pool.Exec(
		ctx,
		`UPDATE accruals SET status=$1, sum=$2 WHERE order_number=$3`,
		accrualInfo.Status, accrualInfo.Sum, accrualInfo.OrderNumber,
	)
	return err
}

func (db *DB) Close() {
	db.pool.Close()
}
