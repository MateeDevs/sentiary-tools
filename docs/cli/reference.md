# CLI reference

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

The table for each command lists every registered option. Unless a row says otherwise, “No” required means optional; no environment fallback exists beyond the listed field. All non-`pull` formats use the legacy four-format list. `get` needs exactly one of `--string-id` or `--key`. For `add`, language is optional; when supplied it triggers a second request to create the initial translation. The current client ignores `--translation` if language is omitted. A translation failure can leave the new term created. For `edit`, nonempty values update fields; `--clear-description` / `--clear-context` explicitly clear them and take precedence over the corresponding text flags.

## list-projects

List projects.

No command-specific options. Use `--help` for help.

```sh
sentiary-cli list-projects
```

## get-project

Get a project and membership.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |

```sh
sentiary-cli get-project
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## create-project

Create a project.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--name` | string | Yes | `""` | Nonblank string | Last value wins | None | `--name=welcome` |

```sh
sentiary-cli create-project --name=welcome
```

## edit-project

Rename a project.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--name` | string | Yes | `""` | Nonblank string | Last value wins | None | `--name=welcome` |

```sh
sentiary-cli edit-project --name=welcome
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## remove-project

Remove a project and its data.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |

```sh
sentiary-cli remove-project
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## list-languages

List all supported languages.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--in-language` | string | No | `""` | Language ID for translated language names; empty uses authenticated user language | Last value wins | None | `--in-language=en` |

```sh
sentiary-cli list-languages
```

## list-project-languages

List project languages.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--in-language` | string | No | `""` | Language ID for translated language names; empty uses authenticated user language | Last value wins | None | `--in-language=en` |

```sh
sentiary-cli list-project-languages
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## add-project-language

Add a language to a project.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--language` | string | Yes | `""` | Exact API language ID (BCP 47 for project language management) | Last value wins | None | `--language=cs` |
| `--in-language` | string | No | `""` | Language ID for translated language names; empty uses authenticated user language | Last value wins | None | `--in-language=en` |

```sh
sentiary-cli add-project-language --language=cs
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## remove-project-language

Remove a language and its translations.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--language` | string | Yes | `""` | Exact API language ID (BCP 47 for project language management) | Last value wins | None | `--language=cs` |
| `--in-language` | string | No | `""` | Language ID for translated language names; empty uses authenticated user language | Last value wins | None | `--in-language=en` |

```sh
sentiary-cli remove-project-language --language=cs
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## list-project-members

List project members.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |

```sh
sentiary-cli list-project-members
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## set-project-member-role

Set a project member role.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--user-id` | string | Yes | `""` | Existing Sentiary user ID | Last value wins | None | `--user-id=USER_ID` |
| `--role` | string | Yes | `""` | administrator, developer, translator | Last value wins | None | `--role=developer` |

```sh
sentiary-cli set-project-member-role --user-id=USER_ID --role=developer
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## remove-project-member

Remove a project member.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--user-id` | string | Yes | `""` | Existing Sentiary user ID | Last value wins | None | `--user-id=USER_ID` |

```sh
sentiary-cli remove-project-member --user-id=USER_ID
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## list-invitations

List invitations for the current user.

No command-specific options. Use `--help` for help.

```sh
sentiary-cli list-invitations
```

## create-invitation

Invite a user to a project.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--user-email` | string | Yes | `""` | Email address | Last value wins | None | `--user-email=developer@example.com` |
| `--role` | string | Yes | `""` | administrator, developer, translator | Last value wins | None | `--role=developer` |

```sh
sentiary-cli create-invitation --user-email=developer@example.com --role=developer
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## accept-invitation

Accept an invitation.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--invitation-id` | string | Yes | `""` | Existing invitation ID | Last value wins | None | `--invitation-id=INVITATION_ID` |

```sh
sentiary-cli accept-invitation --invitation-id=INVITATION_ID
```

## decline-invitation

Decline an invitation.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--invitation-id` | string | Yes | `""` | Existing invitation ID | Last value wins | None | `--invitation-id=INVITATION_ID` |

```sh
sentiary-cli decline-invitation --invitation-id=INVITATION_ID
```

## list-project-invitations

List project invitations.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |

```sh
sentiary-cli list-project-invitations
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## remove-project-invitation

Remove a project invitation.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--invitation-id` | string | Yes | `""` | Existing invitation ID | Last value wins | None | `--invitation-id=INVITATION_ID` |

```sh
sentiary-cli remove-project-invitation --invitation-id=INVITATION_ID
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## list

