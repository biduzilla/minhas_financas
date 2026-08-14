package database

import (
	"context"
	"database/sql"
	"fmt"
	"ms_auth/internal/core/config"
	"time"

	"github.com/XSAM/otelsql"
)

func OpenDB(cfg config.Config) (*sql.DB, error) {
	driverName, err := otelsql.Register("postgres")
	if err != nil {
		return nil, fmt.Errorf("error with the otelsql registry driver: %w", err)
	}

	db, err := sql.Open(driverName, cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)

	duration, err := time.ParseDuration(cfg.DB.MaxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
