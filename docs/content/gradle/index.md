---
tags:
  - Gradle
---

# Gradle plugin

## Setup

Apply the plugin once in root `settings.gradle.kts`:

```kotlin
pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}
plugins {
    id("com.sentiary.gradle") version "2.1.0"
}
```

Create `sentiary.toml` with the [first download guide](../getting-started/quickstart.md), then follow the [Android and Compose guide](guides.md). Download the [mobile settings](../examples/gradle-mobile.settings.gradle.kts) to configure bindings.

[Plugin source on GitHub](https://github.com/MateeDevs/sentiary-tools/tree/main/plugins/sentiary)

Configure project IDs, credentials, output formats and paths, languages, aliases,
and empty-value policy in TOML. Use the settings extension to bind those outputs
to modules and source sets.

Term selection, required translations, and opt-in Android/Compose plurals also belong
in [TOML output options](../configuration/reference.md#selection-required-translations-and-plurals).
Plugin 2.1.0 installs CLI v2.1.0, which supports these fields. The backend must
support POST resource exports.
Default resource folders require explicit `locale_map` entries.

## Resource bindings

The settings `sentiary` extension has two named containers: `android { create("NAME") { ... } }` and `compose { create("NAME") { ... } }`. The named TOML output must use the matching `android` or `compose` format and `per_language` layout. For other outputs, run `sentiaryPull` without a resource binding.

| Field | Type | Required | Default | Behavior / accepted values | Example |
| --- | --- | --- | --- | --- | --- |
| Binding name (`create`) | string | Yes | None | Name within its container; identifies errors, not an output selection by itself | `create("app")` |
| `modulePath` | string | Yes | Empty (invalid) | Absolute Gradle project path beginning with `:`, referring to an included module | `modulePath = ":app"` |
| `output` | string | Yes | Empty (invalid) | Nonblank named TOML output, validated by CLI | `output = "android"` |
| `sourceSet` | string | No | `main` | Matches `[A-Za-z][A-Za-z0-9_]*`; use `commonMain` explicitly for KMP Compose | `sourceSet = "commonMain"` |
| `resourceRoot` | string | Yes | Empty (invalid) | Relative to build root, canonicalized; generated files must be strictly inside this root | `resourceRoot = "app/build/generated/sentiary"` |
| `localDirectories` | list of strings | No | Empty list | Compose-only extra local resource roots, relative to build root; Android requires empty list | `localDirectories = listOf("shared/extraComposeResources")` |

There is at most one binding for each `(modulePath, kind, sourceSet)`. The same output may be selected by more than one binding if each root contains its files.

Android needs `com.android.application` or `com.android.library` in the bound module. Source sets beginning with `test` or `androidTest` (case-insensitive) are rejected. The adapter matches the selected source-set name against `main`, variant name, build type, combined flavor name, and individual flavor names. A source-set name's syntax alone does not prove that a variant will consume it.

Compose needs `org.jetbrains.compose`. Its adapter uses the public Compose `customDirectory` API and stages the conventional `src/<sourceSet>/composeResources` plus any `localDirectories`. Extra roots must have the same resource directory structure as ordinary Compose resources.

## Overrides and credentials

Supply these Gradle project properties with `-P` or `gradle.properties`, not as extension fields. Standard Gradle property resolution applies; command-line `-P` overrides file values.

| Property | Type | Required | Default | Behavior / example |
| --- | --- | --- | --- | --- |
| `sentiary.config` | string path | No | Empty | CLI discovery from build root; `-Psentiary.config=config/sentiary.toml` uses that path |
| `sentiary.cliVersion` | string release tag | No | `v2.1.0` | Managed version; `[A-Za-z0-9][A-Za-z0-9._-]*`; e.g. `-Psentiary.cliVersion=v2.1.0` |
| `sentiary.cliExecutable` | string path | No | Empty | Overrides managed CLI installation; absolute path or relative to build root, e.g. `-Psentiary.cliExecutable=/path/to/sentiary-cli` |

An executable override takes precedence over `cliVersion`. You are responsible for
verifying that executable; the plugin checks its integration compatibility before use.
The CLI inherits the task environment, including credential, cache, and server variables.
TOML destinations are relative to the config directory; binding roots are relative
to the build root.

## Project credentials

Configure credentials in TOML with `api_key = { env = "VARIABLE_NAME" }` and make
the variable available to Gradle. The CLI resolves credentials when the task runs,
including after Configuration Cache reuse; keys are not copied into task properties.
See the [credential reference](../configuration/reference.md#project-environment-credentials)
for alternatives and priority rules.

## Tasks and automatic build dependencies

| Task | Behavior |
| --- | --- |
| Root `:sentiaryPull` | Downloads the selected outputs using the CLI |
| `sentiaryStageAndroid<SourceSet>` in bound module | Depends on root pull; validates checksums/root containment; stages generated and source-set local resources |
| `sentiaryStageCompose<SourceSet>` in bound module | Depends on root pull; stages generated, conventional, and extra Compose roots |

Names capitalize the first character of the source set: `main` becomes `sentiaryStageAndroidMain`, `commonMain` becomes `sentiaryStageComposeCommonMain`. Staging outputs go to the module's `build/sentiary/staging/<kind>/<sourceSet>`.

When bindings exist, `sentiaryPull` downloads only their outputs; unbound outputs
are excluded. With no bindings, it downloads all configured outputs, including
generic formats or Xcode catalogs. It runs at most once per task graph and checks
for updates on every online build, including with Configuration Cache reuse.
Staging also validates current files and collisions on each build.
To adopt existing destinations or discard local edits, review the files and run
the CLI with `pull --force`; Gradle does not forward that option.

Android source-set resource directories are collected into staging, then replaced by the staged directory. Applicable variant pre-build lifecycle tasks depend on staging. Compose registers a task-backed custom directory. Do not add separate manual directory wiring for the same binding. Keep framework source-set configuration and the plugin's generated roots consistent.

## Collisions and staging guarantees

Within a binding, distinct physical files with the same relative path, compared without case, fail even when content is identical. A repeated reference to the same canonical physical file is deduplicated. Rename local resources such as `values/strings.xml` when they collide with a generated file at that relative path; the plugin does not merge XML or implement key-level fallback. Android layering across distinct source sets remains the framework's responsibility.

Generated checksums and resource-root containment are checked before replacement. Missing generated resources and non-directory local inputs fail. Inputs and collisions are validated before replacing the old staged tree. Final directory copying is not a filesystem transaction; an I/O error can still leave incomplete staging. Re-run after fixing the failure.

## Managed CLI and offline builds

The plugin installs the CLI for macOS, Linux, and Windows on AMD64/ARM64. It verifies the release's `SHA256SUMS` before installation and checks cached binaries before reuse.

Managed binaries live in `GRADLE_USER_HOME/caches/sentiary/cli/<version>/<platform>/`. A cached binary is checked against its saved checksum manifest before reuse. With Gradle `--offline`, a missing/corrupt managed binary fails instead of downloading; use an intact primed binary or `sentiary.cliExecutable`.

Gradle `--offline` forwards CLI `--offline`. It also needs a complete matching export-cache set. After deleting generated files, the pull task restores them, then staging runs again. Prime both caches online first. No automatic online-error-to-offline fallback exists. See [CI and offline recovery](../configuration/downloads.md#ci-caching).

## Requirements

| Component | Version |
| --- | --- |
| Gradle | 9.1.0 or newer |
| Java | 21 |
| Android Gradle Plugin | 9.0.0 (compile SDK 36) |
| Kotlin for consuming projects | 2.3.20 |
| Compose Multiplatform | 1.10.3 |
| Sentiary CLI | 2.1.0, integration protocol 1 |

Use this supported framework combination; other framework versions are not covered by this guide. Configuration Cache and Isolated Projects are supported. Online pulls still run to check freshness when Gradle reuses configuration.

For Compose with Isolated Projects, see the required
[framework settings](../troubleshooting.md#compose-and-isolated-projects).

## Limitations

Android bindings support application/library plugins and production source sets
only. Resource staging supports Android and Compose outputs, not generic formats
or combined catalogs. See [offline use](../configuration/downloads.md) for recovery requirements.
