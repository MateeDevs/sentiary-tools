---
tags:
  - CLI
---

# CLI reference

```sh
sentiary-cli init
sentiary-cli pull --dry-run
sentiary-cli pull
```

Use [mobile downloads](mobile.md) for Android, Compose, and Xcode examples. The reference below covers every command and option.

## Invocation, credentials, and streams

Use `sentiary-cli COMMAND [OPTIONS]`. All options below belong after the command. There are no global credential, server, or JSON switches. Go flag syntax accepts `--flag=value` and `--flag value` (also single-dash forms). Boolean flags accept `--flag` for true or `--flag=false`; do not use a separate boolean word. Positional arguments are rejected. Scalar options may be repeated, with the last value winning; only `pull --output` accumulates multiple values. Comma-separated language lists are one scalar option.

`SENTIARY_BASE_URL` optionally overrides the default `https://api.sentiary.com` for API commands and online pulls; it also scopes the export cache. All API commands require `SENTIARY_USER_API_KEY`; project commands use `--project-id` then `SENTIARY_PROJECT_ID`. They do not load TOML or project keys. Authorization and project-role requirements are enforced by the API: a user key does not grant access beyond its owner's permissions. `pull` instead uses [TOML credential precedence](../configuration/reference.md#local-supplement-and-credentials). `init`, help, and version need no network credentials; offline pull needs none.

Management and string commands write indented JSON to stdout, including mutations. `export` writes raw text to stdout or its output file; `import` reads a file or all of stdin and writes a JSON import result. `pull` normally prints `Downloaded N file(s).`, `--dry-run` prints a JSON array, and `--json` prints a protocol result. `init` reads line-oriented answers from stdin, sends prompts/redacted preview to stderr, and prints completion to stdout. Help and errors go to stderr. Separate stdout and stderr when scripting.

Exit `0` means success/help/version; `2` means a CLI usage error (unknown command/flag, missing CLI-required option, unexpected positional argument, invalid version arguments); `1` means another failure, including config, authentication, API, filesystem, cancellation, or init cancellation. Some required values are checked by the API client and therefore produce `1`, not `2`. A nonzero pull has no successful JSON result; check exit status before parsing stdout.

Examples use placeholder IDs and credentials. Commands that create, edit, remove, invite, import, or set translations mutate remote data and run without a confirmation prompt. Help examples are safe locally.

## Common help and version options

| Form | Type / required / default | Accepted values, repeatability, fallback | Example |
| --- | --- | --- | --- |
| `help`, `--help`, `-h` as command | Command; no options required | Prints top-level help; trailing arguments currently ignored; no environment fallback | `sentiary-cli help` |
| `COMMAND --help` or `COMMAND -h` | Boolean help flag; optional, false | Built-in Go flag help; no fallback | `sentiary-cli pull --help` |
| `version`, `--version`, `-v` as command | Command; no version argument | Prints version; optional exact `--json` second argument only | `sentiary-cli version` |
| `version --json` | Literal switch; optional, off | Not repeatable; no environment fallback | `sentiary-cli version --json` |

The table for each command lists every registered option. Each row lists its type, requirement, default, and example. Options have no environment fallback unless stated. Scalar options use the last supplied value; `pull --output` accumulates selections. All non-`pull` formats use the legacy four-format list. `get` needs exactly one of `--string-id` or `--key`. For `add`, language is optional; when supplied it triggers a second request to create the initial translation. The current client ignores `--translation` if language is omitted. A translation failure can leave the new term created. For `edit`, nonempty values update fields; `--clear-description` / `--clear-context` explicitly clear them and take precedence over the corresponding text flags.

## list-projects

List projects.

No command-specific options. Use `--help` for help.

```sh
sentiary-cli list-projects
```

## get-project

Get a project and membership.

```sh
sentiary-cli get-project
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## create-project

Create a project.

```sh
sentiary-cli create-project --name=welcome
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--name` | Nonblank string | string; Required; `""` | `--name=welcome` |

## edit-project

Rename a project.

```sh
sentiary-cli edit-project --name=welcome
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--name` | Nonblank string | string; Required; `""` | `--name=welcome` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## remove-project

Remove a project and its data.

```sh
sentiary-cli remove-project
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## list-languages

