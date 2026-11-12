# CLAUDE.md

Material submittal review: an AI agent checks contractor submittals against QCS 2014 and records issues in a shared table. Go backend, React frontend, Protobuf (Connect) API between them, deployed to GCP with Terraform. The backend calls the Anthropic API (Claude Managed Agents). Product context is in README.md.

## Layout

Each top-level folder is one language or toolchain. Do not mix languages inside a folder.

```
proto/          Protobuf API definitions: the single source of truth for backend/frontend contracts
  buf.yaml, buf.gen.yaml
  <domain>/v1/*.proto
go/             Go module (backend), one go.mod
  cmd/api/        HTTP server
  cmd/migrate/    applies DB migrations
  cmd/seed/       creates the storage bucket and demo users (idempotent)
  cmd/agent-setup/  creates/updates the Managed Agent, writes ids to .env
  internal/
    config/       env-based config; loads the repo-root .env locally
    server/       Connect handlers: decode, call service, encode. No business logic.
    service/      business logic
    store/        Postgres (pgx) + embedded goose migrations in store/migrations/
    storage/      S3-compatible object storage (MinIO locally, GCS interop on GCP)
    llm/          the only package that imports the Anthropic SDK
  gen/            generated from proto/. Do not edit.
  Dockerfile
ts/             pnpm workspace (frontend)
  apps/web/       React + Vite
  packages/api/   generated TS from proto/ (src/gen/, do not edit) re-exported as @app/api
terraform/
  modules/        reusable modules (api = Cloud Run service + SA + secret access)
  envs/dev/       one root module per environment, state in GCS
samples/        sample submittals (input data, not code)
spec/           QCS 2014 PDF, local only (gitignored, copyrighted)
docker-compose.yml  local Postgres + MinIO
```

## Commands

Run from the repo root through `make` (`make help` lists everything):

```
make install      # pnpm install + go mod download
make setup        # docker compose up, migrate, seed (safe to re-run)
make agent-setup  # create/update the Managed Agent
make dev          # API on :8787, web on :5180 (Vite proxies /api to the API)
make gen          # buf generate: proto/ -> go/gen and ts/packages/api/src/gen
make lint         # buf lint, go vet, gofmt, tsc, terraform fmt
make test         # go test ./... and vitest
make breaking     # buf breaking against main
make tf-plan ENV=dev
```

Before finishing a change, run `make lint test`. If proto/ changed, run `make gen` and `make breaking` too.

## API contract (proto/)

- Connect RPC: `connectrpc.com/connect` (v1) in Go, `@connectrpc/connect` + `connect-web` in TS. It works from the browser over plain HTTP, so no gRPC-Web proxy.
- Connect services are mounted under `/api` (e.g. `POST /api/auth.v1.AuthService/SignIn`). `/healthz` is plain HTTP.
- Packages are versioned: `proto/<domain>/v1/*.proto`, package `<domain>.v1`.
- Change the contract in proto/ first, then `make gen`, then implement. Never hand-write Go or TS types that duplicate a proto message.
- Generated code is committed; `make check-gen` fails if it is stale.
- buf plugins are pinned in buf.gen.yaml. The Connect Go plugin stays on v1.x to match `connectrpc.com/connect` v1. Moving to connect-go v2 is a deliberate migration, not a version bump.
- No breaking changes to a published `v1` service: add fields; never renumber or reuse field numbers; `reserved` removed ones.

## Go backend (go/)

- Wrap errors with `fmt.Errorf("doing x: %w", err)`. Map to Connect codes only in `internal/server`; internal errors are logged and returned as a generic `CodeInternal`.
- Pass `context.Context` first through every layer.
- Log with `log/slog` (JSON on Cloud Run). Never log request bodies, prompts with user data, passwords, or tokens.
- Config comes only from environment variables. Locally `config.Load` reads the repo-root `.env` without overriding already-set vars.
- Schema changes: add a new numbered file in `internal/store/migrations/` (goose format). Never edit an applied migration.
- Tests: table-driven, next to the code. Services depend on small interfaces (e.g. `service.UserStore`) so tests use fakes. Tests never call the real Anthropic API.

### Auth

Email/password with bcrypt. Sessions are random tokens in an HttpOnly `session` cookie, and only their SHA-256 is stored (`auth.sessions`). `COOKIE_SECURE=true` behind HTTPS. Demo users come from `cmd/seed`.

## Anthropic API

- Only the backend calls Anthropic. The API key never reaches the browser, and the frontend never imports an Anthropic SDK.
- Go SDK: `github.com/anthropics/anthropic-sdk-go/v2`, used only in `internal/llm`.
- The agent runs on Claude Managed Agents (`client.Beta.Agents/Environments/Sessions`). Agents are persistent: create them in `cmd/agent-setup`, never in the request path; the API uses `AGENT_ID` / `ENVIRONMENT_ID` from config.
- Agent definition (name, model, system prompt, tools) lives in `llm.ReviewerAgent`. Change it there and re-run `make agent-setup`.
- Default model `claude-opus-5-5` (`llm.DefaultModel`, overridable with `ANTHROPIC_MODEL`). Don't hardcode model IDs elsewhere or add date suffixes.
- For direct Messages API calls: use adaptive thinking and `output_config.effort` (`budget_tokens` and disabled thinking are rejected on current models); stream long requests; check `stop_reason` (`max_tokens`, `tool_use`, `refusal`) before reading content; keep system prompt and tools stable and first for prompt caching; handle SDK errors by type, not by message text.
- Don't guess SDK names. Check the SDK source or docs, and let the compiler confirm.

## TypeScript frontend (ts/)

- pnpm workspace; Node version in `ts/.nvmrc`. Strict TS; no `any` without a comment saying why.
- Import API types and service descriptors from `@app/api` only.
- Server state goes through TanStack Query hooks wrapping the Connect client (see `src/session.ts`). Components don't call the client directly.
- Tests: vitest next to the component.

## Terraform (terraform/)

- Cloud Run (API), Artifact Registry (images), Secret Manager (`ANTHROPIC_API_KEY`, `DATABASE_URL`, S3 HMAC keys), and a GCS documents bucket reached through the S3 interop endpoint.
- Terraform creates secret containers only. Values are added with `gcloud secrets versions add`, never through tfvars or state.
- State lives in a GCS bucket (`-backend-config="bucket=..."`), one prefix per env.
- CI authenticates with Workload Identity Federation. No service account keys anywhere.
- Never run `terraform apply`. Run `make tf-plan` and show the output; a person applies.
- Every resource gets `labels` with `app` and `env`.

## Rules

- Never commit `.env`, credentials, keys, or `spec/*.pdf`.
- Ask before adding a top-level folder or a new language.
- Keep this file current when a command, folder, or convention changes.
