---
tags:
  - CLI
  - Gradle
---

# Troubleshooting

| Symptom | What to check |
| --- | --- |
| `command not found` | Installation directory and shell PATH; inspect resolved executable |
| Unknown `pull` or `init` command | Install the current CLI; inspect the resolved executable, version, and command help |
| No `sentiary.toml` found | Run inside its directory tree or pass `--config` |
| Invalid TOML or unsupported fields | Exact field names, types, and `config_version = 1`; local supplement allows only `user_api_key` |
| Unexpected credential used | Check [credential priority](configuration/reference.md#project-environment-credentials); an invalid selected key fails without fallback |
| Standalone API command ignores local key | API commands read `SENTIARY_USER_API_KEY`, not TOML |
| Wrong language folder | Locale mapping changes paths only; use exact remote language IDs and explicit consumer-compatible qualifiers |
| Alias source missing | Every chain must end in a project language; aliases copy whole files |
| Combined catalog rejected | Use `xcstrings`, fixed path, selected source, no locale map/aliases/placeholders |
| Binary export needed | Use configured `pull`; standalone `export` accepts only `json`, `android`, `apple`, and `compose` |
| Existing destination rejected | Review file and use deliberate adoption with `--force` |
| Pull appears to wait | Another process may hold the per-config or cache lock; lock-file presence is not proof |
| Offline cache miss | Cache location, selected requests, saved language expansion, intact single completed set |
| Online authorization/server failure | Resolve online failure; there is no automatic offline fallback |
| Files missing after clean | Restore explicitly using `pull --offline` and the same completed cache set |

For path conflicts, corrupt metadata, partial writes, and stale sets, follow [download failure recovery](configuration/downloads.md#failure-recovery). For syntax, types, defaults, and required options, see [CLI reference](cli/reference.md).

## Compose and Isolated Projects

With Compose Multiplatform 1.10.3 and Gradle Isolated Projects, set these framework
properties in `gradle.properties`:

```properties
org.jetbrains.compose.hot.reload.disable=true
kotlin.internal.collectFUSMetrics=false
```

The first disables incompatible hot-reload integration. The second avoids a Kotlin
statistics build-service classloader conflict. These are framework properties,
not Sentiary extension fields. See the
[Gradle properties example](examples/gradle-evaluation.properties).
