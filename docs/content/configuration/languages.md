---
tags:
  - Configuration
---

# Generic exports and multiple projects/modules

```sh
sentiary-cli pull --output module_a --output module_b --dry-run
sentiary-cli pull --output module_a --output module_b
```

Download [advanced TOML](../examples/advanced.toml) and save it as `sentiary.toml`. Replace both project IDs and supply a user key. It contains no credentials.

## Wildcards and locale mappings

The `app` project uses `languages = "*"`. `module_a` overrides that with explicit target languages, so it generates only those files. Its locale map chooses resource folder names independently of API IDs. Unknown language IDs fail at the service or alias resolver; they are not automatically converted.

`module_b` inherits wildcard selection and uses `{language}` in generic JSON filenames. This avoids having to enumerate a locale map as project languages change. The wildcard records the server's language list in the completed set; offline mode reuses that list.

## Chained aliases

`module_a` maps `en-AU` to `en-GB`, then `en-GB` to `en`. Both target files contain the whole English export. The project must contain `en`, and all configured alias chains must resolve even if a target is not selected. No per-key fallback occurs.

```sh
sentiary-cli pull --output module_a --dry-run
sentiary-cli pull --output module_a
```

## Multiple Gradle modules

`module_a` writes under `feature-a/build/generated/sentiary`; register that resource root in module `feature-a` using the Android setup in the [mobile guide](../cli/mobile.md#android). `module_b` writes JSON under `feature-b/build/generated/locales`; configure that module's consumer to load those files as its application requires. A `.json` extension does not make it an Android or Compose resource.

Run one CLI command before both module builds:

```sh
sentiary-cli pull --output module_a --output module_b
```

This is manual orchestration. The Gradle 2.0 settings plugin can replace manual orchestration for [Android/Compose bindings](../gradle/index.md#setup). Keep output paths distinct across modules and avoid two configs managing the same destinations.

## Multiple remote projects and generic formats

The `website` output uses a second project and explicitly selects one language. A single user key must have access to both configured projects. Alternatively, remove user-key sources and add appropriate project keys to each project table; do not publish real keys.

`catalog` overrides both source language and language selection for a combined Xcode catalog. It also enables empty translations. `sheet` exports binary XLSX bytes to language-specific files; use `pull` rather than legacy `export` for that format.

```sh
sentiary-cli pull --output website --output catalog --output sheet
```

Review the [configuration rules](../configuration/reference.md) before adding locale maps, combined layouts, or new formats.
