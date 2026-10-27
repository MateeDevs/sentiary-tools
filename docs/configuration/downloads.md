---
tags:
  - Configuration
  - Offline
---

# Downloads and offline use

```sh
sentiary-cli pull
sentiary-cli pull --offline
```

Run online first to save a complete set of exports. Offline mode restores that set without credentials or a network connection.

## Ownership and adoption

Each config uses `.sentiary/<config-filename>.json` to record output names, absolute file paths, and SHA-256 checksums. A missing generated file can be restored. An existing untracked destination or a manually changed owned file blocks overwrite or removal, even when its content happens to match the new export.

!!! warning "Replacing existing files"

    Review existing files and the preview, then use `sentiary-cli pull --force` to adopt destinations or discard edits.
 `--force` does not bypass invalid paths, symlink protection, case-only rename restrictions, or ownership of an unselected output. Never expect manual Xcode catalog edits to merge into server translations.

After a successful pull, obsolete owned files for selected outputs are removed. Unrelated files and unselected outputs remain. A full pull also cleans ownership for removed output definitions; a selected pull limits cleanup to those names. Empty directories may remain.

## Concurrent pulls and write failures

Pulls for the same configuration wait for one another. Cancellation stops the wait. A leftover lock file does not prove a pull is running; do not delete it while another pull may be active. Different configs must not manage the same destinations.

All requested exports must succeed before any generated file changes. Individual file replacement is atomic, but the whole pull is not a multi-file transaction. An I/O failure can leave some files updated; inspect affected files before retrying with `--force`.

Online pulls authenticate and check for updates. A server, access, or download error never silently selects offline data. A completed export set may remain available even if a later destination write fails. Consistency is per project; a pull is not a cross-project transaction or a translation lockfile.

## Cache location and identity

Precedence: `--cache-dir`, then `SENTIARY_CACHE_DIR`, then the OS cache directory shown below. Relative overrides are relative to the process working directory.

| OS | Usual default base |
| --- | --- |
| macOS | `$HOME/Library/Caches/sentiary/exports-v1` |
| Linux | `$XDG_CACHE_HOME/sentiary/exports-v1`, otherwise `$HOME/.cache/sentiary/exports-v1` |
| Windows | `%LocalAppData%\sentiary\exports-v1` |

Cached exports are scoped to the server, project, format, layout, languages, source language, and empty-translation policy. Moving a checkout or renaming outputs can reuse exports. Sets do not expire automatically. Protect cached translation content like generated files; it contains no API keys.

## Offline restoration

`pull --offline` makes no network calls and needs no credentials. It finds the newest single completed set covering the requested exports, including a selected subset. It does not combine several sets. Wildcard expansion uses that set's saved project languages. Every selected object's checksum must verify before writes. Completion time and age are printed to stderr.

Deleting `build/` does not invalidate a cache set. Restore it with the same config and cache location. Changing format, selected languages, source language, or empty-value policy may require another online pull. Offline mode does not establish current access or freshness.

## Failure recovery

| Failure | Recovery |
| --- | --- |
| Untracked or edited destination | Review/back up files; deliberately adopt with `--force` |
| Case-only owned rename | Move through a distinct destination in a successful pull |
| Unsupported export/state protocol | Use a compatible backend; legacy commands remain separate |
| Project changed during three attempts | Retry after edits settle |
| No completed set covers request | Run an online pull for that selection and cache location |
| Missing/corrupt object | Run online to fetch valid content; do not edit hashes |
| Corrupt cache metadata | Preserve for diagnosis, choose a fresh cache directory, then pull online |
| Partial destination write or manifest save | Inspect the affected files; retry with `--force` only after accepting replacements |
| Invalid ownership manifest | Preserve it and inspect previous owned paths before rebuilding ownership; unrecorded obsolete files cannot be cleaned automatically |

A cache publication failure leaves destinations unchanged and retains the previously published set where possible. A later destination error leaves the completed set available for recovery. See the recovery steps below.

## CI caching

Choose a cache path explicitly for portable CI. Cache the CLI's exports directory separately from generated `build/` directories. Do not treat a restored cache as proof of current authorization or server freshness; an online pull still authenticates and checks revisions.

A GitHub Actions sequence, after installing the CLI and checking out your application's config:

```yaml
- uses: actions/cache@v4
  with:
    path: .ci-cache/sentiary
    key: sentiary-${{ runner.os }}-${{ hashFiles('sentiary.toml') }}-${{ github.run_id }}
    restore-keys: |
      sentiary-${{ runner.os }}-${{ hashFiles('sentiary.toml') }}-
      sentiary-${{ runner.os }}-
- name: Download translations
  env:
    SENTIARY_USER_API_KEY: ${{ secrets.SENTIARY_USER_API_KEY }}
    SENTIARY_CACHE_DIR: .ci-cache/sentiary
  run: sentiary-cli pull
```

The run-specific key permits saving refreshed sets; restore prefixes reuse previous content. Restrict access to cached translations as appropriate for your project. Do not include credentials or `sentiary.local.toml` in cache paths. Ignore `.ci-cache/` locally if using this convention. Pin the CLI release separately.

## Restore deleted generated files

With the mobile config, prime Android and Compose together:

```sh
sentiary-cli pull --cache-dir .ci-cache/sentiary --output android --output compose
```

Delete only the generated Android and Compose directories using your normal clean command or file manager. Keep `.sentiary/` and `.ci-cache/sentiary/`. Then restore:

```sh
sentiary-cli pull --offline --cache-dir .ci-cache/sentiary --output android --output compose
```

The command restores missing files without a key or network. Verify the resulting paths with your app build. It prints the completed set's timestamp and age to stderr. A cache primed only for Android cannot restore Compose. Separate completed sets are not combined; prime the joint selection online first.

For a planned disconnected CI build, use the same explicit offline invocation after restoring a trusted complete cache. Do not automatically retry a failed online pull offline: doing so can conceal revoked access or service errors. The 2.0 settings plugin forwards Gradle's `--offline` to the CLI. Standalone CLI invocations still need the explicit `--offline` flag.
