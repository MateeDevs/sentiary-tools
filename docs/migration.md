# Migrate from Gradle plugin 1.x to 2.0

**2.0 is unreleased.** The settings plugin is implemented and locally verified; publication and backend rollout remain separate. Keep the working 1.x integration until the compatible CLI and plugin are published, or validate from local source. See the [actual DSL and version matrix](gradle/index.md).

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

Copy the [mobile example](examples/sentiary.toml) and replace IDs, languages, and paths. Convert every naming callback to an explicit destination rule. Use [advanced examples](examples/advanced.toml) for aliases, wildcard selection, output overrides, and multiple projects. Combined catalogs require a selected source language and cannot use aliases or locale mappings.

Move personal credentials out of Gradle scripts. Online TOML precedence is user environment key, local user key, then shared project key. `SENTIARY_PROJECT_API_KEY` is not a direct pull fallback, and legacy API commands do not read local TOML. Check which credential will win before migration.

## Review generated-file ownership

Run `pull --dry-run` first. Existing files written by 1.x are untracked by the CLI even if unchanged. Back them up or review them, then adopt deliberately with `pull --force`. Use generated build directories where possible. Verify that no old task continues writing the same destinations.

Run an online pull to prime a completed set, then validate offline recovery after deleting only generated files. Preserve the cache and ownership state. A Gradle build-cache hit and the CLI export cache solve different problems.

## Switch Gradle wiring

Replace old module/convention-plugin application, old task dependencies, credential properties, and
`exportPaths` with the [settings DSL](gradle/index.md#setup). Register each Android production source-set
layer separately. Compose keeps conventional local resources plus explicit additional directories;
do not register another `customDirectory` for a bound source set. No legacy downloader remains.

The local matrix verifies multiple modules, staging collisions, Configuration Cache reuse, Isolated
Projects, and explicit offline restoration. Keep the unreleased warning until CLI `v2.0.0` and plugin
2.0 are published after the backend export-cache rollout.
