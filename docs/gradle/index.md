# Gradle plugin 2.0 — unreleased

**Unreleased; verify against the completed implementation before release.** The adjacent plugin working tree now declares `2.0.0` and implements `SentiaryPlugin : Plugin<Settings>`. This reference follows that source. Local test reports were inspected for this documentation change; they record successful checks, while publication remains separate.

## Setup

Apply the plugin once in the root `settings.gradle.kts`, rather than module `build.gradle.kts`. For a local source checkout, use `pluginManagement { includeBuild("PATH_TO_PLUGIN_SOURCE") }` before the plugins block and apply `id("com.sentiary.gradle")`. Replace the placeholder with the sibling plugin checkout. See the [canonical mobile settings](../examples/gradle-mobile.settings.gradle.kts).

After publication, the release form will be `id("com.sentiary.gradle") version "2.0.0"` in the settings plugins block with `gradlePluginPortal()` available. That version is unreleased here; do not assume it can be resolved remotely today.

TOML owns project IDs, credentials, output formats/paths, language selection, aliases, and empty-value policy. The settings extension owns only resource consumption. There is no compatibility layer for the old `projectId`, `projectApiKey`, `exportPaths`, or naming callbacks.

## Binding containers and every field

The settings `sentiary` extension has two named containers: `android { create("NAME") { ... } }` and `compose { create("NAME") { ... } }`. Container kind supplies protocol format `android` or `compose`; the named output must use that format and `per_language` layout. A generic or combined Xcode output cannot be bound as Android/Compose resources.

| Field | Type | Required | Default | Behavior / accepted values |
| --- | --- | --- | --- | --- |
| Binding name (`create`) | string | Yes | None | Name within its container; identifies errors, not an output selection by itself |
| `modulePath` | string | Yes | Empty (invalid) | Absolute Gradle project path beginning with `:`, referring to an included module |
| `output` | string | Yes | Empty (invalid) | Nonblank named TOML output, validated by CLI |
| `sourceSet` | string | No | `main` | Matches `[A-Za-z][A-Za-z0-9_]*`; use `commonMain` explicitly for KMP Compose |
| `resourceRoot` | string | Yes | Empty (invalid) | Relative to build root, canonicalized; generated files must be strictly inside this root |
| `localDirectories` | list of strings | No | Empty list | Compose-only extra local resource roots, relative to build root; Android requires empty list |

There is at most one binding for each `(modulePath, kind, sourceSet)`. The same output may be selected by more than one binding if each root contains its files. Additional binding fields, format overrides, variant selectors, or language/credential overrides do not exist.

Android needs `com.android.application` or `com.android.library` in the bound module. Source sets beginning with `test` or `androidTest` (case-insensitive) are rejected. The adapter matches the selected source-set name against `main`, variant name, build type, combined flavor name, and individual flavor names. A source-set name's syntax alone does not prove that a variant will consume it.

Compose needs `org.jetbrains.compose`. Its adapter uses the public Compose `customDirectory` API and stages the conventional `src/<sourceSet>/composeResources` plus any `localDirectories`. Extra roots must have the same resource directory structure as ordinary Compose resources.

## Overrides and credentials

These are Gradle project properties, supplied with `-P` or `gradle.properties`. They are not extension fields. Standard Gradle property resolution applies; command-line `-P` overrides file values. No plugin-specific environment fallback is implemented for these properties.

| Property | Type | Required | Default | Behavior / example |
| --- | --- | --- | --- | --- |
| `sentiary.config` | string path | No | Empty | CLI discovery from build root; `-Psentiary.config=config/sentiary.toml` uses that path |
| `sentiary.cliVersion` | string release tag | No | `v2.0.0` | Managed version; `[A-Za-z0-9][A-Za-z0-9._-]*`; use an actual compatible release when available |
| `sentiary.cliExecutable` | string path | No | Empty | Overrides managed CLI installation; absolute path or relative to build root, e.g. `-Psentiary.cliExecutable=/path/to/sentiary-cli` |

An executable override takes precedence over `cliVersion` and bypasses managed-download checksum verification, but still undergoes protocol negotiation. There is no public checksum, installer URL, or CLI-install-cache directory override. `SENTIARY_USER_API_KEY`, local credentials, `SENTIARY_CACHE_DIR`, and `SENTIARY_BASE_URL` are inherited by the CLI process. The export cache is separate from the managed binary cache. TOML destinations remain relative to the config directory even when `sentiary.config` points below the build root; binding roots remain relative to the build root.

## Tasks and automatic build dependencies

| Task | Behavior |
| --- | --- |
| Root `:sentiaryPull` | Runs version handshake, invokes `pull --json`, validates the envelope, saves `.gradle/sentiary/result.json` |
| `sentiaryStageAndroid<SourceSet>` in bound module | Depends on root pull; validates checksums/root containment; stages generated and source-set local resources |
| `sentiaryStageCompose<SourceSet>` in bound module | Depends on root pull; stages generated, conventional, and extra Compose roots |