List all supported languages.

```sh
sentiary-cli list-languages
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--in-language` | Language ID for translated language names; empty uses authenticated user language | string; Optional; `""` | `--in-language=en` |

## list-project-languages

List project languages.

```sh
sentiary-cli list-project-languages
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--in-language` | Language ID for translated language names; empty uses authenticated user language | string; Optional; `""` | `--in-language=en` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## add-project-language

Add a language to a project.

```sh
sentiary-cli add-project-language --language=cs
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--language` | Exact API language ID (BCP 47 for project language management) | string; Required; `""` | `--language=cs` |
| `--in-language` | Language ID for translated language names; empty uses authenticated user language | string; Optional; `""` | `--in-language=en` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## remove-project-language

Remove a language and its translations.

```sh
sentiary-cli remove-project-language --language=cs
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--language` | Exact API language ID (BCP 47 for project language management) | string; Required; `""` | `--language=cs` |
| `--in-language` | Language ID for translated language names; empty uses authenticated user language | string; Optional; `""` | `--in-language=en` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## list-project-members

List project members.

```sh
sentiary-cli list-project-members
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## set-project-member-role

Set a project member role.

```sh
sentiary-cli set-project-member-role --user-id=USER_ID --role=developer
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--user-id` | Existing Sentiary user ID | string; Required; `""` | `--user-id=USER_ID` |
| `--role` | administrator, developer, translator | string; Required; `""` | `--role=developer` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## remove-project-member

Remove a project member.

```sh
sentiary-cli remove-project-member --user-id=USER_ID
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--user-id` | Existing Sentiary user ID | string; Required; `""` | `--user-id=USER_ID` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## list-invitations

List invitations for the current user.

No command-specific options. Use `--help` for help.

```sh
sentiary-cli list-invitations
```

## create-invitation

Invite a user to a project.

```sh
sentiary-cli create-invitation --user-email=developer@example.com --role=developer
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--user-email` | Email address | string; Required; `""` | `--user-email=developer@example.com` |
| `--role` | administrator, developer, translator | string; Required; `""` | `--role=developer` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## accept-invitation

Accept an invitation.

```sh
sentiary-cli accept-invitation --invitation-id=INVITATION_ID
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--invitation-id` | Existing invitation ID | string; Required; `""` | `--invitation-id=INVITATION_ID` |

## decline-invitation

Decline an invitation.

```sh
sentiary-cli decline-invitation --invitation-id=INVITATION_ID
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--invitation-id` | Existing invitation ID | string; Required; `""` | `--invitation-id=INVITATION_ID` |

## list-project-invitations

List project invitations.

```sh
sentiary-cli list-project-invitations
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## remove-project-invitation

Remove a project invitation.

```sh
sentiary-cli remove-project-invitation --invitation-id=INVITATION_ID
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--invitation-id` | Existing invitation ID | string; Required; `""` | `--invitation-id=INVITATION_ID` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## list

List project strings.

```sh
sentiary-cli list --size=25 --include-language=en,cs --filter=NOT_TRANSLATED --filter-language=cs --order=UPDATED_DESC
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--size` | Integer; list sends only positive values, search normalizes nonpositive to 25 | int; Optional; `0` | `--size=25` |
| `--cursor` | Opaque pagination cursor from prior response | string; Optional; `""` | `--cursor=CURSOR` |
| `--include-language` | Comma-separated language IDs; trims items and omits empty ones | string; Optional; `""` | `--include-language=en,cs` |
| `--filter` | ALL, TRANSLATED, NOT_TRANSLATED; passed to API | string; Optional; `""` | `--filter=NOT_TRANSLATED` |
| `--filter-language` | API language ID used by filter | string; Optional; `""` | `--filter-language=cs` |
| `--order` | UPDATED_ASC or UPDATED_DESC; passed to API | string; Optional; `""` | `--order=UPDATED_DESC` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

Use `pagination.next` from the JSON response as the next `--cursor`; the CLI does not automatically fetch every page.

## search

Search project strings.

