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
| `api_key` | string or table | No | Literal credential or explicit environment reference; reference table must contain only a string `env` field | `api_key = "PROJECT_KEY"` or `api_key = { env = "SENTIARY_MOBILE_API_KEY" }` |
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
| `keys_file` | string | No | UTF-8 exact-key selection file; an empty file explicitly selects no keys | `keys_file = "localization/keys.txt"` |
| `contexts` | array of strings | No | Exact, case-sensitive contexts; union with selected keys | `contexts = ["checkout"]` |
| `required_keys_file` | string | No | Selected keys that need nonblank translations in every target language | `required_keys_file = "localization/required.txt"` |
| `plural_bases_file` | string | No | Opt-in plural bases for Android/Compose XML only | `plural_bases_file = "localization/plurals.txt"` |

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

`include_empty = false` omits empty translations. For combined catalogs, `true` includes empty localizations with state `new`; it does not invent fallback text. `pull` preserves downloaded bytes, including binary spreadsheet/Gettext exports. Standalone `export` and `import` accept only `json`, `android`, `apple`, and `compose`.

## Selection, required translations, and plurals

These four options belong to each output, not the project. They require a backend
supporting POST resource exports. List paths resolve relative to the directory
containing `sentiary.toml`, not the current working
directory. Multiple outputs may reuse one list. Files use UTF-8, one literal key
per line. Surrounding whitespace is trimmed and blank lines are ignored. There
are no comments or escapes: `#example` is a literal key. Duplicate entries fail
with a file and line number. Missing or invalid files also fail offline pulls.
Each list accepts up to 10,000 unique items of at most 1,024 UTF-8 bytes, without
NUL. List files are limited to 10,260,000 bytes, including whitespace and line endings.

Exact keys and contexts form a union. Without selectors, the full catalog is
exported. An explicitly empty `keys_file` is not shorthand for the full catalog;
contexts can still select terms. Unknown keys, contexts, or plural bases and an
empty term selection fail. Contexts must not be blank and are case-sensitive.
Terms without a context can be selected by key. Required keys do not expand the
selection. They must exist, be selected, and have nonblank translations in every
target language. `include_empty = true` does not weaken this rule. Empty required
and plural files are valid and declare nothing.

For example, in an Android or Compose output:

```toml
keys_file = "localization/selected.txt"
contexts = ["checkout"]
required_keys_file = "localization/required.txt"
plural_bases_file = "localization/plurals.txt"
```

Put `truck_count` in the plural list and use catalog terms `truck_count.one`,
`truck_count.other`, and optionally `.zero`, `.two`, `.few`, or `.many`. Select
and require the base `truck_count`, never an individual quantity. A context that
selects any member selects the entire group. The entire prefix is reserved:
`truck_count.description` is an error, as is a plain `truck_count` string.
Names must match `[a-z_][a-z0-9_]*`; exporters do not rename resources.

Each emitted group needs nonblank `other`. Empty optional groups are omitted
with `include_empty = false`; nonempty groups without usable `other` fail.
With `include_empty = true`, an empty `other` still fails. Required plurals need
`other` in every language, not identical quantity sets. No language fallback is applied.
Without plural bases, dotted terms remain ordinary strings. Other string keys retain
their spelling; the naming restriction above applies only to declared plural bases.
Structured plural import and Apple plural assembly are not supported.

### Formatting arguments

Android and Compose exports convert `%@` to `%s` and `%1$@` to `%1$s`.
Compose also numbers unnumbered `%s` and `%d` in order: `%s: %d` becomes
`%1$s: %2$d`. Explicit positions are left unchanged. These rules apply equally
to ordinary strings and plural items.

Use explicit positions consistently when a translation has multiple arguments.
Width and precision forms such as `%04d` and `%.2f` are not automatically numbered.
Mixing explicit and implicit positions can produce conflicting argument numbers.
Sentiary does not validate argument compatibility; check generated resources with
your platform's build tools and formatting APIs.

The application passes the count used to select a plural quantity separately from
formatting arguments. Quantities do not have to contain the same arguments.

### Download behavior