Names capitalize the first character of the source set: `main` becomes `sentiaryStageAndroidMain`, `commonMain` becomes `sentiaryStageComposeCommonMain`. Staging outputs go to the module's `build/sentiary/staging/<kind>/<sourceSet>`.

When bindings exist, the shared CLI request selects only their outputs; unbound outputs are excluded. With no bindings, `sentiaryPull` runs a normal full-config pull, useful for generic exports or Xcode. A build runs the shared task at most once in its task graph. The task is deliberately not cached or skipped for online freshness, even when Configuration Cache is reused. Staging also runs again to validate current files/collisions. Old `sentiaryUpdateLocalizations`, `sentiaryUpdateProjectInfo`, and `--force-update` are removed; no plugin `--force` option forwards ownership adoption. Perform deliberate adoption with the CLI.

Android source-set resource directories are collected into staging, then replaced by the staged directory. Applicable variant pre-build lifecycle tasks depend on staging. Compose registers a task-backed custom directory. Do not add separate manual directory wiring for the same binding. Keep framework source-set configuration and the plugin's generated roots consistent.

## Collisions and staging guarantees

Within a binding, distinct physical files with the same relative path, compared without case, fail even when content is identical. A repeated reference to the same canonical physical file is deduplicated. Rename local resources such as `values/strings.xml` when they collide with a generated file at that relative path; the plugin does not merge XML or implement key-level fallback. Android layering across distinct source sets remains the framework's responsibility.

Generated checksums and resource-root containment are checked before replacement. Missing generated resources and non-directory local inputs fail. Inputs and collisions are validated before replacing the old staged tree. Final directory copying is not a filesystem transaction; an I/O error can still leave incomplete staging. Re-run after fixing the failure.

## Managed CLI and offline builds

The managed installer supports macOS, Linux, and Windows on AMD64/ARM64. It downloads raw `sentiary-cli-<os>-<arch>[.exe]` release assets and `SHA256SUMS` from `MateeDevs/sentiary-tools`. It requires exactly one matching checksum line, verifies SHA-256, sets executable permissions on Unix, and atomically replaces binary and manifest files. Downloads use 30-second connection/read timeouts. A per-version/platform file lock coordinates installation.

Managed binaries live in `GRADLE_USER_HOME/caches/sentiary/cli/<version>/<platform>/`. A cached binary is checked against its saved checksum manifest before reuse. With Gradle `--offline`, a missing/corrupt managed binary fails instead of downloading; use an intact primed binary or `sentiary.cliExecutable`.

Gradle `--offline` forwards CLI `--offline`. It also needs a complete matching export-cache set. After deleting generated files, the pull task restores them, then staging runs again. Prime both caches online first. No automatic online-error-to-offline fallback exists. See [CI and offline recovery](../examples/recovery.md).

## Versions, Configuration Cache, and Isolated Projects

The source enforces **Gradle 9.1.0 or newer** and builds with **Java 21** and Kotlin 2.2.0. Source fixtures use AGP 9.0.0 (compile SDK 36), Kotlin 2.3.20, and Compose Multiplatform 1.10.3. These are the local consumer test matrix; no broader compatibility range is promised.

The implementation uses isolated lifecycle actions and task execution for CLI/config work. Test fixtures request Configuration Cache, Isolated Projects (`-Dorg.gradle.unsafe.isolated-projects=true`), and parallel execution. The existing local XML reports dated October 27, 2026 record 14 passing tests (7 distribution tests and 7 plugin tests), with no failures, errors, or skips. They include Android app/library resource tasks, two KMP Compose modules, Configuration Cache reuse, parallel Isolated Projects, collision rejection, incompatible protocol rejection, and an actual-CLI local-server fixture that checks online freshness and offline restoration. This documentation run inspected those reports; it did not rerun the plugin suite.

For Isolated Projects with the matrix Compose version, set `org.jetbrains.compose.hot.reload.disable=true`. The fixtures also set `kotlin.internal.collectFUSMetrics=false` to avoid a Kotlin statistics build-service classloader conflict. Use the [canonical evaluation properties](../examples/gradle-evaluation.properties) as a starting point and replace its executable placeholder.

Before release, complete the backend export-cache rollout/migration through 6, verify authenticated export/state protocol 1, publish compatible CLI `v2.0.0` binaries and `SHA256SUMS`, verify those assets, then publish the plugin. Keep the unreleased label until release verification is complete.

Current limitations include one binding per source set/kind/module, Android application/library plugins only, no Android test-source bindings, no generic/combined resource adapter, no public managed-checksum override, and framework API compatibility tied to the adapters above. The [integration protocol](../integration/protocol.md) remains a separate machine interface.
