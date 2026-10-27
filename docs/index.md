# Sentiary documentation

Manage translations in Sentiary, define generated files in `sentiary.toml`, and run `sentiary-cli pull` to download them. Keep credentials in your environment or a local supplement.

1. [Install the CLI](getting-started/installation.md).
2. [Create your configuration](getting-started/quickstart.md).
3. Choose an [example](examples/index.md) and review the [download and recovery rules](configuration/downloads.md).

## Implementation status

This site describes the CLI implementation in this source checkout. A source feature is not a promise that every published CLI or deployed backend supports it. Inspect `sentiary-cli version --json` and command help before using the new workflow. Online pulls require backend export/state protocol 1.

**Gradle plugin 2.0 is unreleased.** Its settings plugin, bindings, installer, and resource adapters are
implemented and locally verified in the adjacent repository. The [Gradle section](gradle/index.md)
records the actual DSL, tested versions, checks, and remaining release steps.

Use [CLI reference](cli/reference.md) for all commands and options, [configuration reference](configuration/reference.md) for TOML fields, and [integration protocol](integration/protocol.md) only when building an integration.

[Build and maintain this site](maintenance.md). Public hosting is outside this repository change.
