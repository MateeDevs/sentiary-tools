# Publishing the CLI and Gradle plugin

Use this guide to validate and publish CLI assets and the Gradle settings plugin.
See the [Gradle guide](../gradle/index.md) for consumer setup.

## Versions, Configuration Cache, and Isolated Projects

The source enforces **Gradle 9.1.0 or newer** and builds with **Java 21** and Kotlin 2.2.0. Source fixtures use AGP 9.0.0 (compile SDK 36), Kotlin 2.3.20, and Compose Multiplatform 1.10.3. These are the local consumer test matrix; no broader compatibility range is promised.

The implementation uses isolated lifecycle actions and task execution for CLI/config work. Test fixtures request Configuration Cache, Isolated Projects (`-Dorg.gradle.unsafe.isolated-projects=true`), and parallel execution.

For Isolated Projects with the matrix Compose version, set `org.jetbrains.compose.hot.reload.disable=true`. The fixtures also set `kotlin.internal.collectFUSMetrics=false` to avoid a Kotlin statistics build-service classloader conflict. Use the [canonical evaluation properties](../examples/gradle-evaluation.properties) as a starting point and replace its executable placeholder.

## Publication checklist

1. Run Go tests/vet, plugin integration tests/lint/validation, and strict documentation builds.
2. Deploy and verify the backend POST export contract, including validation and conditional responses.
3. Push the CLI release tag `vX.Y.Z` for the version being published. The release workflow publishes CLI/MCP assets, not the plugin.
4. Verify the published binaries, their version, and `SHA256SUMS`.
5. Publish the plugin separately using `./plugins/gradlew -p plugins :sentiary:publishPlugins`.
   Supply Gradle Plugin Portal credentials through the normal local/CI secret configuration.
6. Verify plugin marker and implementation coordinates and the pinned CLI version.
7. Verify production documentation describes the published behavior and versions.

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

The implementation coordinates are `com.sentiary:sentiary:2.1.0`.
The plugin marker is `com.sentiary.gradle:com.sentiary.gradle.gradle.plugin:2.1.0`
and references that implementation. From the repository root, generate both POMs
locally without publishing:

```sh
./plugins/gradlew -p plugins :sentiary:generatePomFileForPluginMavenPublication :sentiary:generatePomFileForSentiaryPluginMarkerMavenPublication
```
