---
tags:
  - Configuration
---

# Configuration reference

```sh
sentiary-cli init
sentiary-cli pull --dry-run
```

The wizard writes `sentiary.toml`. Start from the [mobile configuration](../examples/sentiary.toml) or [advanced configuration](../examples/advanced.toml) for an editable example. See [languages and aliases](languages.md) for wildcard and multi-project usage.

## Discovery and paths

`pull --config PATH` resolves PATH relative to the current working directory. Without it, the CLI searches the working directory and then its parents for the nearest `sentiary.toml`. It does not merge ancestor configs. The containing directory is resolved through symlinks and becomes the root for relative destinations. Absolute destinations are accepted, subject to ownership and path checks. Changing the config filename also changes its ownership manifest name.

All TOML fields are listed below. Unknown fields, invalid TOML, and unsupported types fail. There are no top-level global language defaults and no server URL field in TOML. The CLI defaults to `https://api.sentiary.com`; `SENTIARY_BASE_URL` selects a different API base URL (surrounding whitespace and trailing slashes are removed).

## Top level

| Field | Type | Required | Default / accepted values | Example |
| --- | --- | --- | --- | --- |
| `config_version` | integer | Yes | Must be `1` | `config_version = 1` |
| `projects` | table of named project tables | Yes for referenced projects | Aliases must be nonempty | `[projects.mobile]` |
| `outputs` | table of named output tables | Yes | At least one output; names must be nonempty | `[outputs.android]` |

## Project: `[projects.NAME]`

| Field | Type | Required | Default / behavior | Example |
| --- | --- | --- | --- | --- |
| `id` | string | Yes | Nonempty existing Sentiary project ID | `id = "PROJECT_ID"` |
| `api_key` | string | Online if no user key | Empty when omitted; project credential; lower priority than either user-key source | `api_key = "PROJECT_KEY"` |
| `languages` | array of strings or string `"*"` | Unless each output supplies it | Nonempty unique IDs, or all project languages | `languages = ["en", "cs"]` |
| `source_language` | string | For combined output unless overridden | No default | `source_language = "en"` |
| `include_empty` | boolean | No | `false` | `include_empty = false` |

## Output: `[outputs.NAME]`

| Field | Type | Required | Default / behavior | Example |
| --- | --- | --- | --- | --- |
| `project` | string | Yes | Defined project alias, not its remote ID | `project = "mobile"` |
| `format` | string | Yes | Exact token from the format table below | `format = "android"` |
| `layout` | string | No | `per_language`; also `combined` for `xcstrings` | `layout = "per_language"` |
| `path` | string | Yes | Nonempty destination or template | `path = "build/resources/{locale}/strings.xml"` |
| `languages` | array of strings or `"*"` | No if project defines it | Replaces project selection in full | `languages = ["en", "cs"]` |
| `source_language` | string | For combined unless inherited | Replaces project source; ignored for per-language files | `source_language = "en"` |
| `include_empty` | boolean | No | Replaces project value, including explicit `false` | `include_empty = false` |
| `locale_map` | table of string values | When path contains `{locale}` | Empty map; target language ID to path component; no automatic mapping | `locale_map = { en = "values" }` |
| `language_aliases` | table of string values | No | Empty map; target ID to source language ID; chains allowed | `language_aliases = { en-GB = "en" }` |

`languages`, `source_language`, and `include_empty` inherit independently from the project. No list or map merging occurs. `locale_map` and `language_aliases` belong only to outputs. Missing languages fail; an empty array is not shorthand for all languages.

`{language}` expands to the exact selected target ID. `{locale}` expands using that target's locale map entry. No locale normalization happens automatically. IDs such as `cs`, `cs-CZ`, and `cs_CZ` are distinct. Use your project's actual IDs. For Android qualifiers such as `values-en-rGB` or Compose resource qualifiers, supply an explicit map that matches the resource consumer.

Aliases copy a complete exported file; they do not fill individual missing keys. For `en-AU → en-GB → en`, the target filename uses `en-AU` while the API/cache language is `en`. Every chain must end in a language returned by the project. Cycles, missing sources, and invalid IDs fail, including aliases not selected for output. Explicit selections include only listed target IDs; wildcard selections add all alias targets to all project languages.

## Formats and layouts

Every token below supports `per_language`:

There is no default format; set `format` explicitly. All formats default to `per_language` layout. Set the output `path` explicitly, for example:

| Format | Use | Example `path` |
| --- | --- | --- |
| `json` | Sentiary JSON | `"build/locales/{language}.json"` |
| `key_value_json` | Flat key-value JSON | `"build/locales/{language}.json"` |
| `i18next` | i18next translations | `"build/locales/{language}.json"` |
| `poeditor_json` | POEditor JSON | `"build/locales/{language}.json"` |
| `arb` | Application Resource Bundle | `"build/locales/{language}.arb"` |
| `android` | Android string resources | `"build/resources/{locale}/strings.xml"` |
| `compose` | Compose string resources | `"build/resources/{locale}/strings.xml"` |
| `apple` | Apple strings | `"build/locales/{language}.strings"` |
| `xcstrings` | Xcode string catalog | `"build/locales/{language}.xcstrings"` |
| `stringsdict` | Apple string dictionary | `"build/locales/{language}.stringsdict"` |
| `resw` | Windows resources | `"build/locales/{language}.resw"` |
| `resx` | .NET resources | `"build/locales/{language}.resx"` |
| `po` | Gettext translations | `"build/locales/{language}.po"` |
| `pot` | Gettext template | `"build/locales/{language}.pot"` |
| `mo` | Compiled Gettext | `"build/locales/{language}.mo"` |
| `csv` | Comma-separated values | `"build/locales/{language}.csv"` |
| `ini` | INI translations | `"build/locales/{language}.ini"` |
| `properties` | Java properties | `"build/locales/{language}.properties"` |
| `ts` | Qt translation source | `"build/locales/{language}.ts"` |
| `yml` | YAML translations | `"build/locales/{language}.yml"` |
| `xliff` | XLIFF interchange | `"build/locales/{language}.xliff"` |
| `xlf` | XLF interchange | `"build/locales/{language}.xlf"` |
| `xliff_1_2` | XLIFF 1.2 interchange | `"build/locales/{language}.xliff"` |
| `rise_360_xliff` | Rise 360 XLIFF | `"build/locales/{language}.xliff"` |
| `xmb` | XML message bundle | `"build/locales/{language}.xmb"` |
| `xtb` | XML translation bundle | `"build/locales/{language}.xtb"` |
| `xls` | Excel binary spreadsheet | `"build/locales/{language}.xls"` |
| `xlsx` | Excel workbook | `"build/locales/{language}.xlsx"` |

Only `xcstrings` supports `combined`. That layout creates one catalog for all selected languages. It requires a fixed path, no placeholders, no `locale_map`, no `language_aliases`, and a source language included in the resolved selection. Combined catalogs support plain strings; plurals and device variants are outside the current exporter contract.

`include_empty = false` omits empty translations. For combined catalogs, `true` includes empty localizations with state `new`; it does not invent fallback text. `pull` preserves downloaded bytes, including binary spreadsheet/Gettext exports. Legacy `export` and `import` accept only `json`, `android`, `apple`, and `compose`.

## Local supplement and credentials

`sentiary.local.toml` is read beside the selected shared config, even when that config has another filename. It accepts one optional field, `user_api_key` (string, default empty), for example `user_api_key = "YOUR_USER_API_KEY"`. Copy the [local credential template](../examples/local-credentials.toml) to that filename and replace its placeholder; keep it out of source control. Unknown fields fail. It cannot override projects, outputs, server, or paths.

Online credentials for every project are selected in this order:

1. Nonblank `SENTIARY_USER_API_KEY`.
2. Nonblank local `user_api_key`.
3. That project's nonblank `api_key`.

An invalid selected key fails without falling back to another. All configured projects require credentials during online config loading, including projects with no selected output. `SENTIARY_PROJECT_API_KEY` is an `init` input only; `pull` does not read it directly. `SENTIARY_PROJECT_ID` applies to legacy API commands, not TOML project IDs. Legacy API commands do not load TOML or local credentials.

Offline loading ignores the local supplement entirely and permits missing credentials. Cache data contains translation content even though it contains no keys. Ignore `sentiary.local.toml` and `.sentiary/` in version control. Shared `api_key` retains the project's existing permissions; it is not a download-only key.

## Validation failures

Language IDs must be nonblank and unique. They cannot be `.`, `..`, or contain slash, backslash, braces, or NUL. Unknown path expressions and NUL paths fail. Missing locale entries fail when planning the resolved targets. Duplicate output selections, unknown output names, and combined catalogs whose source is not selected fail.

Destinations must be distinct ignoring case, and no destination may be another destination's parent directory. Case-only renames of owned paths fail. Config files, local credentials, `.sentiary/`, and symbolic links in destination paths are protected. A preview validates the download plan, but ownership and filesystem protections are checked during the actual write workflow. See [failure recovery](downloads.md#failure-recovery).