List project strings.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--size` | int | No | `0` | Integer; list sends only positive values, search normalizes nonpositive to 25 | Last value wins | None | `--size=25` |
| `--cursor` | string | No | `""` | Opaque pagination cursor from prior response | Last value wins | None | `--cursor=CURSOR` |
| `--include-language` | string | No | `""` | Comma-separated language IDs; trims items and omits empty ones | Last value wins | None | `--include-language=en,cs` |
| `--filter` | string | No | `""` | ALL, TRANSLATED, NOT_TRANSLATED; passed to API | Last value wins | None | `--filter=NOT_TRANSLATED` |
| `--filter-language` | string | No | `""` | API language ID used by filter | Last value wins | None | `--filter-language=cs` |
| `--order` | string | No | `""` | UPDATED_ASC or UPDATED_DESC; passed to API | Last value wins | None | `--order=UPDATED_DESC` |

```sh
sentiary-cli list --size=25 --include-language=en,cs --filter=NOT_TRANSLATED --filter-language=cs --order=UPDATED_DESC
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

Use `pagination.next` from the JSON response as the next `--cursor`; the CLI does not automatically fetch every page.

## search

Search project strings.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--query` | string | Yes | `""` | Nonblank search text | Last value wins | None | `--query=welcome` |
| `--size` | int | No | `25` | Integer; list sends only positive values, search normalizes nonpositive to 25 | Last value wins | None | `--size=25` |
| `--cursor` | string | No | `""` | Opaque pagination cursor from prior response | Last value wins | None | `--cursor=CURSOR` |
| `--include-language` | string | No | `""` | Comma-separated language IDs; trims items and omits empty ones | Last value wins | None | `--include-language=en,cs` |
| `--filter` | string | No | `""` | ALL, TRANSLATED, NOT_TRANSLATED; passed to API | Last value wins | None | `--filter=NOT_TRANSLATED` |
| `--filter-language` | string | No | `""` | API language ID used by filter | Last value wins | None | `--filter-language=cs` |
| `--order` | string | No | `""` | UPDATED_ASC or UPDATED_DESC; passed to API | Last value wins | None | `--order=UPDATED_DESC` |

```sh
sentiary-cli search --query=welcome --size=25
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

Use `pagination.next` from the JSON response as the next `--cursor`; the CLI does not automatically fetch every page.

## get

Get a string by ID or key.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--string-id` | string | Alternative to key | `""` | Existing term ID | Last value wins | None | `--string-id=STRING_ID` |
| `--key` | string | Alternative to string-id | `""` | Exact key, mutually exclusive with string-id | Last value wins | None | `--key=welcome` |
| `--include-language` | string | No | `""` | Comma-separated language IDs; trims items and omits empty ones | Last value wins | None | `--include-language=en,cs` |

```sh
sentiary-cli get --key=welcome --include-language=en,cs
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## add

Add a string.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--name` | string | Yes | `""` | Nonblank string | Last value wins | None | `--name=welcome` |
| `--description` | string | No | `""` | String; empty omits update in edit | Last value wins | None | `--description=Greeting` |
| `--context` | string | No | `""` | String; empty omits update in edit | Last value wins | None | `--context=Home` |
| `--language` | string | No; needed to apply initial translation | `""` | Exact API language ID (BCP 47 for project language management) | Last value wins | None | `--language=cs` |
| `--translation` | string | No | `""` | Initial translation text; ignored without language | Last value wins | None | `--translation=Ahoj` |

```sh
sentiary-cli add --name=welcome --language=cs --translation=Ahoj
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## edit

Edit a string.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--string-id` | string | Yes | `""` | Existing term ID | Last value wins | None | `--string-id=STRING_ID` |
| `--name` | string | No | `""` | Nonempty string updates name; empty leaves unchanged | Last value wins | None | `--name=welcome` |
| `--description` | string | No | `""` | String; empty omits update in edit | Last value wins | None | `--description=Greeting` |
| `--context` | string | No | `""` | String; empty omits update in edit | Last value wins | None | `--context=Home` |
| `--clear-description` | bool | No | `false` | true or false; true clears description | Last value wins | None | `--clear-description=true` |
| `--clear-context` | bool | No | `false` | true or false; true clears context | Last value wins | None | `--clear-context=true` |

```sh
sentiary-cli edit --string-id=STRING_ID --clear-description
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## remove

