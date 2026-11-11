# Document Review

## The problem

Engineering teams review technical documents submitted for approval to identify issues that need resolving before work proceeds.

We're building a system that combines an AI agent for reviewing submittals of documents with a shared table for managing its issues. The agent identifies issues and adds them to the table, linking each issue to its submittal and source document and including a description, severity, status (open/close), and other metadata.

Several engineers may use the table at the same time to update issues. They should see each other's changes without refreshing.

We must maintain a clear record of who changed what.

## Prerequisites

- Node.js 22 or 24 (LTS)
- Docker
- An Anthropic API key

## Quick start

```bash
npm install
cp .env.example .env        # then set ANTHROPIC_API_KEY (and BETTER_AUTH_SECRET)
npm run setup               # starts Postgres + MinIO, runs migrations, seeds users
npm run agent:setup         # creates the agent + environment, writes their ids to .env
npm run dev                 # API on :8787, web on :5180
```

Open http://localhost:5180 and sign in:

| User  | Email             | Password    |
|-------|-------------------|-------------|
| Alice | alice@example.com | password123 |
| Bob   | bob@example.com   | password123 |

App tables are yours to design. The login tables live in the `auth` schema, so `public` is empty. Use any ORM or plain SQL against `DATABASE_URL`.

Sample submittals are in `samples/`.

`npm run setup` is safe to re-run. If the database gets reset, run it again to restore Alice and Bob.

## Scripts

| Script                 | What it does                                             |
|------------------------|----------------------------------------------------------|
| `npm run setup`        | `docker compose up`, then `db:migrate` and `db:seed`      |
| `npm run dev`          | Runs the API and the web app together                    |
| `npm run db:generate`  | Generates a Drizzle migration from the schema            |
| `npm run db:migrate`   | Applies Drizzle migrations                               |
| `npm run db:seed`      | Creates the storage bucket and the two users             |
| `npm run agent:setup`  | Creates or updates the agent and environment             |
| `npm run typecheck`    | Type-checks both apps                                    |

## Ports

| Service        | URL                                                  |
|----------------|------------------------------------------------------|
| Web (Vite)     | http://localhost:5180 (proxies `/api` to the API)    |
| API (Hono)     | http://localhost:8787                                |
| Postgres 16    | localhost:5433, user `postgres`, password `postgres` |
| MinIO (S3 API) | http://localhost:9000                                |
| MinIO console  | http://localhost:9001, `minioadmin` / `minioadmin`   |

## Agent

The agent runs on Claude Managed Agents: Anthropic hosts the agent loop and a sandbox per session. Calls go through `@anthropic-ai/sdk` under `client.beta`.

1. **Ids.** `npm run agent:setup` (`apps/api/src/scripts/agent-setup.ts`) creates the agent and its environment once and writes `AGENT_ID` and `ENVIRONMENT_ID` to `.env`. Re-running it updates the same agent instead of creating a new one.
2. **Instructions.** The agent starts with a placeholder system prompt and the built-in toolset (read, write, bash and so on; web tools off). To change it, edit `agentConfig` in the setup script and re-run `npm run agent:setup`, or call `client.beta.agents.update(AGENT_ID, { system, tools })`. Each change creates a new agent version.
3. **Upload a file.** `client.beta.files.upload({ file })` returns a file id.
4. **Start a session.** `client.beta.sessions.create({ agent: AGENT_ID, environment_id: ENVIRONMENT_ID, resources: [{ type: "file", file_id, mount_path: "/spec.md" }] })`. Mounted files appear under `/mnt/session/uploads/`, so `/spec.md` is read at `/mnt/session/uploads/spec.md`.
5. **Follow it.** Open `client.beta.sessions.events.stream(sessionId)` first, then send the prompt with `client.beta.sessions.events.send(sessionId, { events: [{ type: "user.message", content: [...] }] })`. The session is done when you get `session.status_idle` with `stop_reason.type` other than `"requires_action"`, or `session.status_terminated`.
6. **Get results back.** Pick one:
   - the text of the last `agent.message` event;
   - files the agent writes to `/mnt/session/outputs/`, listed with `client.beta.files.list({ scope_id: sessionId, betas: ["managed-agents-2026-04-01"] })` (keep the ones with `downloadable: true`) and fetched with `client.beta.files.download(id)`;
   - a custom tool (`{ type: "custom", name, description, input_schema }` in the agent's `tools`): the agent emits `agent.custom_tool_use`, your code handles it and replies with `user.custom_tool_result`.

Every session also shows up in the Claude Console under Managed Agents → Sessions, with the full transcript.

Docs:
[Overview](https://platform.claude.com/docs/en/managed-agents/overview) ·
[Quickstart](https://platform.claude.com/docs/en/managed-agents/quickstart) ·
[Agent setup](https://platform.claude.com/docs/en/managed-agents/agent-setup) ·
[Sessions](https://platform.claude.com/docs/en/managed-agents/sessions) ·
[Events and streaming](https://platform.claude.com/docs/en/managed-agents/events-and-streaming) ·
[Files](https://platform.claude.com/docs/en/managed-agents/files) ·
[Tools](https://platform.claude.com/docs/en/managed-agents/tools)
