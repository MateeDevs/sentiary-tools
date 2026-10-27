# Examples

The downloadable TOML files are canonical: guides link to them instead of keeping divergent copies. Save the chosen file as `sentiary.toml` in your application's root. Replace placeholder project IDs and credentials before online use. Paths are relative to that file's directory.

| Walkthrough | Canonical configuration |
| --- | --- |
| [Android, KMP Compose, Xcode](mobile.md) | [Mobile TOML](sentiary.toml) |
| [Generic exports, aliases, multiple projects/modules](advanced.md) | [Advanced TOML](advanced.toml) |
| [Gradle 2.0 Android/Compose and multiple modules](gradle.md) | [Mobile settings](gradle-mobile.settings.gradle.kts), [multi-module settings](gradle-multi.settings.gradle.kts), [multi-module TOML](gradle-multi.toml) |
| [CI cache and offline recovery](recovery.md) | Uses either configuration above |

For local user-key storage, copy the [credential template](local-credentials.toml) as `sentiary.local.toml` and keep it ignored.

The mobile TOML is sourced from the existing `tools/examples/sentiary.toml` through a symlink in the documentation tree. There is one maintained copy; the static build includes its contents. Enable symlink checkout when building the repository on Windows, or use a Unix checkout for the documentation build.

[Gradle evaluation properties](gradle-evaluation.properties) provide the local CLI override and fixture compatibility settings.

Gradle 2.0 is unreleased. Its canonical settings examples follow the new implementation source; the separate manual resource setup works without that plugin.
