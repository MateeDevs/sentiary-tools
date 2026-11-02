---
tags:
  - CLI
  - Configuration
---

# First configured download

Start with an existing Sentiary project ID, its language IDs, and an API key with access to the project.

## Create a configuration

```sh
export SENTIARY_USER_API_KEY='YOUR_USER_API_KEY'
sentiary-cli init --directory .
```

The wizard asks for project IDs, languages, output formats, and destination paths.
Review its preview and confirm to create `sentiary.toml`. It does not create remote projects
or overwrite an existing configuration. See the [init reference](../cli/reference.md#init)
for wizard options.

Alternatively, copy the [mobile configuration](../examples/sentiary.toml) and edit its
project ID, credentials, languages, and paths. Keep secrets out of committed files;
use [environment references](../configuration/reference.md#project-environment-credentials)
for shared configurations.

## Preview and download

```sh
sentiary-cli pull --dry-run
sentiary-cli pull
sentiary-cli pull --output android
```

The preview lists destination files without writing them. Online previews require credentials.
A pull downloads and verifies all selected exports before replacing generated files.
If a destination already exists, review [file adoption](../configuration/downloads.md#ownership-and-adoption)
before using `--force`, which can discard local edits.

## Restore offline

```sh
sentiary-cli pull --offline
```

After a successful online pull, offline mode can restore matching files without
credentials or a network connection. Keep the cache intact. See
[offline recovery](../configuration/downloads.md#restore-deleted-generated-files).
