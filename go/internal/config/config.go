// Package config loads settings from environment variables. Locally they come
// from the repo-root .env file; on Cloud Run they are injected by the service.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL  string
	Port         int
	CookieSecure bool

	S3 S3

	AgentID       string
	EnvironmentID string
}

type S3 struct {
	Endpoint        string // host[:port], no scheme
	UseSSL          bool
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

// Load reads the .env file (if one is found) into the process environment
// without overriding variables that are already set, then builds a Config.
func Load() (Config, error) {
	if path, ok := FindEnvFile(); ok {
		if err := loadEnvFile(path); err != nil {
			return Config{}, err
		}
	}

	var missing []string
	required := func(name string) string {
		v := os.Getenv(name)
		if v == "" {
			missing = append(missing, name)
		}
		return v
	}

	endpoint, useSSL := splitEndpoint(required("S3_ENDPOINT"))
	cfg := Config{
		DatabaseURL:  required("DATABASE_URL"),
		Port:         intOr("PORT", 8787),
		CookieSecure: os.Getenv("COOKIE_SECURE") == "true",
		S3: S3{
			Endpoint:        endpoint,
			UseSSL:          useSSL,
			Region:          stringOr("S3_REGION", "us-east-1"),
			AccessKeyID:     required("S3_ACCESS_KEY_ID"),
			SecretAccessKey: required("S3_SECRET_ACCESS_KEY"),
			Bucket:          required("S3_BUCKET"),
		},
		AgentID:       os.Getenv("AGENT_ID"),
		EnvironmentID: os.Getenv("ENVIRONMENT_ID"),
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing %s (see .env.example)", strings.Join(missing, ", "))
	}
	return cfg, nil
}

// FindEnvFile walks up from the working directory looking for a .env file.
func FindEnvFile() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		path := filepath.Join(dir, ".env")
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, set := os.LookupEnv(name); !set {
			os.Setenv(name, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	return nil
}

// WriteEnvValues sets name=value lines in the .env file at path, replacing
// existing lines for those names and appending the rest.
func WriteEnvValues(path string, values map[string]string) error {
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	for name, value := range values {
		replaced := false
		for i, line := range lines {
			if strings.HasPrefix(line, name+"=") {
				lines[i] = name + "=" + value
				replaced = true
			}
		}
		if !replaced {
			lines = append(lines, name+"="+value)
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)
}

func splitEndpoint(raw string) (host string, useSSL bool) {
	switch {
	case strings.HasPrefix(raw, "https://"):
		return strings.TrimPrefix(raw, "https://"), true
	case strings.HasPrefix(raw, "http://"):
		return strings.TrimPrefix(raw, "http://"), false
	default:
		return raw, true
	}
}

func stringOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func intOr(name string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(name)); err == nil {
		return v
	}
	return fallback
}
