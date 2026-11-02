# Gradle release operations

Use this guide to validate and publish the Gradle settings plugin. See the [Gradle guide](../gradle/index.md) for consumer setup.

## Versions, Configuration Cache, and Isolated Projects

The source enforces **Gradle 9.1.0 or newer** and builds with **Java 21** and Kotlin 2.2.0. Source fixtures use AGP 9.0.0 (compile SDK 36), Kotlin 2.3.20, and Compose Multiplatform 1.10.3. These are the local consumer test matrix; no broader compatibility range is promised.

The implementation uses isolated lifecycle actions and task execution for CLI/config work. Test fixtures request Configuration Cache, Isolated Projects (`-Dorg.gradle.unsafe.isolated-projects=true`), and parallel execution.

For Isolated Projects with the matrix Compose version, set `org.jetbrains.compose.hot.reload.disable=true`. The fixtures also set `kotlin.internal.collectFUSMetrics=false` to avoid a Kotlin statistics build-service classloader conflict. Use the [canonical evaluation properties](../examples/gradle-evaluation.properties) as a starting point and replace its executable placeholder.

Release in this order: backend compatibility, CLI assets, asset verification, then
plugin publication. Version 2.1.0 requires no new database or client migration.
Configured pull requires POST resource exports; verify the backend before publishing.

Current limitations include one binding per source set/kind/module, Android application/library plugins only, no Android test-source bindings, no generic/combined resource adapter, no public managed-checksum override, and framework API compatibility tied to the adapters above. The [integration protocol](../integration/protocol.md) remains a separate machine interface.

## CLI and plugin 2.1.0

- Extend `projects.NAME.api_key` to accept a literal string or `{ env = "NAME" }`
  for explicit, independently resolved project environment credentials. Strings
  never interpolate; empty/missing referenced values fail without user-key fallback.
- Configured pull now honors nonblank `SENTIARY_PROJECT_API_KEY` for exactly one
  declared project, after project TOML and local user credentials. Multi-project
  configs ignore it. Project TOML now wins over user environment/local credentials;
  local user credentials win over user environment credentials. Lower-priority
  sources are not conflicts.
- All configured pulls require POST resource export. Capability negotiation and
  export-state protocol versions are removed; consistency metadata remains.
- Offline pull does not read credential variables or the local supplement;
  shared source declarations still validate. Credentials/source names stay outside
  export-cache identity and plugin task properties.
- `init` and legacy API command credential selection are unchanged.

- Add output-level exact-key/context selection, required translations, and opt-in
  Android/Compose dotted-key plural assembly. CLI v2.1.0 is the first supporting release.
- Plugin 2.1.0 pins managed CLI v2.1.0. The binding DSL and integration protocol remain unchanged.
- Structured plural import, Apple plural assembly, English-source coverage validation,
  and Android resource-attribute preservation remain outside this release.

### Backend resource-export compatibility correction

The backend correction uses one XMLUtil writer and the same legacy text
transformations for Android/Compose strings and plural items. It supersedes the
original plural plan's stricter placeholder policy: mixed numbered/unnumbered and
relative arguments are no longer rejected. Compose still numbers only unnumbered
`%s` and `%d`; this is not general width/precision normalization, and mixed numbering
can collide. See the [formatting limits](../configuration/reference.md#selection-required-translations-and-plurals).
Plural structural validation and ordinary mixed-case/dotted keys remain unchanged.

No new CLI/plugin version, config field, cache deletion, or `--force` is required.
CLI/plugin 2.1.0 remain compatible. Deployment can change export bytes and ETags;
offline pulls still restore the exact cached bytes rather than regenerating XML.
This note describes the backend correction, not a new tools release or a deployment
confirmation.

### Publication checklist

1. Run Go tests/vet, plugin integration tests/lint/validation, and strict documentation builds.
2. Deploy and verify the backend POST export contract, including validation and conditional responses.
3. Push the CLI release tag `v2.1.0`. The release workflow publishes CLI/MCP assets, not the plugin.
4. Verify the published binaries, their version, and `SHA256SUMS`.
5. Publish the plugin separately using `./plugins/gradlew -p plugins :sentiary:publishPlugins`.
   Supply Gradle Plugin Portal credentials through the normal local/CI secret configuration.
6. Verify plugin marker and implementation coordinates and the pinned CLI version.
7. Confirm production documentation and publish the maintainer-facing report.

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

The implementation coordinates are `com.sentiary:sentiary:2.1.0` (previous artifact:
`com.sentiary:gradle-plugin`). The plugin marker remains
`com.sentiary.gradle:com.sentiary.gradle.gradle.plugin:2.1.0` and references the new
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
The imported plugin retained its ID, implementation class, packages, and version
`2.0.0`. The current release is 2.1.0.
