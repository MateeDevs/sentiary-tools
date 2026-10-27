# Build and maintain the documentation

Run these commands from the standalone tools repository, or first `cd tools` in the main Sentiary checkout. Python 3.11 or newer is required. Use a virtual environment; the documentation dependency is pinned to [Zensical 0.0.72](https://pypi.org/project/zensical/0.0.72/).

```sh
python3 -m venv .venv-docs
. .venv-docs/bin/activate
python -m pip install -r requirements-docs.txt
zensical serve
# In a separate invocation, build static output:
zensical build --strict
python scripts/check-docs-exports.py
```

On Windows activate with `.venv-docs\Scripts\Activate.ps1`. The config is `zensical.toml`, sources are `docs/`, and output is `site/`. Preview defaults to localhost port 8000. See Zensical's [preview](https://zensical.org/docs/usage/preview/) and [strict build](https://zensical.org/docs/usage/build/) reference. Broken internal links and anchors are enabled in configuration; strict mode fails on warnings. The staging site uses `https://docs.staging.sentiary.coolify.matee.cz/` as its canonical URL.

The documentation workflow runs a strict build, checks the generated text exports, and uploads `sentiary-docs` as a static artifact. It has read-only repository permissions and no deploy job. Download and serve the artifact with an HTTP server to inspect it; search requires a served site.

## AI documentation exports

The native `llmstxt` plugin generates `llms.txt`, `llms-full.txt`, and Markdown copies
of the public pages explicitly listed in `zensical.toml`. Update that selection when adding
user guides. Keep topic indexes, redirects, repository documentation, and Scalar HTML out
of the selection. Links follow `site_url`; do not set a separate `base_url`.
The API guide links to the existing `/api/openapi.yaml` proxy without copying the schema.

The home page links to both text exports. Selected pages offer **Copy as Markdown**.
Use explicit HTML links for pages excluded from the Markdown selection (for example,
`topics/index.html`); Zensical otherwise rewrites directory links to missing Markdown exports.
The checker verifies page selection and order, local link targets, full-output coverage,
copy actions, and preservation of code blocks, tables, and collapsed content.
The CI artifact and `docs` Docker target include all generated files in `site/`.
nginx serves `.txt` as `text/plain` and `.md` as `text/markdown`, both with UTF-8.

## Source inventory and coverage

| Source | Site coverage / maintenance action |
| --- | --- |
| `internal/frontend/cli/app.go`, `management.go` | [Every command and flag](https://github.com/MateeDevs/sentiary-tools/blob/staging/docs/cli/reference.md); include helper-registered flags |
| `internal/frontend/cli/init.go`, `pull.go` | Wizard and pull behavior; shared option table in CLI reference |
| `cmd/sentiary-cli/main.go`, `internal/buildinfo` | Version forms, handshake, exit codes |
| `internal/sentiary/config.go`, `client.go`, `client_management.go`, `models.go` | Environment fallbacks, legacy format restrictions, required fields, accepted role/filter/order values |
| `internal/configfile/config.go` | [Every TOML field](https://github.com/MateeDevs/sentiary-tools/blob/staging/docs/configuration/reference.md), inheritance, discovery, unknown-field rejection |
| `internal/download/plan.go` | Format inventory, language selection, alias and path validation |
| `internal/download/snapshot.go`, `internal/exportcache/cache.go` | Completed sets, revisions, cache identity, offline and failure semantics |
| `internal/fileoutput/store.go` | Ownership, cleanup, locks, adoption, partial-write guarantees |
| `install.sh`, `install.ps1`, `.github/workflows/release.yml`, `go.mod` | [Installation](https://github.com/MateeDevs/sentiary-tools/blob/staging/docs/getting-started/installation.md), targets, checksum manifest and toolchain |
| `internal/frontend/cli/integration.go`, preview/result structs in `pull.go` | [Protocol](https://github.com/MateeDevs/sentiary-tools/blob/staging/repository-docs/integration-protocol.md), downloadable schema bundle |
| `sentiary/` build, extension, tasks, source and tests | [Gradle reference](https://github.com/MateeDevs/sentiary-tools/blob/staging/docs/gradle/index.md) and [migration](https://github.com/MateeDevs/sentiary-tools/blob/staging/docs/migration.md) |

When commands change, enumerate command registrations and run every command with `--help` without credentials. Compare flags, types, defaults, required validation, environment fallbacks, repeatability, and examples. Help alone does not expose validations or API-only restrictions. Check the source as well. Never run mutating command examples against an account just to check syntax.

Validate the canonical TOML using the existing `configfile.Load`/`LoadOffline` facilities and `download.Plan` with a local language fixture for wildcard and alias examples. The existing config-loader tests cover the mobile example at `examples/sentiary.toml`. Validate the local credential template by copying it beside a temporary shared config as `sentiary.local.toml` and loading without a user-key override. Use placeholder keys or offline loading; never print real credentials. Keep the downloadable files authoritative.

Preview the home page, CLI tables, TOML reference, mobile examples, and protocol page at desktop and narrow widths. Check navigation, search, tables, code scrolling, and code-copy behavior. A strict build checks links and rendering, not the correctness of remote API behavior or app builds.

The [Gradle guide](../docs/gradle/index.md) records the settings DSL, overrides, and supported versions. Recheck the `sentiary/` plugin tests when the plugin or CLI changes. Keep release procedures and validation evidence in repository documentation, outside `docs/`.

## Staging hosting

Coolify Cloud hosts the **Documentation** application in Sentiary's staging environment.
Use repository `MateeDevs/sentiary-tools`, branch `staging`, base directory `/`,
Dockerfile `/Dockerfile`, build target `docs`, exposed port `80`, and health check `/`.
Use the existing staging server and destination. Deploy through the Coolify CLI and verify
HTTPS, navigation, search, assets, and reference pages after the build finishes.
The final Dockerfile target remains the MCP runtime, so builds without a target preserve
existing MCP behavior. The staging CLI workflow only uploads build artifacts; it does not
create releases.

### API reference

The [API section](https://github.com/MateeDevs/sentiary-tools/blob/staging/docs/api/index.md) links to Scalar API Reference, using the pinned `@scalar/api-reference` 1.76.0 standalone browser bundle.
The CDN script has a SHA-384 integrity hash; update the version and hash together when upgrading.
`hideTestRequestButton` and `hideClientButton` keep the reference read-only. Developer tools and external font loading are disabled.
The reference loads `/api/openapi.yaml` from the docs domain. `docs-nginx.conf`
proxies only that exact path to the staging backend's `/swagger/documentation.yaml`, verifies
upstream TLS, and strips credentials and cookies. The backend resource
`backend/app/src/main/resources/openapi/documentation.yaml` remains the authoritative schema;
the docs repository keeps no duplicate. A backend deployment updates the reference automatically.

`zensical serve` previews the static reference page but does not run this nginx proxy. Preview
the schema integration through the `docs` Docker target or the deployed docs application.
Verify the proxied schema matches the backend schema and the reference assets load after deployment.

## Build from source

Use the Go version required by `go.mod` (currently `go 1.25.5`) or a compatible newer toolchain:

```sh
git clone https://github.com/MateeDevs/sentiary-tools.git
cd sentiary-tools
go build -o bin/sentiary-cli ./cmd/sentiary-cli
./bin/sentiary-cli version --json
# Or install to GOBIN, or GOPATH/bin when GOBIN is unset:
go install ./cmd/sentiary-cli
```

On Windows the built executable needs a `.exe` suffix. Add `GOBIN` or the Go workspace's `bin` directory to PATH when using `go install`. A local build normally reports `dev`; release builds report their release tag.

## Repository references

- [CLI integration protocol](integration-protocol.md)
- [Download and cache internals](download-internals.md)
- [Gradle release operations](gradle-release.md)
