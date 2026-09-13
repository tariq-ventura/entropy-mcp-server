package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	ConnectionString string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
}

type Database struct {
	gorm *gorm.DB
	sql  *sql.DB
}

func Open(ctx context.Context, cfg Config) (*Database, error) {
	cfg.ConnectionString = strings.TrimSpace(cfg.ConnectionString)
	if cfg.ConnectionString == "" {
		return nil, errors.New("MCP_DB_STRING is required")
	}
	if cfg.MaxOpenConns <= 0 {
		cfg.MaxOpenConns = 10
	}
	if cfg.MaxIdleConns < 0 {
		cfg.MaxIdleConns = 0
	}
	if cfg.MaxIdleConns > cfg.MaxOpenConns {
		cfg.MaxIdleConns = cfg.MaxOpenConns
	}
	if cfg.ConnMaxLifetime <= 0 {
		cfg.ConnMaxLifetime = 30 * time.Minute
	}

	gormDB, err := gorm.Open(postgres.Open(cfg.ConnectionString), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("opening MCP PostgreSQL database: %w", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("obtaining MCP PostgreSQL connection pool: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("pinging MCP PostgreSQL database: %w", err)
	}
	return &Database{gorm: gormDB, sql: sqlDB}, nil
}

func (d *Database) GORM() *gorm.DB { return d.gorm }

func (d *Database) Ping(ctx context.Context) error {
	if d == nil || d.sql == nil {
		return errors.New("MCP PostgreSQL database is not initialized")
	}
	return d.sql.PingContext(ctx)
}

func (d *Database) Close() error {
	if d == nil || d.sql == nil {
		return nil
	}
	return d.sql.Close()
}
