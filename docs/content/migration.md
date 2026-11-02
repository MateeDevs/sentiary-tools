---
tags:
  - Migration
  - Gradle
---

# Migrate from Gradle plugin 1.x to 2.0

```sh
sentiary-cli pull --dry-run
```

Preview your converted configuration before replacing generated files. Gradle 2.0 uses a settings plugin and CLI-managed downloads. Follow the [setup and requirements](gradle/index.md).

| 1.x concept | New workflow |
| --- | --- |
| Project plugin in module/root `build.gradle.kts`, possibly convention plugins | One `com.sentiary.gradle` settings plugin in root `settings.gradle.kts` |
| `projectId` | `[projects.NAME].id` |
| `projectApiKey`, Gradle/local properties | User-key environment/local TOML, or project `api_key`; review changed precedence |
| `exportPaths` | Named `[outputs.NAME]` entries plus separate settings module/source-set/resource bindings |
| `Format.Android`, `ComposeResources`, `Json`, `Apple` | TOML `android`, `compose`, `json`, `apple` |
| `defaultLanguage` | Explicit `locale_map` for resource folder naming; combined `source_language` is a separate semantic choice |
| Folder/file naming callbacks | Explicit `path`, `{language}` / `{locale}`, and `locale_map`; arbitrary callbacks do not transfer |
| `disabledLanguages` | Explicit allowed `languages` list; wildcard has no exclusion syntax |
| `languageOverrides.fallbackTo` | `language_aliases` target-to-source map; chains allowed, whole-file copies only |
| `caching.enabled`, project-info timestamp cache | CLI conditional exports and completed sets; no equivalent disable-cache field |
| `requestTimeoutMillis` | No TOML/CLI timeout override; current executable uses a 30-second HTTP client |
| `--force-update` | No direct equivalent; online pull always checks state. `--force` instead discards local edits/adopts ownership |
| `sentiaryUpdateProjectInfo` / old update task | `:sentiaryPull`; resource staging adds its dependency automatically |

## Prepare shared configuration

Copy the [mobile example](examples/sentiary.toml) and replace IDs, languages, and paths. Convert every naming callback to an explicit destination rule. Use [advanced configuration](examples/advanced.toml) for aliases, wildcard selection, output overrides, and multiple projects. Combined catalogs require a selected source language and cannot use aliases or locale mappings.

Move personal credentials out of Gradle scripts. In CLI `v2.0.0`, online TOML
precedence is user environment key, local user key, then shared project key;
`SENTIARY_PROJECT_API_KEY` is ignored by pull. Version 2.1.0 adds project
environment sources. Legacy API commands still do not read local TOML.
Check which credential will win before migration.

## Project credentials in 2.1.0

CLI v2.1.0 supports `api_key = { env = "NAME" }`. Plugin 2.1.0 installs this CLI
by default; plugin 2.0.0 can select it with `-Psentiary.cliVersion=v2.1.0`.
No configuration conversion, cache reset, or database migration is required.
Existing literal `api_key` values remain supported. Review the priority below
when multiple credential sources are available.

**Behavior change:** configured pull now chooses project TOML `api_key` before
local `user_api_key`, then single-project `SENTIARY_PROJECT_API_KEY`, then
`SENTIARY_USER_API_KEY`. A user environment key no longer overrides project TOML
or the local user key. Lower-priority sources are not conflicts. Multi-project
configurations ignore the automatic project environment key. Blank literals and
fallback values are skipped. The wizard and legacy command rules do not change.

Prefer per-project `api_key = { env = "NAME" }` references for multiple projects.
Replace the literal `api_key` value rather than adding a second definition. Explicit
references override global user credentials and fail without fallback when missing
or blank. Empty literals, unlike explicit references, allow fallback.
See the [complete credential rules](configuration/reference.md#project-environment-credentials)
and [multi-project example](examples/project-credentials.toml).

Plain string values remain literal, including strings containing `${NAME}` or
reference-looking text.

Offline restore skips credential lookup and ignores local/value conflicts, but
still rejects invalid shared declarations. Key rotation and source changes do not
change export-cache identity. No new credential-bearing config file or Gradle
extension is required.

Configured pull also requires a backend with POST resource export for every format
and layout. A failed POST does not fall back to legacy GET. Capability negotiation
and export-state protocol versions are removed; generation, revision, exporter
version, and languages still protect complete sets. Cache format and Gradle–CLI
integration versions remain unchanged. Public GET exports and legacy commands
retain their existing behavior.

## Review generated-file ownership

Run `pull --dry-run` first. Existing files written by 1.x are untracked by the CLI even if unchanged. Back them up or review them, then adopt deliberately with `pull --force`. Use generated build directories where possible. Verify that no old task continues writing the same destinations.

Run an online pull to prime a completed set, then validate offline recovery after deleting only generated files. Preserve the cache and ownership state. A Gradle build-cache hit and the CLI export cache solve different problems.

## Switch Gradle wiring

Replace old module/convention-plugin application, old task dependencies, credential properties, and
`exportPaths` with the [settings DSL](gradle/index.md#setup). Register each Android production source-set
layer separately. Compose keeps conventional local resources plus explicit additional directories;
do not register another `customDirectory` for a bound source set. No legacy downloader remains.
