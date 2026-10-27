<p align="center">
  <img src="docs/content/assets/logo.svg" width="144" height="144" alt="Sentiary logo">
</p>

<h1 align="center">Sentiary Tools</h1>

<p align="center">
  Localization tools for your terminal, builds, and AI assistants.
</p>

<p align="center">
  <a href="https://github.com/MateeDevs/sentiary-tools/actions/workflows/staging.yml"><img src="https://github.com/MateeDevs/sentiary-tools/actions/workflows/staging.yml/badge.svg?branch=staging" alt="CLI checks"></a>
  <a href="https://github.com/MateeDevs/sentiary-tools/actions/workflows/staging-plugin.yml"><img src="https://github.com/MateeDevs/sentiary-tools/actions/workflows/staging-plugin.yml/badge.svg?branch=staging" alt="Gradle checks"></a>
  <a href="https://github.com/MateeDevs/sentiary-tools/actions/workflows/docs.yml"><img src="https://github.com/MateeDevs/sentiary-tools/actions/workflows/docs.yml/badge.svg?branch=staging" alt="Documentation checks"></a>
  <a href="https://github.com/MateeDevs/sentiary-tools/releases"><img src="https://img.shields.io/github/v/release/MateeDevs/sentiary-tools?color=2196F3" alt="Latest release"></a>
</p>

<p align="center">
  <a href="https://docs.sentiary.com/">Documentation</a> ·
  <a href="https://docs.sentiary.com/getting-started/installation/">Install</a> ·
  <a href="https://docs.sentiary.com/getting-started/quickstart/">Quickstart</a> ·
  <a href="https://docs.sentiary.com/contributing/development/">Contribute</a>
</p>

## From translations to your app

- **One configuration.** Download Android, Compose Multiplatform, and Xcode resources with a TOML config.
- **Part of your build.** Generate and stage translations with the Gradle settings plugin.
- **Ready offline.** Restore a previously downloaded export set without a network connection.
- **Made for automation.** Manage projects, languages, and translations from the CLI or through hosted MCP.

## Get started

[Install the CLI](https://docs.sentiary.com/getting-started/installation/), then configure your project and download translations:

```sh
sentiary-cli init
sentiary-cli pull --dry-run
sentiary-cli pull
```

## Explore

| Tool | Guide |
| --- | --- |
| CLI | [Commands](https://docs.sentiary.com/cli/reference/) · [Configuration](https://docs.sentiary.com/configuration/reference/) · [Mobile apps](https://docs.sentiary.com/cli/mobile/) |
| Gradle plugin | [Setup](https://docs.sentiary.com/gradle/) · [Android and Compose](https://docs.sentiary.com/gradle/guides/) |
| AI assistants | [Hosted MCP](https://docs.sentiary.com/mcp/#hosted-mcp) · [CLI skill](.agents/skills/sentiary-cli/SKILL.md) · [MCP skill](.agents/skills/sentiary-mcp/SKILL.md) |

[Development](https://docs.sentiary.com/contributing/development/) · [Documentation maintenance](https://docs.sentiary.com/contributing/documentation/) · [Troubleshooting](https://docs.sentiary.com/troubleshooting/)
