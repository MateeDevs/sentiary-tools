# Downloads, ownership, and cache

## Ownership and adoption

Each config uses `.sentiary/<config-filename>.json` to record output names, absolute file paths, and SHA-256 checksums. A missing generated file can be restored. An existing untracked destination or a manually changed owned file blocks overwrite or removal, even when its content happens to match the new export.

Review existing files and the preview, then use `sentiary-cli pull --force` to adopt destinations or discard edits. `--force` does not bypass invalid paths, symlink protection, case-only rename restrictions, or ownership of an unselected output. Never expect manual Xcode catalog edits to merge into server translations.

After a successful pull, obsolete owned files for selected outputs are removed. Unrelated files and unselected outputs remain. A full pull also cleans ownership for removed output definitions; a selected pull limits cleanup to those names. Empty directories may remain.

## Concurrency and write guarantees

A kernel-backed per-config lock at `.sentiary/<config-filename>.lock` covers fetching, cache publication, and destination writes. Concurrent pulls sharing the same manifest wait until the lock is available; cancellation stops the wait. A process exit releases the lock. The lock file may remain: its presence alone does not indicate an active lock. Do not delete a lock file while another pull may hold it.

Different configs sharing destinations are not coordinated. Avoid overlapping ownership. The cache has a separate process lock around reads, publication, and cleanup, released during network requests. Online and offline previews do not acquire the ownership lock.

All requested exports must succeed before any generated destination changes. Files are staged beside their destinations and renamed individually. The manifest is replaced after file writes and obsolete-file cleanup. This is not a multi-file filesystem transaction: an I/O failure can leave some files updated.

## Conditional requests and completed sets

Online pulls authenticate and request backend `export/state` protocol 1, then send stored ETags via `If-None-Match`. A valid `304` reuses verified bytes. The response generation, data revision, and exporter version must match the observed project state. Each changing project is retried up to three times, then fails. A server/access/download error never silently selects offline data.

A completed set is published only after all requested exports are verified and saved as immutable SHA-256 objects. Publication precedes destination validation/writes. A destination conflict can therefore leave a usable completed set. Snapshot consistency is per project; it is not a cross-project transaction or a translation lockfile. A committed intermediate state of a bulk edit may be captured.

## Cache location and identity

Precedence: `--cache-dir`, then `SENTIARY_CACHE_DIR`, then `os.UserCacheDir()/sentiary/exports-v1`. Relative overrides are relative to the process working directory.

| OS | Usual default base |
| --- | --- |
| macOS | `$HOME/Library/Caches/sentiary/exports-v1` |
| Linux | `$XDG_CACHE_HOME/sentiary/exports-v1`, otherwise `$HOME/.cache/sentiary/exports-v1` |
| Windows | `%LocalAppData%\sentiary\exports-v1` |

The cache adds a hash of the normalized server URL, then stores `objects/` and `sets/` with a `cache.lock`. Identity includes project ID, format, layout, language selection, combined source language, and empty-translation policy. Keys, output names, checkout paths, and destination paths are excluded. Renaming outputs, changing keys, or moving checkouts can reuse exports.

There is one latest completed set per exact export request selection. Replacement removes only objects unreferenced by remaining selections; cleanup is best effort. Sets do not expire automatically. Protect cache contents like generated translations.

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

A cache publication failure leaves destinations unchanged and retains the previously published set where possible. A later destination error leaves the completed set available for recovery. See the [recovery walkthrough](../examples/recovery.md).