Remove a string.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--string-id` | string | Yes | `""` | Existing term ID | Last value wins | None | `--string-id=STRING_ID` |

```sh
sentiary-cli remove --string-id=STRING_ID
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## set-translation

Set a translation.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--string-id` | string | Yes | `""` | Existing term ID | Last value wins | None | `--string-id=STRING_ID` |
| `--language` | string | Yes | `""` | Exact API language ID (BCP 47 for project language management) | Last value wins | None | `--language=cs` |
| `--text` | string | No (defaults to empty) | `""` | Translation text, including empty string | Last value wins | None | `--text=Ahoj` |

```sh
sentiary-cli set-translation --string-id=STRING_ID --language=cs --text=Ahoj
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## remove-translation

Remove a translation.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--string-id` | string | Yes | `""` | Existing term ID | Last value wins | None | `--string-id=STRING_ID` |
| `--language` | string | Yes | `""` | Exact API language ID (BCP 47 for project language management) | Last value wins | None | `--language=cs` |

```sh
sentiary-cli remove-translation --string-id=STRING_ID --language=cs
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## info

Get project string information.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |

```sh
sentiary-cli info
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## export

Export project strings.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--format` | string | Yes | `""` | json, android, apple, compose (trimmed and case-normalized) | Last value wins | None | `--format=json` |
| `--language` | string | Yes | `""` | Exact API language ID (BCP 47 for project language management) | Last value wins | None | `--language=cs` |
| `--include-empty` | bool | No | `false` | true or false | Last value wins | None | `--include-empty=true` |
| `--output` | string | No | `"-"` | File path or - for stdout | Last value wins | None | `--output=strings.json` |

```sh
sentiary-cli export --format=json --language=cs --output=strings.json
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.

## init

Create a config for existing projects.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--directory` | string | No | `"."` | Existing target directory for init | Last value wins | None | `--directory=.` |

```sh
sentiary-cli init --directory=.
```

## pull

Download configured translation outputs.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--config` | string | No | `""` | Config file path; empty searches current directory and parents | Last value wins | None | `--config=sentiary.toml` |
| `--offline` | bool | No | `false` | true or false; restore from one complete cached set | Last value wins | None | `--offline=true` |
| `--cache-dir` | string | No | `""` | Directory path; empty selects environment/OS cache | Last value wins | SENTIARY_CACHE_DIR | `--cache-dir=/tmp/sentiary-cache` |
| `--dry-run` | bool | No | `false` | true or false; emit preview without generated-file writes | Last value wins | None | `--dry-run=true` |
| `--json` | bool | No | `false` | true or false; emit protocol result on successful non-preview pull | Last value wins | None | `--json=true` |
| `--integration-request` | string | No | `""` | Protocol 1 JSON file path; cannot combine with output selection | Last value wins | None | `--integration-request=request.json` |
| `--force` | bool | No | `false` | true or false; adopt existing files / discard edits | Last value wins | None | `--force=true` |
| `--output` | string list | No | `all outputs` | Defined output name; omit for all; duplicate selections fail | Accumulates | None | `--output=android` |

```sh
sentiary-cli pull --output=android --output=compose --dry-run
```

For ownership, cache location, conditional requests, cleanup, and failure semantics, see [downloads](../configuration/downloads.md). For machine requests/results see [protocol](../integration/protocol.md). `--dry-run` takes preview output precedence over `--json`.

## import

Import project strings.

| Option | Type | Required | Default | Accepted values / behavior | Repeatability | Environment fallback | Example |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `--project-id` | string | Required after fallback | `""` | Existing project ID | Last value wins | SENTIARY_PROJECT_ID | `--project-id=PROJECT_ID` |
| `--format` | string | Yes | `""` | json, android, apple, compose (trimmed and case-normalized) | Last value wins | None | `--format=json` |
| `--language` | string | Yes | `""` | Exact API language ID (BCP 47 for project language management) | Last value wins | None | `--language=cs` |
| `--input` | string | No | `"-"` | File path or - for stdin | Last value wins | None | `--input=strings.json` |
| `--override-existing` | bool | No | `false` | true or false; true permits replacing existing translations | Last value wins | None | `--override-existing=true` |

```sh
sentiary-cli import --format=json --language=cs --input=strings.json
```

The example assumes `SENTIARY_PROJECT_ID=PROJECT_ID`; `--project-id=PROJECT_ID` overrides it.
