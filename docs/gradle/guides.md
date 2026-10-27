---
tags:
  - Gradle
  - Android
  - Compose
---

# Android and Compose builds

```sh
./gradlew sentiaryPull
./gradlew :android:app:mergeDebugResources :shared:prepareComposeResourcesTaskForCommonMain
```

Configure your existing Android/KMP build as follows. See the [reference](index.md) for every binding and override.

## Android and KMP Compose

1. Copy [mobile TOML](../examples/sentiary.toml) as `sentiary.toml`. Replace its project ID/key and verify language IDs.
2. Merge the [canonical mobile settings](../examples/gradle-mobile.settings.gradle.kts) into your root `settings.gradle.kts`. Keep the required settings block order, repositories, and your existing modules. The example assumes modules `:android:app` and `:shared`.
3. Ensure the app module applies `com.android.application` (or `com.android.library`) and the shared module applies `org.jetbrains.compose` with its existing KMP/Compose plugins. Use the [supported versions](index.md#requirements).
4. Provide credentials through the [CLI environment or local TOML](../configuration/reference.md#local-supplement-and-credentials). The plugin manages the CLI installation.
5. Run a consuming resource task or `sentiaryPull` as shown below.

```sh
./gradlew sentiaryPull
./gradlew :android:app:mergeDebugResources :shared:prepareComposeResourcesTaskForCommonMain
```

The shared task selects only the bound `android` and `compose` outputs. It does not download the unbound `ios` catalog. Staging uses the generated root from each binding, preserves local resources at distinct relative paths, and wires task dependencies automatically. Do not also apply the manual resource-directory wiring from the separate mobile guide.

The Android adapter collects the source set's configured resource directories. The Compose adapter includes `shared/src/commonMain/composeResources` automatically. To include extra local Compose roots, set `localDirectories = listOf("shared/extraComposeResources")` inside that binding. Those paths are relative to the build root and must have a standard Compose resource tree. A duplicate relative filename across generated/local roots fails instead of overriding. See [collision rules](index.md#collisions-and-staging-guarantees).

## Multiple modules

Copy [multi-module TOML](../examples/gradle-multi.toml) as `sentiary.toml`, and merge [multi-module settings](../examples/gradle-multi.settings.gradle.kts) into your existing settings. Replace project IDs. The app module applies `com.android.application`; the library applies `com.android.library`. Both need ordinary Android namespace/SDK configuration.

The example binds `app_main`, `app_debug`, and `library_main` to `:app` main/debug and `:library` main. Main and debug use separate generated roots. Android's ordinary source-set priority applies across layers; files colliding inside one layer fail during staging. All three bindings belong to one shared request, even when only one module resource task is requested.

```sh
./gradlew :app:mergeDebugResources :library:packageDebugResources --parallel
```

Use a user key with access to both remote projects, or project keys with no higher-priority user key configured. Add more bindings for additional modules/source sets, keeping at most one per kind/module/source set. The same mechanism supports multiple Compose modules by creating one `compose` binding per module and matching each output/root; use `commonMain` explicitly.

## Offline clean recovery and configuration reuse

Run the consuming tasks online first to prime a complete export set (and the binary cache if using managed installation). Delete only generated directories and module staging outputs. Keep the export cache and an intact CLI binary. Then use the same consuming tasks with Gradle `--offline`; it forwards the offline flag, restores generated files, and stages resources again.

```sh
./gradlew sentiaryPull --offline
```

Run consuming tasks twice with `--configuration-cache` and `-Dorg.gradle.unsafe.isolated-projects=true`. Check for configuration reuse and one pull per task graph. Online pulls still run again to check freshness; configuration reuse does not skip them.

## Generic exports and Xcode

For a build that only needs generic/Xcode files, apply the settings plugin with no resource bindings. `sentiaryPull` then downloads all configured outputs without Android/Compose adapters. A mixed build with resource bindings downloads only bound outputs; use a separate explicit CLI selection for an unbound catalog:

```sh
sentiary-cli pull --output ios
```

Run that before the Xcode build and register the catalog in its target. It is not a Gradle resource staging binding. See [CI caching](../configuration/downloads.md#ci-caching) for export-cache storage and [migration](../migration.md) for old-task and ownership differences.
