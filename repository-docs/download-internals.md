# Download and cache internals

User-facing guarantees and recovery steps live in [downloads and offline use](../docs/configuration/downloads.md).

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
