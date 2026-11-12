# Document Review

## The problem

Engineering teams review material submittals from contractors and check them against the project specification. Every item that does not comply is an issue that must be resolved before the material is approved and work proceeds.

We're building a system that combines an AI agent for reviewing material submittals with a shared table for managing its issues. The agent checks each submittal against the specification, identifies issues and adds them to the table, linking each issue to its submittal, its source document and the specification clause it breaks, and including a description, severity, status (open/close), and other metadata.

Several engineers may use the table at the same time to update issues. They should see each other's changes without refreshing.

We must maintain a clear record of who changed what.

## The specification: QCS 2014

**Submittals are checked against QCS 2014 (Qatar Construction Specifications).** It is the only specification. An issue is any place where a submittal does not meet a QCS 2014 clause, and every issue cites that clause (section, part, clause number).

For example, the S-001 mix design proposes a water/cementitious ratio of 0.55 for Grade C40 concrete. QCS 2014 Section 05 Part 06, Table 6.4 allows at most 0.50. That is one issue.

QCS 2014 is copyrighted, so it is not in this repo. Get your own copy and save it as `spec/QCS-2014.pdf`. Git ignores that folder's PDFs, so the copy stays on your machine.

The parts the sample submittals need are below. The page numbers are for the single 4,556-page PDF of the full QCS 2014. If your copy is split into one PDF per part, use the section and part names instead.

| Submittal                          | QCS 2014 part                                          | PDF pages |
|------------------------------------|--------------------------------------------------------|-----------|
| All                                | Section 01 Part 07 Submittals (7.11 Materials Approvals) | 29–47     |
| S-001 Concrete mix design          | Section 05 Part 02 Aggregates                          | 624–633   |
| S-001 Concrete mix design          | Section 05 Part 05 Admixtures                          | 644–648   |
| S-001 Concrete mix design          | Section 05 Part 06 Property Requirements               | 649–679   |
| S-001 Concrete mix design          | Section 05 Part 10 Curing                              | 719–724   |
| S-002 Fire-rated doors             | Section 18 Part 05 Joinery (5.3 Timber Door Types)     | 3549–3554 |
| S-002 Fire-rated doors             | Section 18 Part 07 Ironmongery                         | 3558–3564 |

The PDF's bookmarks list every section and part with its page range.

## Sample submittals

Each folder in `samples/` is one material submittal from the contractor. Each holds one document: the submittal itself, with the supplier's data. Each submittal states the requirements it was designed for (grade, exposure class, fire rating) in its "Design basis" line. Everything else comes from QCS 2014.

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

Sample submittals are in `samples/`. The specification they are checked against is QCS 2014, which you save yourself as `spec/QCS-2014.pdf` (see above).

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
3. **Upload a file.** `client.beta.files.upload({ file })` returns a file id. Upload `spec/QCS-2014.pdf` once and reuse its file id in every session.
4. **Start a session.** `client.beta.sessions.create({ agent: AGENT_ID, environment_id: ENVIRONMENT_ID, resources: [{ type: "file", file_id, mount_path: "/qcs-2014.pdf" }, { type: "file", file_id: submittalFileId, mount_path: "/submittal.md" }] })`. Mounted files appear under `/mnt/session/uploads/`, so `/qcs-2014.pdf` is read at `/mnt/session/uploads/qcs-2014.pdf`.
5. **Follow it.** Open `client.beta.sessions.events.stream(sessionId)` first, then send the prompt with `client.beta.sessions.events.send(sessionId, { events: [{ type: "user.message", content: [...] }] })`. The session is done when you get `session.status_idle` with `stop_reason.type` other than `"requires_action"`, or `session.status_terminated`.
6. **Get results back.** Pick one:
   - the text of the last `agent.message` event;
   - files the agent writes to `/mnt/session/outputs/`, listed with `client.beta.files.list({ scope_id: sessionId, betas: ["managed-agents-2026-04-01"] })` (keep the ones with `downloadable: true`) and fetched with `client.beta.files.download(id)`. Output files can take a few seconds to appear after the session goes idle, so retry the list briefly;
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
