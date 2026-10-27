---
tags:
  - CLI
  - Configuration
---

# First configured download

Start with an existing Sentiary project ID and its language IDs. The current `init` wizard accepts those IDs locally; it does not browse or create remote projects.

```sh
export SENTIARY_USER_API_KEY='YOUR_USER_API_KEY'
sentiary-cli init --directory .
```

The directory must already exist. The wizard asks for project aliases, IDs, language selections, output names, formats, paths, and locale mappings. It supports multiple projects and outputs. For combined Xcode catalogs it asks for a selected source language. Enter `cancel` at a prompt or end stdin to abort. It refuses to overwrite `sentiary.toml` or a local credentials file that it would create.

The defaults produce Android/Compose paths under `build/sentiary/`, or a combined `Localizable.xcstrings`. Generic formats require an explicit path. The preview redacts project keys and requires a final `y` before writing. Saving a project key from `SENTIARY_PROJECT_API_KEY` in shared TOML or a user key from `SENTIARY_USER_API_KEY` in local TOML requires explicit opt-in. The default uses environment credentials. New config files use mode `0600` where supported. Ignore entries for `sentiary.local.toml` and `.sentiary/` are offered by default.

Alternatively, copy a [mobile configuration](../examples/sentiary.toml), replace project IDs, and edit languages and paths. Do not commit personal credentials. The local supplement supports only `user_api_key`.

```sh
sentiary-cli pull --dry-run
sentiary-cli pull
sentiary-cli pull --output android
```

The preview prints resolved file paths as JSON. Wildcard languages and aliases require a read request to resolve project languages. It does not download exports or write generated files. An ordinary online preview still requires credentials.

A pull downloads and verifies the full selection, publishes a completed cache set, then replaces generated files. Existing files need deliberate [adoption](../configuration/downloads.md#ownership-and-adoption). Review the plan and affected files before using `--force`.

```sh
sentiary-cli pull --offline
```

Offline mode restores the newest single completed set covering the request. It needs no credentials or network, but needs a previous successful online pull and intact cache objects. See [offline recovery](../configuration/downloads.md#restore-deleted-generated-files).
