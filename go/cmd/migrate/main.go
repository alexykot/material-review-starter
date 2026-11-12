// Command migrate applies pending database migrations.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/sektor-ai/material-review-starter/go/internal/config"
	"github.com/sektor-ai/material-review-starter/go/internal/store"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}
	st, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		fail(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		fail(err)
	}
	slog.Info("migrations applied")
}

func fail(err error) {
	slog.Error("migrate failed", "err", err)
	os.Exit(1)
}
