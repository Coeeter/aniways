package database

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/coeeter/aniways/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbRetryInterval = 2 * time.Second
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func New(ctx context.Context, env *config.Env, log *slog.Logger) (*pgxpool.Pool, error) {
	log.Info("initialising database connection")

	cfg, err := pgxpool.ParseConfig(env.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}

	isDev := env.AppEnv == "development"

	for attempt := 1; ; attempt++ {
		pingCtx, pingCancel := context.WithTimeout(ctx, 3*time.Second)
		pingErr := pool.Ping(pingCtx)
		pingCancel()
		if pingErr == nil {
			break
		}
		if !isDev {
			pool.Close()
			return nil, fmt.Errorf("db ping: %w", pingErr)
		}
		log.Warn("database not ready, retrying...",
			"attempt", attempt,
			"error", pingErr,
			"retry_in", dbRetryInterval,
		)
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("db ping cancelled: %w", ctx.Err())
		case <-time.After(dbRetryInterval):
		}
	}
	log.Info("database ping OK")

	d, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate source init: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, env.DatabaseURL)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate init: %w", err)
	}
	defer m.Close()

	switch err := m.Up(); err {
	case nil:
		log.Info("migrations applied")
	case migrate.ErrNoChange:
		log.Info("migrations already up‑to‑date")
	default:
		pool.Close()
		return nil, fmt.Errorf("migrate up: %w", err)
	}

	log.Info("database ready",
		"max_conns", cfg.MaxConns,
		"min_conns", cfg.MinConns,
	)
	return pool, nil
}
