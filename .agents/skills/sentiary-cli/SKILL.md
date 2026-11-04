---
name: sentiary-cli
description: Use sentiary-cli to manage Sentiary projects, languages, members, invitations, strings, and translations, or configure and download localization resources. Use for sentiary.toml, init/pull, offline restoration, CI downloads, Android/Compose/Xcode exports, credentials, and CLI installation or updates—even when the user only asks to sync translations.
---

# Sentiary CLI

Use `sentiary-cli` for both remote API operations and configured resource downloads. Choose the workflow before selecting credentials: API commands and `pull` use different credential rules.

## Requirements

- Check that `sentiary-cli` is on `PATH` and inspect `sentiary-cli version --json`.
- Check `sentiary-cli COMMAND --help` before relying on options; the installed version may differ from current documentation.
- Never print secrets, put them in command arguments, or write them to tracked files. Prefer environment references over literal TOML keys. Do not dump the environment or credential files to diagnose authentication.
- If the API key is missing, ask the user to configure it. Do not ask the user to paste it.

Put flags after the command. There are no global `--json` or server flags and no positional command arguments. Boolean flags use `--flag` or `--flag=false`, not `--flag false`. Quote translation text to preserve shell-sensitive characters.

### Credentials

- **API commands** (projects, strings, import/export, etc.) require `SENTIARY_USER_API_KEY`. Project selection is `--project-id`, then `SENTIARY_PROJECT_ID`. They do not read TOML, local supplements, or project keys.
- **Configured online pull** uses, in order: project `api_key` (literal or `{ env = "NAME" }`), `user_api_key` in adjacent `sentiary.local.toml`, `SENTIARY_PROJECT_API_KEY` only when exactly one project is declared, then `SENTIARY_USER_API_KEY`. A selected invalid key does not fall back. An explicit environment reference must resolve to a nonblank value. All declared projects need credentials, even when selecting one output.
- **Offline pull** needs no credentials and ignores the local supplement and credential environment values. Shared TOML structure still validates.
- **Explicit dotenv**: use `--env-file .env.local` after the command when requested. There is no automatic discovery. It supplies only credential variables, including names referenced by TOML. Process values win, even when explicitly empty. Only one file can be selected; missing or malformed selected files fail. It performs no shell execution or interpolation. Do not source the file. Offline pull, help, version, and update do not read it.
- `SENTIARY_BASE_URL` overrides the default API server, `https://api.sentiary.com`; it also scopes the export cache. TOML has no server URL setting. `SENTIARY_PROJECT_ID` does not override TOML IDs.

Keep `sentiary.local.toml`, dotenv credential files, and `.sentiary/` out of version control. `api_key = "${VARIABLE}"` is a literal string, not interpolation.

## Configure and Download Resources

Use `pull` for repeatable application-resource downloads; use standalone `export` for a one-off file. Inspect existing configuration before creating or changing it.

```sh
sentiary-cli pull --config=sentiary.toml --dry-run
sentiary-cli pull --config=sentiary.toml --output=android --output=compose --json
```

Without `--config`, discovery searches the working directory and its parents for the nearest `sentiary.toml`; configs are not merged. Relative output and list-file paths resolve from the config directory. Repeat `--output` for distinct output names, not remote project IDs. Omit it to download all outputs.

For a new setup, offer `sentiary-cli init --directory=.`. This interactive wizard configures existing projects; it does not discover or create remote projects. It requires a final `y`, refuses to overwrite configs, and aborts on `cancel` or EOF. Do not blindly pipe assumed answers. For automation, edit a reviewed TOML directly, preserving existing settings.

