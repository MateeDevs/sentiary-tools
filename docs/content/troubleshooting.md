---
tags:
  - CLI
  - Gradle
---

# Troubleshooting

| Symptom | What to check |
| --- | --- |
| `command not found` | Installation directory and shell PATH; inspect resolved executable |
| Unknown `pull` or `init` command | Install CLI 2.0; inspect version and command help |
| No `sentiary.toml` found | Run inside its directory tree or pass `--config` |
| Invalid TOML or unsupported fields | Exact field names, types, and `config_version = 1`; local supplement allows only `user_api_key` |
| Unexpected credential used | User environment key wins over local user key, then project key; invalid keys do not fall back |
| Legacy command ignores local key | Legacy API commands read `SENTIARY_USER_API_KEY`, not TOML |
| Wrong language folder | Locale mapping changes paths only; use exact remote language IDs and explicit consumer-compatible qualifiers |
| Alias source missing | Every chain must end in a project language; aliases copy whole files |
| Combined catalog rejected | Use `xcstrings`, fixed path, selected source, no locale map/aliases/placeholders |
| Binary export needed | Use configured `pull`; legacy `export` accepts only four text-format tokens |
| Existing destination rejected | Review file and use deliberate adoption with `--force` |
| Pull appears to wait | Another process may hold the per-config or cache lock; lock-file presence is not proof |
| Offline cache miss | Cache location, selected requests, saved language expansion, intact single completed set |
| Online authorization/server failure | Resolve online failure; there is no automatic offline fallback |
| Files missing after clean | Restore explicitly using `pull --offline` and the same completed cache set |

For path conflicts, corrupt metadata, partial writes, and stale sets, follow [download failure recovery](configuration/downloads.md#failure-recovery). For syntax, types, defaults, and required options, see [CLI reference](cli/reference.md).
