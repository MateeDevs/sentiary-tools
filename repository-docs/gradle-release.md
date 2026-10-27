# Gradle release operations

The public [Gradle guide](../docs/gradle/index.md) describes 2.0 as the current version. Keep publication steps and historical validation evidence here. The following notes were retained from the pre-publication documentation.

## Versions, Configuration Cache, and Isolated Projects

The source enforces **Gradle 9.1.0 or newer** and builds with **Java 21** and Kotlin 2.2.0. Source fixtures use AGP 9.0.0 (compile SDK 36), Kotlin 2.3.20, and Compose Multiplatform 1.10.3. These are the local consumer test matrix; no broader compatibility range is promised.

The implementation uses isolated lifecycle actions and task execution for CLI/config work. Test fixtures request Configuration Cache, Isolated Projects (`-Dorg.gradle.unsafe.isolated-projects=true`), and parallel execution. The existing local XML reports dated October 27, 2026 record 14 passing tests (7 distribution tests and 7 plugin tests), with no failures, errors, or skips. They include Android app/library resource tasks, two KMP Compose modules, Configuration Cache reuse, parallel Isolated Projects, collision rejection, incompatible protocol rejection, and an actual-CLI local-server fixture that checks online freshness and offline restoration. This documentation run inspected those reports; it did not rerun the plugin suite.

For Isolated Projects with the matrix Compose version, set `org.jetbrains.compose.hot.reload.disable=true`. The fixtures also set `kotlin.internal.collectFUSMetrics=false` to avoid a Kotlin statistics build-service classloader conflict. Use the [canonical evaluation properties](../docs/examples/gradle-evaluation.properties) as a starting point and replace its executable placeholder.

Before release, complete the backend export-cache rollout/migration through 6, verify authenticated export/state protocol 1, publish compatible CLI `v2.0.0` binaries and `SHA256SUMS`, verify those assets, then publish the plugin. Keep the unreleased label until release verification is complete.

Current limitations include one binding per source set/kind/module, Android application/library plugins only, no Android test-source bindings, no generic/combined resource adapter, no public managed-checksum override, and framework API compatibility tied to the adapters above. The [integration protocol](integration-protocol.md) remains a separate machine interface.
