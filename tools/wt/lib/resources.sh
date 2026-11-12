#!/usr/bin/env bash
# Per-environment data resources for `wt`: one Postgres database and one object-storage bucket each,
# carved out of the local Postgres and MinIO that docker-compose.yml runs.
#
# The work is done by go/cmd/devenv, run from the primary checkout against the primary's .env (the
# "source env"). It derives the names — database <base>_wt_<slug_with_underscores>, bucket
# <base>-wt-<slug> — and refuses to touch anything but local services. Keeping that logic in Go means
# one place parses DATABASE_URL, and no psql or mc is needed.
#
# A fresh database is empty: `wt create` and `wt reset-db` then run the worktree's own migrations and
# seed (demo users, bucket) into it, so the schema always matches the branch's code.

# devenv runs go/cmd/devenv from the primary checkout. Its stdout is dotenv lines (DATABASE_NAME,
# DATABASE_URL, S3_BUCKET); its logs go to stderr.
devenv() {
	local primary env_file
	primary=$(primary_root)
	env_file=$(source_env_file)
	[[ -f $env_file ]] ||
		die "missing ${env_file} — copy .env.example to .env in the primary checkout first"
	(cd "$primary/go" && ENV_FILE="$env_file" go run ./cmd/devenv "$@")
}

# dotenv_value prints the value of KEY from dotenv text held in a variable.
dotenv_value() {
	local text=$1 key=$2
	printf '%s\n' "$text" | awk -F= -v key="$key" '$1 == key { sub(/^[^=]*=/, ""); print; exit }'
}

# resources_create creates an environment's database (idempotent) and prints devenv's dotenv lines.
resources_create() { devenv create "$1"; }

# resources_names prints an environment's dotenv lines without creating anything.
resources_names() { devenv names "$1"; }

# resources_destroy drops an environment's database and bucket. Never fatal: `wt gc` runs unattended
# from the post-merge hook, so stopped services must not block teardown. Worst case the database
# outlives the environment, and running `wt rm` again (or devenv destroy) finishes it off.
resources_destroy() {
	local slug=$1
	devenv destroy "$slug" >/dev/null 2>&1 ||
		warn "could not drop ${slug}'s database or bucket (are Postgres and MinIO up?); finish with: (cd $(primary_root)/go && go run ./cmd/devenv destroy ${slug})"
}

# resources_seed runs the worktree's own migrations and seed against the environment's database and
# bucket. It uses the worktree's code, not the primary's, so a branch that adds a migration gets its
# schema.
resources_seed() {
	local slug=$1 path env_file
	path=$(state_get "$slug" '.path')
	is_promoted "$slug" && path=$(primary_root)
	env_file="$(state_dir "$slug")/app.env"

	log 'migrating and seeding the environment database'
	(cd "$path/go" && ENV_FILE="$env_file" go run ./cmd/migrate && ENV_FILE="$env_file" go run ./cmd/seed) ||
		die "migrate/seed failed for ${slug}; fix it, then: wt reset-db ${slug}"
}