Configured pulls require a backend supporting POST resource exports. A failed
request does not fall back to an unfiltered export. Moving a list file or reordering
its entries does not invalidate matching cached exports, but changing the selected
values can.
Offline pulls restore matching server files; they cannot create a new selection
from an old full-catalog export. Validation failures leave all destinations unchanged.
See [download safety and recovery](downloads.md) for write-failure handling.

## Local supplement and credentials

`sentiary.local.toml` is read beside the selected shared config, even when that config has another filename. It accepts one optional field, `user_api_key` (string, default empty), for example `user_api_key = "YOUR_USER_API_KEY"`. Copy the [local credential template](../examples/local-credentials.toml) to that filename and replace its placeholder; keep it out of source control. Unknown fields fail. It cannot override projects, outputs, server, or paths.

### Project environment credentials

Download the [multi-project credential example](../examples/project-credentials.toml).
Each project can choose a literal string or an environment reference:

```toml
# Literal value (keep real secrets out of committed files):
api_key = "the-actual-api-key"
```

```toml
# Environment reference instead of the literal value:
api_key = { env = "SENTIARY_MOBILE_API_KEY" }
```

Strings are always literal: `api_key = "${VARIABLE}"` does not interpolate.
Only a reference table triggers lookup. It must contain exactly one field, `env`,
whose value is a literal variable name, not general TOML interpolation. The name
must be nonempty and match `[A-Za-z_][A-Za-z0-9_]*` without surrounding whitespace.
The referenced value is trimmed; a missing, empty, or whitespace-only value fails
online without fallback.
Unknown table fields, missing/non-string `env`, other `api_key` types, and duplicate
`api_key` definitions are structural errors.

After validating shared configuration, online credentials are selected as follows:

1. Project TOML `api_key`: a nonblank literal or the exact explicit `env` reference.
2. Nonblank `user_api_key` from the local supplement.
3. Nonblank `SENTIARY_PROJECT_API_KEY`, only with exactly one declared project.
4. Nonblank `SENTIARY_USER_API_KEY`.

Lower-priority sources never conflict with or override the selected source.
Whitespace-only literals and fallback values are absent. Multi-project configurations
ignore `SENTIARY_PROJECT_API_KEY`, even when `--output` selects one project.
Online loading always parses and validates an existing local supplement, even
when its credential will not be used.

An invalid selected key fails without trying alternative sources. All declared
projects require usable credentials online, including those with no selected
output. Missing credentials fail before export requests or destination replacement.
The credential and server determine permissions, not the environment variable's
name. Keys supplied through a user-named variable do not acquire user permissions.
Standalone API commands use separate [credential rules](../cli/reference.md#invocation-credentials-and-streams)
and do not read TOML credentials. `SENTIARY_PROJECT_ID` does not override TOML project IDs.

Offline loading never evaluates credential environment variables and ignores the
local supplement entirely, including malformed TOML. Referenced variables need
not exist. Shared structure, including valid variable names and the strict
`api_key` value/table shape, still validates. Operational variables such as `SENTIARY_BASE_URL` and
`SENTIARY_CACHE_DIR` still apply. Neither keys nor their source names enter cache
identity; key rotation or changing sources does not invalidate cached exports.
Resolved keys are never written back to TOML, preview output, or cache metadata.
Ignore `sentiary.local.toml` and `.sentiary/` in version control. Shared `api_key`
grants the project's permissions; it is not a download-only key.

## Validation failures

Language IDs must be nonblank and unique. They cannot be `.`, `..`, or contain slash, backslash, braces, or NUL. Unknown path expressions and NUL paths fail. Missing locale entries fail when planning the resolved targets. Duplicate output selections, unknown output names, and combined catalogs whose source is not selected fail.

Destinations must be distinct ignoring case, and no destination may be another destination's parent directory. Case-only renames of owned paths fail. Config files, local credentials, `.sentiary/`, and symbolic links in destination paths are protected. A preview validates the download plan, but ownership and filesystem protections are checked during the actual write workflow. See [failure recovery](downloads.md#failure-recovery).
