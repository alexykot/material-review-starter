// Command seed creates the storage bucket and the demo users. Safe to re-run.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/sektor-ai/material-review-starter/go/internal/config"
	"github.com/sektor-ai/material-review-starter/go/internal/service"
	"github.com/sektor-ai/material-review-starter/go/internal/storage"
	"github.com/sektor-ai/material-review-starter/go/internal/store"
)

var engineers = []struct{ Name, Email, Password string }{
	{"Alice Morgan", "alice@example.com", "password123"},
	{"Bob Chen", "bob@example.com", "password123"},
}

func main() {
	if err := run(context.Background()); err != nil {
		slog.Error("seed failed", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	stg, err := storage.New(cfg.S3)
	if err != nil {
		return err
	}
	if err := stg.EnsureBucket(ctx); err != nil {
		return err
	}
	slog.Info("bucket ready", "bucket", cfg.S3.Bucket)

	st, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()
	auth := service.NewAuth(st)

	for _, e := range engineers {
		_, err := st.UserByEmail(ctx, e.Email)
		if err == nil {
			slog.Info("user already exists", "email", e.Email)
			continue
		}
		if !errors.Is(err, store.ErrNotFound) {
			return err
		}
		if _, err := auth.CreateUser(ctx, e.Name, e.Email, e.Password); err != nil {
			return err
		}
		slog.Info("created user", "email", e.Email)
	}
	return nil
}
