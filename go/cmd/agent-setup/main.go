// Command agent-setup creates or updates the Managed Agent and its
// environment, and writes AGENT_ID and ENVIRONMENT_ID to the repo .env.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/sektor-ai/material-review-starter/go/internal/config"
	"github.com/sektor-ai/material-review-starter/go/internal/llm"
)

func main() {
	if err := run(context.Background()); err != nil {
		slog.Error("agent setup failed", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	envPath, ok := config.FindEnvFile()
	if !ok {
		return errors.New("no .env found; copy .env.example to .env first")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	spec := llm.ReviewerAgent
	if m := os.Getenv("ANTHROPIC_MODEL"); m != "" {
		spec.Model = m
	}

	client := llm.NewClient()
	envID, created, err := client.EnsureEnvironment(ctx, cfg.EnvironmentID)
	if err != nil {
		return err
	}
	slog.Info("environment ready", "id", envID, "created", created)

	agentID, version, err := client.EnsureAgent(ctx, cfg.AgentID, spec)
	if err != nil {
		return err
	}
	slog.Info("agent ready", "id", agentID, "version", version)

	if err := config.WriteEnvValues(envPath, map[string]string{
		"ENVIRONMENT_ID": envID,
		"AGENT_ID":       agentID,
	}); err != nil {
		return err
	}
	slog.Info("wrote AGENT_ID and ENVIRONMENT_ID", "file", envPath)
	return nil
}
