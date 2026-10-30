# Gradle release operations

Use this guide to validate and publish the Gradle settings plugin. See the [Gradle guide](../gradle/index.md) for consumer setup.

## Versions, Configuration Cache, and Isolated Projects

The source enforces **Gradle 9.1.0 or newer** and builds with **Java 21** and Kotlin 2.2.0. Source fixtures use AGP 9.0.0 (compile SDK 36), Kotlin 2.3.20, and Compose Multiplatform 1.10.3. These are the local consumer test matrix; no broader compatibility range is promised.

The implementation uses isolated lifecycle actions and task execution for CLI/config work. Test fixtures request Configuration Cache, Isolated Projects (`-Dorg.gradle.unsafe.isolated-projects=true`), and parallel execution.

For Isolated Projects with the matrix Compose version, set `org.jetbrains.compose.hot.reload.disable=true`. The fixtures also set `kotlin.internal.collectFUSMetrics=false` to avoid a Kotlin statistics build-service classloader conflict. Use the [canonical evaluation properties](../examples/gradle-evaluation.properties) as a starting point and replace its executable placeholder.

Version 2.0.0 was published after the backend export-cache rollout and migration through 6, authenticated export/state protocol 1 verification, and publication and verification of the CLI `v2.0.0` binaries and `SHA256SUMS`. For future releases, preserve this order: backend compatibility, CLI assets, asset verification, then plugin publication.

Current limitations include one binding per source set/kind/module, Android application/library plugins only, no Android test-source bindings, no generic/combined resource adapter, no public managed-checksum override, and framework API compatibility tied to the adapters above. The [integration protocol](../integration/protocol.md) remains a separate machine interface.

## Monorepo development

Follow the [development guide](development.md) to build the CLI and run the plugin integration checks.

To exercise the source in a consumer, include the plugin build in `settings.gradle.kts`:

```kotlin
pluginManagement {
    includeBuild("../sentiary-tools/plugins")
    repositories { gradlePluginPortal(); google(); mavenCentral() }
}
plugins { id("com.sentiary.gradle") }
```

The implementation coordinates are `com.sentiary:sentiary:2.0.0` (previous artifact:
`com.sentiary:gradle-plugin`). The plugin marker remains
`com.sentiary.gradle:com.sentiary.gradle.gradle.plugin:2.0.0` and references the new
implementation. From the repository root, generate both POMs locally without publishing:

```sh
./plugins/gradlew -p plugins :sentiary:generatePomFileForPluginMavenPublication :sentiary:generatePomFileForSentiaryPluginMarkerMavenPublication
```

## Import provenance

The plugin was copied without history from
[`MateeDevs/sentiary-gradle-plugin`](https://github.com/MateeDevs/sentiary-gradle-plugin),
branch `staging`, commit `7d4902adfc6a37519fe11f7315bd16a29d577c27`.
Only tracked files were imported. `gradle-plugin/` became `plugins/sentiary/`; the wrapper,
version catalog, and shared Gradle configuration live in `plugins/`.
The original plugin license is retained at `plugins/sentiary/LICENSE`.
Plugin ID, implementation class, packages, and version `2.0.0` are unchanged.
