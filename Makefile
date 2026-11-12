# Single entry point for the monorepo. Run from the repo root.
.DEFAULT_GOAL := help
ENV ?= dev

.PHONY: help
help: ## List targets
	@grep -hE '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-14s %s\n", $$1, $$2}'

.PHONY: install
install: ## Install TS dependencies and Go modules
	cd ts && pnpm install
	cd go && go mod download

.PHONY: setup
setup: ## Start Postgres + MinIO, migrate, seed users (safe to re-run)
	docker compose up -d --wait
	cd go && go run ./cmd/migrate
	cd go && go run ./cmd/seed

.PHONY: agent-setup
agent-setup: ## Create/update the Managed Agent and write its ids to .env
	cd go && go run ./cmd/agent-setup

.PHONY: gen
gen: ## Generate Go and TS code from proto/
	cd proto && buf generate

.PHONY: dev
dev: ## Run API (:8787) and web (:5180) together
	@trap 'kill 0' INT TERM EXIT; \
	(cd go && go run ./cmd/api) & \
	(cd ts && pnpm dev) & \
	wait

.PHONY: lint
lint: ## Lint proto, Go, TS and Terraform
	cd proto && buf lint
	cd go && go vet ./...
	@test -z "$$(cd go && gofmt -l cmd internal)" || (echo "gofmt needed:"; cd go && gofmt -l cmd internal; exit 1)
	cd ts && pnpm typecheck
	terraform -chdir=terraform fmt -check -recursive

.PHONY: breaking
breaking: ## Check proto/ for breaking changes against main
	cd proto && buf breaking --against '../.git#branch=main,subdir=proto'

.PHONY: test
test: ## Run Go and TS tests
	cd go && go test ./...
	cd ts && pnpm test

.PHONY: build
build: ## Build the API binary and the web bundle
	cd go && go build -o bin/api ./cmd/api
	cd ts && pnpm build

.PHONY: check-gen
check-gen: gen ## Fail if generated code is out of date (CI)
	git diff --exit-code -- go/gen ts/packages/api/src/gen

.PHONY: tf-plan
tf-plan: ## terraform plan for ENV (default dev). Never apply from here.
	terraform -chdir=terraform/envs/$(ENV) init
	terraform -chdir=terraform/envs/$(ENV) plan
