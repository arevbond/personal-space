package storage

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"github.com/arevbond/arevbond-blog/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	DB  *sqlx.DB
	log *slog.Logger
}

func New(log *slog.Logger, cfg config.Storage) (*Storage, error) {
	hostWithPort := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	uri := fmt.Sprintf("postgresql://%s:%s@%s/%s", cfg.User, cfg.Password,
		hostWithPort, cfg.DatabaseName)
	connStr, err := pgx.ParseConfig(uri)

	if err != nil {
		return nil, fmt.Errorf("can't parse pg uri: %w", err)
	}

	pgxdb := stdlib.OpenDB(*connStr)

	if err = pgxdb.Ping(); err != nil {
		return nil, fmt.Errorf("can't ping db: %w", err)
	}

	return &Storage{
		DB:  sqlx.NewDb(pgxdb, "pgx"),
		log: log,
	}, nil
}