```sh
sentiary-cli search --query=welcome --size=25
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--query` | Nonblank search text | string; Required; `""` | `--query=welcome` |
| `--size` | Integer; list sends only positive values, search normalizes nonpositive to 25 | int; Optional; `25` | `--size=25` |
| `--cursor` | Opaque pagination cursor from prior response | string; Optional; `""` | `--cursor=CURSOR` |
| `--include-language` | Comma-separated language IDs; trims items and omits empty ones | string; Optional; `""` | `--include-language=en,cs` |
| `--filter` | ALL, TRANSLATED, NOT_TRANSLATED; passed to API | string; Optional; `""` | `--filter=NOT_TRANSLATED` |
| `--filter-language` | API language ID used by filter | string; Optional; `""` | `--filter-language=cs` |
| `--order` | UPDATED_ASC or UPDATED_DESC; passed to API | string; Optional; `""` | `--order=UPDATED_DESC` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

Use `pagination.next` from the JSON response as the next `--cursor`; the CLI does not automatically fetch every page.

## get

Get a string by ID or key.

```sh
sentiary-cli get --key=welcome --include-language=en,cs
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--string-id` | Existing term ID | string; Alternative to key; `""` | `--string-id=STRING_ID` |
| `--key` | Exact key, mutually exclusive with string-id | string; Alternative to string-id; `""` | `--key=welcome` |
| `--include-language` | Comma-separated language IDs; trims items and omits empty ones | string; Optional; `""` | `--include-language=en,cs` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## add

Add a string.

```sh
sentiary-cli add --name=welcome --language=cs --translation=Ahoj
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--name` | Nonblank string | string; Required; `""` | `--name=welcome` |
| `--description` | String; empty omits update in edit | string; Optional; `""` | `--description=Greeting` |
| `--context` | String; empty omits update in edit | string; Optional; `""` | `--context=Home` |
| `--language` | Exact API language ID (BCP 47 for project language management) | string; No; needed to apply initial translation; `""` | `--language=cs` |
| `--translation` | Initial translation text; ignored without language | string; Optional; `""` | `--translation=Ahoj` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## edit

Edit a string.

```sh
sentiary-cli edit --string-id=STRING_ID --clear-description
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--string-id` | Existing term ID | string; Required; `""` | `--string-id=STRING_ID` |
| `--name` | Nonempty string updates name; empty leaves unchanged | string; Optional; `""` | `--name=welcome` |
| `--description` | String; empty omits update in edit | string; Optional; `""` | `--description=Greeting` |
| `--context` | String; empty omits update in edit | string; Optional; `""` | `--context=Home` |
| `--clear-description` | true or false; true clears description | bool; Optional; `false` | `--clear-description=true` |
| `--clear-context` | true or false; true clears context | bool; Optional; `false` | `--clear-context=true` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## remove

Remove a string.

```sh
sentiary-cli remove --string-id=STRING_ID
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--string-id` | Existing term ID | string; Required; `""` | `--string-id=STRING_ID` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## set-translation

Set a translation.

```sh
sentiary-cli set-translation --string-id=STRING_ID --language=cs --text=Ahoj
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--string-id` | Existing term ID | string; Required; `""` | `--string-id=STRING_ID` |
| `--language` | Exact API language ID (BCP 47 for project language management) | string; Required; `""` | `--language=cs` |
| `--text` | Translation text, including empty string | string; No (defaults to empty); `""` | `--text=Ahoj` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## remove-translation

Remove a translation.

```sh
sentiary-cli remove-translation --string-id=STRING_ID --language=cs
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--string-id` | Existing term ID | string; Required; `""` | `--string-id=STRING_ID` |
| `--language` | Exact API language ID (BCP 47 for project language management) | string; Required; `""` | `--language=cs` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## info

Get project string information.

