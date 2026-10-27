# Gradle release operations

The public [Gradle guide](../docs/gradle/index.md) describes 2.0 as the current version. Keep publication steps and historical validation evidence here. The following notes include evidence retained from the pre-publication documentation.

## Versions, Configuration Cache, and Isolated Projects

The source enforces **Gradle 9.1.0 or newer** and builds with **Java 21** and Kotlin 2.2.0. Source fixtures use AGP 9.0.0 (compile SDK 36), Kotlin 2.3.20, and Compose Multiplatform 1.10.3. These are the local consumer test matrix; no broader compatibility range is promised.

The implementation uses isolated lifecycle actions and task execution for CLI/config work. Test fixtures request Configuration Cache, Isolated Projects (`-Dorg.gradle.unsafe.isolated-projects=true`), and parallel execution. The existing local XML reports dated October 27, 2026 record 14 passing tests (7 distribution tests and 7 plugin tests), with no failures, errors, or skips. They include Android app/library resource tasks, two KMP Compose modules, Configuration Cache reuse, parallel Isolated Projects, collision rejection, incompatible protocol rejection, and an actual-CLI local-server fixture that checks online freshness and offline restoration. This documentation run inspected those reports; it did not rerun the plugin suite.

For Isolated Projects with the matrix Compose version, set `org.jetbrains.compose.hot.reload.disable=true`. The fixtures also set `kotlin.internal.collectFUSMetrics=false` to avoid a Kotlin statistics build-service classloader conflict. Use the [canonical evaluation properties](../docs/examples/gradle-evaluation.properties) as a starting point and replace its executable placeholder.

Before release, complete the backend export-cache rollout/migration through 6, verify authenticated export/state protocol 1, publish compatible CLI `v2.0.0` binaries and `SHA256SUMS`, verify those assets, then publish the plugin. Keep the unreleased label until release verification is complete.

Current limitations include one binding per source set/kind/module, Android application/library plugins only, no Android test-source bindings, no generic/combined resource adapter, no public managed-checksum override, and framework API compatibility tied to the adapters above. The [integration protocol](integration-protocol.md) remains a separate machine interface.

## Monorepo development

Run from the `sentiary-tools` root (or `cd tools` in the application checkout):

```sh
go build -o bin/sentiary-cli ./cmd/sentiary-cli
SENTIARY_TEST_CLI="$PWD/bin/sentiary-cli" ./gradlew :sentiary:test :sentiary:ktlintCheck :sentiary:validatePlugins
```

Use JDK 21, Python 3, and an Android SDK with platform 36. `SENTIARY_TEST_CLI` enables
an actual-CLI local-server fixture; omitting it skips that fixture. Android fixtures
use `ANDROID_HOME`, `ANDROID_SDK_ROOT`, or the standard macOS SDK location.
The staging plugin workflow builds the CLI from the same checkout, then runs these
checks and assembles the plugin. Artifacts and reports are in `sentiary/build/`.

To exercise the source in a consumer, include the monorepo root in `settings.gradle.kts`:

```kotlin
pluginManagement {
    includeBuild("../sentiary-tools")
    repositories { gradlePluginPortal(); google(); mavenCentral() }
}
plugins { id("com.sentiary.gradle") }
```

The implementation coordinates are `com.sentiary:sentiary:2.0.0` (previous artifact:
`com.sentiary:gradle-plugin`). The plugin marker remains
`com.sentiary.gradle:com.sentiary.gradle.gradle.plugin:2.0.0` and references the new
implementation. Generate both POMs locally without publishing:

```sh
./gradlew :sentiary:generatePomFileForPluginMavenPublication :sentiary:generatePomFileForSentiaryPluginMarkerMavenPublication
```

## Import provenance

The plugin was copied without history from
[`MateeDevs/sentiary-gradle-plugin`](https://github.com/MateeDevs/sentiary-gradle-plugin),
branch `staging`, commit `7d4902adfc6a37519fe11f7315bd16a29d577c27`.
Only tracked files were imported. `gradle-plugin/` became `sentiary/`; the wrapper,
version catalog, and shared Gradle configuration live at the repository root.
The original plugin license is retained at `sentiary/LICENSE`.
Plugin ID, implementation class, packages, and version `2.0.0` are unchanged.
The original `main` was fast-forwarded to that exact commit, including the move notice.
Archiving the original repository remains an owner action.

## Migration validation

On October 27, 2026, the migrated module passed 16 tests (7 distribution and 9 plugin
integration tests), with no failures, errors, or skips. The actual CLI was built from
the same working tree and supplied through `SENTIARY_TEST_CLI`. This run covered
Android app/library resources, Compose KMP, Configuration Cache reuse, Isolated
Projects, collision failures, online freshness, and offline restoration against a
local fixture server. Kotlin lint, plugin validation, and assembly also passed.

Generated implementation and marker POMs were checked for the coordinates above,
including the marker's implementation dependency. The plugin descriptor still
names `com.sentiary.SentiaryPlugin`. A separate temporary consumer applied the
plugin using `includeBuild` of the monorepo root and exposed `sentiaryPull` while
successfully storing its Configuration Cache. Go tests, `go vet ./...`, and the
strict Zensical build passed. Imported source, tests, wrapper, and shared build
configuration were compared byte-for-byte with the recorded source commit;
only the documented module/settings/URL changes and CI adaptation were made.
No public release was created.

## Plugin Portal URL handoff

On October 27, 2026, the public page still linked to the original repository.
The local publishing configuration now uses the monorepo root URLs, but this does
not update the already published page. No browser or authenticated owner management
session was available during migration. No version was published to change URLs.

The owner can edit the existing plugin page, then verify both public links at
<https://plugins.gradle.org/plugin/com.sentiary.gradle>. If direct editing is unavailable,
submit the following request to [Plugin Portal support](https://github.com/gradle/plugin-portal-requests/issues):

> Subject: Update repository URLs for com.sentiary.gradle
>
> Please update the existing public metadata for `com.sentiary.gradle`
> (<https://plugins.gradle.org/plugin/com.sentiary.gradle>, owner Vojtěch Hořánek).
> The plugin moved from `MateeDevs/sentiary-gradle-plugin` into `MateeDevs/sentiary-tools`.
> Set Website to `https://github.com/MateeDevs/sentiary-tools` and Sources/VCS URL to
> `https://github.com/MateeDevs/sentiary-tools.git`, without a branch or subdirectory.
> The original repository README records the move. Keep the plugin ID, ownership,
> and published versions unchanged. This request does not publish a new release.