Minimal Android example (replace the project ID and use the project's actual language IDs):

```toml
config_version = 1

[projects.mobile]
id = "PROJECT_ID"
api_key = { env = "SENTIARY_MOBILE_API_KEY" }
languages = ["en", "cs"]

[outputs.android]
project = "mobile"
format = "android"
path = "build/sentiary/android/{locale}/strings.xml"
locale_map = { en = "values", cs = "values-cs" }
```

Configuration rules:

- `config_version = 1` is required. Outputs reference project aliases. Unknown fields fail.
- Project `languages`, `source_language`, and `include_empty` can be overridden per output; lists are replaced, not merged. `languages = "*"` selects project languages.
- Use exact API language IDs from `list-project-languages` or `info`; `cs`, `cs-CZ`, and `cs_CZ` are distinct. Do not infer or normalize IDs.
- `{language}` uses the target ID. `{locale}` requires an explicit output `locale_map`. `language_aliases` maps target IDs to whole-file source languages, not missing-key fallbacks.
- `pull` supports more formats than standalone import/export: Android/Compose, Apple, `xcstrings`, `stringsdict`, JSON variants, ARB, Gettext, spreadsheets, and others. Consult the configuration reference for exact tokens.
- Only `xcstrings` supports `layout = "combined"`: use a fixed path, select a `source_language`, and omit locale maps and aliases. Manual catalog edits are not merged.
- Output-only `keys_file` and `contexts` select a union of exact keys and case-sensitive contexts. `required_keys_file` validates selected translations; it does not expand selection. List files contain one literal key per line, no comments, and no duplicates. Empty `keys_file` does not mean the full catalog.
- `plural_bases_file` opts into Android/Compose plurals (`count.one`, `count.other`, etc.). Select/require the base, not quantity keys. Every emitted group needs usable `other`; structured plural import and Apple plural assembly are unsupported.

### Safe writes, cache, and offline use

Preview first when changing destinations or selection. `--dry-run` prints JSON without generated-file writes, but is not proof that ownership or filesystem checks will pass.

- Pull tracks owned paths/checksums in `.sentiary/<config-filename>.json`. Untracked destinations and manual edits block replacement or removal. Obsolete owned files can be removed after a successful pull.
- Use `--force` only after the user accepts adoption/discarding edits. Do not delete ownership manifests, locks, or caches to bypass a conflict. Force does not bypass path and symlink protections.
- All exports must succeed before writes, but the whole write is not transactional: an I/O failure may leave some files updated. Inspect failures before retrying.
- Prime a complete selection online, then use `pull --offline` with the same config/selection/cache. Offline restoration requires one matching complete cached set; it cannot combine sets or produce a new filtered selection from an old full export.
- Cache precedence is `--cache-dir`, `SENTIARY_CACHE_DIR`, then OS cache. Relative overrides use the working directory. Protect cached translation content; do not include credential files in CI caches.
- Do not automatically retry failed online pulls offline: this can hide revoked access or stale data. Offline freshness is not guaranteed.

For Gradle-managed resources, follow the project's settings plugin setup rather than introducing competing pull steps or overlapping destinations.

## Output and Automation

API commands emit JSON without `--json`. Export emits raw file content; import reads a file or stdin (`-`) and emits a JSON result. Pull emits a human summary by default, a protocol result with `--json`, or a preview array with `--dry-run` (which takes precedence over `--json`). Keep stderr separate from stdout.

Check exit status before parsing output: `0` is success, `2` is CLI usage failure, and `1` covers configuration, authentication, API, filesystem, cancellation, and other failures. A failed pull has no successful JSON result. Do not claim a dry run downloaded files.

`--integration-request` is for protocol-aware integrations and cannot combine with `--output`; ordinary automation should use named outputs. Consult the protocol reference before constructing requests or parsing versioned results.

## Installation and Updates

If the executable is missing, offer the official installation guide rather than silently installing software. Official installers support AMD64/ARM64 on macOS, Linux, and Windows and verify release checksums.

Use `sentiary-cli update --check` for an explicit availability check. Run `update` or `update --yes` only when the user approves installation. Self-replacement requires a stable official build and installer ownership record; source/package-manager or unmarked installs need their appropriate installation method. Do not request elevation or bypass ownership/checksum failures.

Automatic update checks/prompts are disabled in CI, noninteractive terminals, JSON/offline modes, and with `SENTIARY_NO_UPDATE=1`. Do not upgrade merely to troubleshoot without approval.

## Select a Command

Projects:

- Use `list-projects` to list projects for the current user.
- Use `get-project` to read one project and the current user's membership.
- Use `create-project` to create a project.
- Use `edit-project` to rename a project.
- Use `remove-project` to remove a project and its data.

Languages:

- Use `list-languages` to list all languages that Sentiary supports.
- Use `list-project-languages` to list the languages in a project.
- Use `add-project-language` to add a language to a project.
- Use `remove-project-language` to remove a language and its translations.

Members and invitations:

- Use `list-project-members` to list members and roles.
- Use `set-project-member-role` to set a member role.
- Use `remove-project-member` to remove a member.
- Use `list-invitations` to list invitations for the current user.
- Use `create-invitation` to invite a user to a project.
- Use `accept-invitation` or `decline-invitation` to respond to an invitation.
- Use `list-project-invitations` to list outstanding project invitations.
- Use `remove-project-invitation` to revoke a project invitation.

Strings and files:

- Use `list` to read project strings with pagination and optional translation filters.
- Use `search` to find text in keys, translations, descriptions, or context.
- Use `get --key` for an exact key lookup.
- Use `get --string-id` when the internal string ID is known.
- Use `add` to create one string and an optional first translation.
- Use `edit` to change a string name, description, or context.
- Use `remove` to remove one string and all its translations.
- Use `set-translation` to create or replace one translation.
- Use `remove-translation` to remove one translation.
- Use `info` to read the project name, languages, and modification date.
- Use `export` to get a localization file from Sentiary.
- Use `import` to send a localization file to Sentiary.

Use `administrator`, `developer`, or `translator` for a member or invitation role.

## Manage Access and Project Settings

List project members before you change a role or remove a member. List project invitations before you revoke one.

Only remove a project, project language, member, or invitation when the user explicitly requests that change. A project removal deletes project data. A project language removal also deletes translations in that language. A member role change can change project permissions.

When a user wants to leave a project, use `remove-project-member` with the user's ID. The API returns an empty member list after self-removal.

## Read Strings

Search with a general query:

```sh
sentiary-cli search \
  --project-id="project-id" \
  --query="Ahoj" \
  --include-language="cs,en"
```

The search can match translation content even when the string key does not match.

Use the returned `pagination.next` value as `--cursor` to read the next page. Continue until `pagination.next` is empty.

Use an exact key lookup before a targeted change:

```sh
sentiary-cli get \
  --project-id="project-id" \
  --key="welcome_message" \
  --include-language="cs,en"
```

## Change Strings

Resolve an exact key before an edit, removal, or translation change. These commands require the internal string ID.

Set a translation:

```sh
sentiary-cli set-translation \
  --project-id="project-id" \
  --string-id="string-id" \
  --language="cs" \
  --text="Ahoj"
```

Only remove a string or translation when the user explicitly requests removal. Report that string removal also removes every translation.

For `edit`, empty text leaves description/context unchanged; use `--clear-description` or `--clear-context` to clear them. `add --translation` needs `--language`; creating the term and its first translation uses separate requests. If the translation fails, check whether the term was created before retrying to avoid duplicates. `set-translation --text=""` sets empty text; it is not removal.

Do not use `--override-existing` during an import unless the user explicitly requests replacement.

## Synchronize Files

Use batch export and import for full localization files. Do not run one command for each key.

Export to a file:

```sh
sentiary-cli export \
  --project-id="project-id" \
  --format="json" \
  --language="cs" \
  --output="strings.json"
```

Import from a file:

```sh
sentiary-cli import \
  --project-id="project-id" \
  --format="json" \
  --language="cs" \
  --input="strings.json"
```

Standalone export/import formats are only `json`, `android`, `apple`, and `compose`; do not apply this restriction to configured pulls.

Before an import, inspect the input file and confirm its format and language. Preserve placeholders, escapes, newlines, and message syntax.

Use `-` as the input or output path for a standard stream. Use a file path when the user requests a file change.

## Report Results

Report the project, command, affected users, keys or file, and languages. Include role changes, invitation decisions, skipped imports, and removals when they occur.

For pulls, report selected outputs, written paths/counts, online versus offline status, and any ownership conflicts or partial failures. Never include credentials.

## Authoritative References

Use installed command help for available flags and these guides for validation and workflow details:

- [CLI reference](https://docs.sentiary.com/cli/reference/)
- [TOML configuration and credentials](https://docs.sentiary.com/configuration/reference/)
- [Download ownership, cache, and recovery](https://docs.sentiary.com/configuration/downloads/)
- [Android, Compose, and Xcode](https://docs.sentiary.com/cli/mobile/)
- [Installation and updates](https://docs.sentiary.com/getting-started/installation/)
- [Integration protocol](https://docs.sentiary.com/integration/protocol/)

In the Sentiary Tools checkout, the corresponding sources are under `docs/content/`; CLI registrations live in `go/internal/frontend/cli/`.