```sh
sentiary-cli info
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## export

Export project strings.

```sh
sentiary-cli export --format=json --language=cs --output=strings.json
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--format` | json, android, apple, compose (trimmed and case-normalized) | string; Required; `""` | `--format=json` |
| `--language` | Exact API language ID (BCP 47 for project language management) | string; Required; `""` | `--language=cs` |
| `--include-empty` | true or false | bool; Optional; `false` | `--include-empty=true` |
| `--output` | File path or - for stdout | string; Optional; `"-"` | `--output=strings.json` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## init

Create a config for existing projects.

```sh
sentiary-cli init --directory=.
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--directory` | Existing target directory for init | string; Optional; `"."` | `--directory=.` |

## pull

Download configured translation outputs.

```sh
sentiary-cli pull --output=android --output=compose --dry-run
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--config` | Config file path; empty searches current directory and parents | string; Optional; `""` | `--config=sentiary.toml` |
| `--offline` | true or false; restore from one complete cached set | bool; Optional; `false` | `--offline=true` |
| `--cache-dir` | Directory path; empty selects environment/OS cache; fallback: `SENTIARY_CACHE_DIR` | string; Optional; `""` | `--cache-dir=/tmp/sentiary-cache` |
| `--dry-run` | true or false; emit preview without generated-file writes | bool; Optional; `false` | `--dry-run=true` |
| `--json` | true or false; emit protocol result on successful non-preview pull | bool; Optional; `false` | `--json=true` |
| `--integration-request` | Protocol 1 JSON file path; cannot combine with output selection | string; Optional; `""` | `--integration-request=request.json` |
| `--force` | true or false; adopt existing files / discard edits | bool; Optional; `false` | `--force=true` |
| `--output` | Defined output name; omit for all; duplicate selections fail | string list; Optional; `all outputs` | `--output=android` |

For ownership, cache location, conditional requests, cleanup, and failure semantics, see [downloads](../configuration/downloads.md). For machine requests/results see [protocol](https://github.com/MateeDevs/sentiary-tools/blob/staging/contributing/integration-protocol.md). `--dry-run` takes preview output precedence over `--json`.

## import

Import project strings.

```sh
sentiary-cli import --format=json --language=cs --input=strings.json
```

| Option | Purpose and accepted values | Type / requirement / default | Example |
| --- | --- | --- | --- |
| `--project-id` | Existing project ID; fallback: `SENTIARY_PROJECT_ID` | string; Required after fallback; `""` | `--project-id=PROJECT_ID` |
| `--format` | json, android, apple, compose (trimmed and case-normalized) | string; Required; `""` | `--format=json` |
| `--language` | Exact API language ID (BCP 47 for project language management) | string; Required; `""` | `--language=cs` |
| `--input` | File path or - for stdin | string; Optional; `"-"` | `--input=strings.json` |
| `--override-existing` | true or false; true permits replacing existing translations | bool; Optional; `false` | `--override-existing=true` |

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## Environment variables

| Variable | Purpose and default | Example |
| --- | --- | --- |
| `SENTIARY_USER_API_KEY` | Account credential; unset by default. Required for API commands; highest-priority credential for online pulls. | `export SENTIARY_USER_API_KEY='YOUR_USER_API_KEY'` |
| `SENTIARY_PROJECT_ID` | Project fallback for API commands; no default. | `export SENTIARY_PROJECT_ID='PROJECT_ID'` |
| `SENTIARY_BASE_URL` | API server; defaults to `https://api.sentiary.com`. | `export SENTIARY_BASE_URL='https://api.staging.sentiary.coolify.matee.cz'` |
| `SENTIARY_CACHE_DIR` | Export cache directory; defaults to the [OS cache path](../configuration/downloads.md#cache-location-and-identity). | `export SENTIARY_CACHE_DIR='.ci-cache/sentiary'` |
| `SENTIARY_PROJECT_API_KEY` | Optional key offered by `init`; no default and no direct `pull` fallback. | `export SENTIARY_PROJECT_API_KEY='PROJECT_KEY'` |

Installer-only `INSTALL_DIR` and `SENTIARY_VERSION` are covered in [installation](../getting-started/installation.md).

## JSON output for scripts

`version --json` reports the CLI version and supported integration protocol versions. `pull --json` returns a versioned list of generated files and checksums on success. `pull --dry-run` returns a preview array, including when combined with `--json`. Check the exit status before reading JSON; errors go to stderr.

`--integration-request` accepts a JSON file selecting outputs and validating resource roots. It cannot be combined with `--output`. Most CLI users do not need this option. Integration authors should use the [complete protocol](https://github.com/MateeDevs/sentiary-tools/blob/staging/contributing/integration-protocol.md) and [JSON Schema bundle](../integration/schemas.json).
