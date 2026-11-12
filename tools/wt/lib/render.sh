#!/usr/bin/env bash
# Config rendering for `wt`.
#
# Two files are generated into the environment's state directory:
#
#   app.env              the primary's .env, copied and rewritten to this environment's database,
#                        bucket and port. Every Go command finds it through ENV_FILE. Copied rather
#                        than templated, so the secrets it carries (ANTHROPIC_API_KEY) stay in files
#                        nobody commits, and a key added to the primary's .env reaches environments on
#                        the next `wt recreate-config`.
#   process-compose.yml  rendered from templates/process-compose.yml.tpl.

readonly TEMPLATE_DIR_NAME='templates'

# render_template expands only the WT_* variables, leaving any other $-sign in the template alone.
render_template() {
	local template=$1 output=$2
	local vars='${WT_PATH} ${WT_STATE} ${WT_SLUG} ${WT_API_PORT} ${WT_WEB_PORT}'
	envsubst "$vars" <"$template" >"$output"
}

# rewrite_env_var replaces a KEY=... line in a dotenv file, appending it when absent. The value is
# always double-quoted (the Go loader strips the quotes), so the caller never has to know how the
# source file happened to quote that key. awk is used rather than sed because sed treats & and \ in the
# replacement text as backreferences.
rewrite_env_var() {
	local file=$1 key=$2 value=$3 tmp
	tmp=$(mktemp)

	awk -v key="$key" -v value="$value" '
		$0 ~ "^" key "=" { print key "=\"" value "\""; found = 1; next }
		{ print }
		END { if (!found) print key "=\"" value "\"" }
	' "$file" >"$tmp" && mv "$tmp" "$file"
}

# render_env renders app.env and process-compose.yml for an environment. `names` is devenv's dotenv
# output for it (DATABASE_URL, S3_BUCKET).
render_env() {
	local slug=$1 names=$2 tools_dir=$3
	local dir path slot src app_env
	dir=$(state_dir "$slug")
	path=$(state_get "$slug" '.path')
	slot=$(state_get "$slug" '.slot')
	src=$(source_env_file)
	[[ -f $src ]] || die "missing ${src} — cannot render the environment's settings"

	export WT_SLUG="$slug"
	export WT_PATH="$path"
	export WT_STATE="$dir"
	export WT_API_PORT && WT_API_PORT=$(port_api "$slot")
	export WT_WEB_PORT && WT_WEB_PORT=$(port_web "$slot")

	app_env="$dir/app.env"
	cp "$src" "$app_env"
	chmod 600 "$app_env"
	rewrite_env_var "$app_env" DATABASE_URL "$(dotenv_value "$names" DATABASE_URL)"
	rewrite_env_var "$app_env" S3_BUCKET "$(dotenv_value "$names" S3_BUCKET)"
	rewrite_env_var "$app_env" PORT "$WT_API_PORT"
	rewrite_env_var "$app_env" COOKIE_SECURE false

	render_template "$tools_dir/$TEMPLATE_DIR_NAME/process-compose.yml.tpl" "$dir/process-compose.yml"
}
