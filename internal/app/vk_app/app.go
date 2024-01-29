package vk_app

import (
	"context"
	"database/sql"
	"github.com/sirupsen/logrus"
	"main/internal/app/migrations"
	"main/internal/app/store/sqlstore"
	_ "modernc.org/sqlite"
	"path/filepath"
)

func Start(ctx context.Context, config *Config) (*App, error) {
	db, err := newDB(config.DatabaseURL)
	if err != nil {
		return nil, err
	}
	store := sqlstore.New(db)
	migrations.MakeMigrations(db, logrus.New())
	srv := newApp(ctx, &store, config.BindAddr, *config)
	return srv, nil
}

func newDB(dbURL string) (*sql.DB, error) {
	fn := filepath.Join("./", "db")

	db, err := sql.Open("sqlite", fn)
	//db, err := sql.Open("sqlite", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
