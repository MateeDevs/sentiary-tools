# Configuration reference

## Discovery and paths

`pull --config PATH` resolves PATH relative to the current working directory. Without it, the CLI searches the working directory and then its parents for the nearest `sentiary.toml`. It does not merge ancestor configs. The containing directory is resolved through symlinks and becomes the root for relative destinations. Absolute destinations are accepted, subject to ownership and path checks. Changing the config filename also changes its ownership manifest name.

All TOML fields are listed below. Unknown fields, invalid TOML, and unsupported types fail. There are no top-level global language defaults and no server URL field in TOML. The CLI defaults to `https://api.sentiary.com`; `SENTIARY_BASE_URL` selects a different API base URL (surrounding whitespace and trailing slashes are removed).

## Top level

| Field | Type | Required | Default / accepted values |
| --- | --- | --- | --- |
| `config_version` | integer | Yes | Must be `1` |
| `projects` | table of named project tables | Yes for referenced projects | Aliases must be nonempty |
| `outputs` | table of named output tables | Yes | At least one output; names must be nonempty |

## Project: `[projects.NAME]`

| Field | Type | Required | Default / behavior |
| --- | --- | --- | --- |
| `id` | string | Yes | Nonempty existing Sentiary project ID |
| `api_key` | string | Online if no user key | Project credential; lower priority than either user-key source |
| `languages` | array of strings or string `"*"` | Unless each output supplies it | Nonempty unique IDs, or all project languages |
| `source_language` | string | For combined output unless overridden | No default |
| `include_empty` | boolean | No | `false` |

## Output: `[outputs.NAME]`

| Field | Type | Required | Default / behavior |
| --- | --- | --- | --- |
| `project` | string | Yes | Defined project alias, not its remote ID |
| `format` | string | Yes | Exact token from the format table below |
| `layout` | string | No | `per_language`; also `combined` for `xcstrings` |
| `path` | string | Yes | Nonempty destination or template |
| `languages` | array of strings or `"*"` | No if project defines it | Replaces project selection in full |
| `source_language` | string | For combined unless inherited | Replaces project source; ignored for per-language files |
| `include_empty` | boolean | No | Replaces project value, including explicit `false` |
| `locale_map` | table of string values | When path contains `{locale}` | Target language ID to path component; no automatic mapping |
| `language_aliases` | table of string values | No | Target ID to source language ID; chains allowed |

`languages`, `source_language`, and `include_empty` inherit independently from the project. No list or map merging occurs. `locale_map` and `language_aliases` belong only to outputs. Missing languages fail; an empty array is not shorthand for all languages.

`{language}` expands to the exact selected target ID. `{locale}` expands using that target's locale map entry. No locale normalization happens automatically. IDs such as `cs`, `cs-CZ`, and `cs_CZ` are distinct. Use your project's actual IDs. For Android qualifiers such as `values-en-rGB` or Compose resource qualifiers, supply an explicit map that matches the resource consumer.

Aliases copy a complete exported file; they do not fill individual missing keys. For `en-AU → en-GB → en`, the target filename uses `en-AU` while the API/cache language is `en`. Every chain must end in a language returned by the project. Cycles, missing sources, and invalid IDs fail, including aliases not selected for output. Explicit selections include only listed target IDs; wildcard selections add all alias targets to all project languages.

## Formats and layouts

Every token below supports `per_language`:

| Family | Exact tokens |
| --- | --- |
| JSON | `json`, `key_value_json`, `i18next`, `poeditor_json`, `arb` |
| Platform resources | `android`, `compose`, `apple`, `xcstrings`, `stringsdict`, `resw`, `resx` |
| Gettext | `po`, `pot`, `mo` |
| Delimited / structured text | `csv`, `ini`, `properties`, `ts`, `yml` |
| XML interchange | `xliff`, `xlf`, `xliff_1_2`, `rise_360_xliff`, `xmb`, `xtb` |
| Spreadsheets | `xls`, `xlsx` |

Only `xcstrings` supports `combined`. That layout creates one catalog for all selected languages. It requires a fixed path, no placeholders, no `locale_map`, no `language_aliases`, and a source language included in the resolved selection. Combined catalogs support plain strings; plurals and device variants are outside the current exporter contract.

`include_empty = false` omits empty translations. For combined catalogs, `true` includes empty localizations with state `new`; it does not invent fallback text. `pull` preserves downloaded bytes, including binary spreadsheet/Gettext exports. Legacy `export` and `import` accept only `json`, `android`, `apple`, and `compose`.

## Local supplement and credentials

`sentiary.local.toml` is read beside the selected shared config, even when that config has another filename. It accepts one optional field, `user_api_key` (string). Copy the [local credential template](../examples/local-credentials.toml) to that filename and replace its placeholder; keep it out of source control. Unknown fields fail. It cannot override projects, outputs, server, or paths.

Online credentials for every project are selected in this order:

1. Nonblank `SENTIARY_USER_API_KEY`.
2. Nonblank local `user_api_key`.
3. That project's nonblank `api_key`.

An invalid selected key fails without falling back to another. All configured projects require credentials during online config loading, including projects with no selected output. `SENTIARY_PROJECT_API_KEY` is an `init` input only; `pull` does not read it directly. `SENTIARY_PROJECT_ID` applies to legacy API commands, not TOML project IDs. Legacy API commands do not load TOML or local credentials.

Offline loading ignores the local supplement entirely and permits missing credentials. Cache data contains translation content even though it contains no keys. Ignore `sentiary.local.toml` and `.sentiary/` in version control. Shared `api_key` retains the project's existing permissions; it is not a download-only key.

## Validation failures

Language IDs must be nonblank and unique. They cannot be `.`, `..`, or contain slash, backslash, braces, or NUL. Unknown path expressions and NUL paths fail. Missing locale entries fail when planning the resolved targets. Duplicate output selections, unknown output names, and combined catalogs whose source is not selected fail.

Destinations must be distinct ignoring case, and no destination may be another destination's parent directory. Case-only renames of owned paths fail. Config files, local credentials, `.sentiary/`, and symbolic links in destination paths are protected. A preview validates the download plan, but ownership and filesystem protections are checked during the actual write workflow. See [failure recovery](downloads.md#failure-recovery).
