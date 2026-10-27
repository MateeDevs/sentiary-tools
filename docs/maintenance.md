# Build and maintain the documentation

Run these commands from the standalone tools repository, or first `cd tools` in the main Sentiary checkout. Python 3.11 or newer is required. Use a virtual environment; the documentation dependency is pinned to [Zensical 0.0.72](https://pypi.org/project/zensical/0.0.72/).

```sh
python3 -m venv .venv-docs
. .venv-docs/bin/activate
python -m pip install -r requirements-docs.txt
zensical serve
# In a separate invocation, build static output:
zensical build --strict
```

On Windows activate with `.venv-docs\Scripts\Activate.ps1`. The config is `zensical.toml`, sources are `docs/`, and output is `site/`. Preview defaults to localhost port 8000. See Zensical's [preview](https://zensical.org/docs/usage/preview/) and [strict build](https://zensical.org/docs/usage/build/) reference. Broken internal links and anchors are enabled in configuration; strict mode fails on warnings. The staging site uses `https://docs.staging.sentiary.coolify.matee.cz/` as its canonical URL.

The documentation workflow runs a strict build and uploads `sentiary-docs` as a static artifact. It has read-only repository permissions and no deploy job. Download and serve the artifact with an HTTP server to inspect it; search requires a served site.

## Source inventory and coverage

| Source | Site coverage / maintenance action |
| --- | --- |
| `internal/frontend/cli/app.go`, `management.go` | [Every command and flag](cli/reference.md); include helper-registered flags |
| `internal/frontend/cli/init.go`, `pull.go` | Wizard and pull behavior; shared option table in CLI reference |
| `cmd/sentiary-cli/main.go`, `internal/buildinfo` | Version forms, handshake, exit codes |
| `internal/sentiary/config.go`, `client.go`, `client_management.go`, `models.go` | Environment fallbacks, legacy format restrictions, required fields, accepted role/filter/order values |
| `internal/configfile/config.go` | [Every TOML field](configuration/reference.md), inheritance, discovery, unknown-field rejection |
| `internal/download/plan.go` | Format inventory, language selection, alias and path validation |
| `internal/download/snapshot.go`, `internal/exportcache/cache.go` | Completed sets, revisions, cache identity, offline and failure semantics |
| `internal/fileoutput/store.go` | Ownership, cleanup, locks, adoption, partial-write guarantees |
| `install.sh`, `install.ps1`, `.github/workflows/release.yml`, `go.mod` | [Installation](getting-started/installation.md), targets, checksum manifest and toolchain |
| `internal/frontend/cli/integration.go`, preview/result structs in `pull.go` | [Protocol](integration/protocol.md), downloadable schema bundle |
| Adjacent `sentiary-gradle-plugin` build, extension, tasks, source and tests | [Unreleased Gradle status](gradle/index.md) and [migration](migration.md) |

When commands change, enumerate command registrations and run every command with `--help` without credentials. Compare flags, types, defaults, required validation, environment fallbacks, repeatability, and examples. Help alone does not expose validations or API-only restrictions. Check the source as well. Never run mutating command examples against an account just to check syntax.

Validate the canonical TOML using the existing `configfile.Load`/`LoadOffline` facilities and `download.Plan` with a local language fixture for wildcard and alias examples. The existing config-loader tests cover the mobile example at `examples/sentiary.toml`. Validate the local credential template by copying it beside a temporary shared config as `sentiary.local.toml` and loading without a user-key override. Use placeholder keys or offline loading; never print real credentials. Keep the downloadable files authoritative.

Preview the home page, CLI tables, TOML reference, mobile examples, and protocol page at desktop and narrow widths. Check navigation, search, tables, code scrolling, and code-copy behavior. A strict build checks links and rendering, not the correctness of remote API behavior or app builds.

The [Gradle guide](gradle/index.md) now records the actual settings DSL, overrides, local TestKit results,
and version matrix. Recheck those tests when the plugin or CLI changes. Keep the unreleased warning
until compatible CLI/plugin assets are published, and keep backend rollout status separate from local tests.

## Staging hosting

Coolify Cloud hosts the **Documentation** application in Sentiary's staging environment.
Use repository `MateeDevs/sentiary-tools`, branch `staging`, base directory `/`,
Dockerfile `/Dockerfile`, build target `docs`, exposed port `80`, and health check `/`.
Use the existing staging server and destination. Deploy through the Coolify CLI and verify
HTTPS, navigation, search, assets, and reference pages after the build finishes.
The final Dockerfile target remains the MCP runtime, so builds without a target preserve
existing MCP behavior. The staging CLI workflow only uploads build artifacts; it does not
create releases. Gradle plugin 2.0 remains unreleased.
