# Build and maintain the documentation

Run these commands from the `sentiary-tools` repository root. Python 3.11 or newer is required. Use a virtual environment; the documentation dependency is pinned to [Zensical 0.0.72](https://pypi.org/project/zensical/0.0.72/).

```sh
python3 -m venv .venv-docs
. .venv-docs/bin/activate
python -m pip install -r docs/requirements.txt
zensical serve --config-file docs/zensical.toml
# In a separate invocation, build static output:
python scripts/build-docs.py --environment production
python scripts/build-docs.py --environment staging
```

On Windows activate with `.venv-docs\Scripts\Activate.ps1`. The config is `docs/zensical.toml`, sources are `docs/content/`, and output is `docs/site/`. Zensical requires output to stay inside the directory containing its configuration. Preview defaults to localhost port 8000. See Zensical's [preview](https://zensical.org/docs/usage/preview/) and [strict build](https://zensical.org/docs/usage/build/) reference. Broken internal links and anchors are enabled in configuration; strict mode fails on warnings. Production uses `https://docs.sentiary.com/`; staging uses `https://docs.staging.sentiary.coolify.matee.cz/`. The environment build command selects canonical URLs, source branches, and API schema upstreams together.

The documentation workflow runs a strict build, checks the generated text exports, and uploads `sentiary-docs` as a static artifact. It has read-only repository permissions and no deploy job. Download and serve the artifact with an HTTP server to inspect it; search requires a served site.

## AI documentation exports

The native `llmstxt` plugin generates `llms.txt`, `llms-full.txt`, and Markdown copies
of the public pages explicitly listed in `docs/zensical.toml`. Update that selection when adding
user guides. Keep topic indexes, redirects, contributor operations, and Scalar HTML out
of the selection. Links follow `site_url`; do not set a separate `base_url`.
The API guide links to `/api/openapi.yaml`, which proxies the selected environment without copying the schema.

The home page links to both text exports. Selected pages offer **Copy as Markdown**.
Use explicit HTML links for pages excluded from the Markdown selection (for example,
`topics/index.html`); Zensical otherwise rewrites directory links to missing Markdown exports.
The checker verifies page selection and order, local link targets, full-output coverage,
copy actions, and preservation of code blocks, tables, and collapsed content.
The CI artifact and `docs` Docker target include all generated files in `docs/site/`.
nginx serves `.txt` as `text/plain` and `.md` as `text/markdown`, both with UTF-8.

## Source inventory and coverage

| Source | Site coverage / maintenance action |
| --- | --- |
| `go/internal/frontend/cli/app.go`, `management.go` | [Every command and flag](../cli/reference.md); include helper-registered flags |
| `go/internal/frontend/cli/init.go`, `pull.go` | Wizard and pull behavior; shared option table in CLI reference |
| `go/cmd/sentiary-cli/main.go`, `go/internal/buildinfo` | Version forms, handshake, exit codes |
| `go/internal/sentiary/config.go`, `client.go`, `client_management.go`, `models.go` | Environment fallbacks, legacy format restrictions, required fields, accepted role/filter/order values |
| `go/internal/configfile/config.go` | [Every TOML field](../configuration/reference.md), inheritance, discovery, unknown-field rejection |
| `go/internal/download/plan.go` | Format inventory, language selection, alias and path validation |
| `go/internal/download/snapshot.go`, `go/internal/exportcache/cache.go` | Completed sets, revisions, cache identity, offline and failure semantics |
| `go/internal/fileoutput/store.go` | Ownership, cleanup, locks, adoption, partial-write guarantees |
| `go/install.sh`, `go/install.ps1`, `.github/workflows/release.yml`, `go/go.mod` | [Installation](../getting-started/installation.md), targets, checksum manifest and toolchain |
| `go/internal/frontend/cli/integration.go`, preview/result structs in `pull.go` | [Protocol](../integration/protocol.md), downloadable schema bundle |
| `plugins/sentiary/` build, extension, tasks, source and tests | [Gradle reference](../gradle/index.md) and [migration](../migration.md) |

When commands change, enumerate command registrations and run every command with `--help` without credentials. Compare flags, types, defaults, required validation, environment fallbacks, repeatability, and examples. Help alone does not expose validations or API-only restrictions. Check the source as well. Never run mutating command examples against an account just to check syntax.

Validate the canonical TOML using the existing `configfile.Load`/`LoadOffline` facilities and `download.Plan` with a local language fixture for wildcard and alias examples. The existing config-loader tests cover the mobile example at `go/examples/sentiary.toml`. Validate the local credential template by copying it beside a temporary shared config as `sentiary.local.toml` and loading without a user-key override. Use placeholder keys or offline loading; never print real credentials. Keep the downloadable files authoritative.

Preview the home page, CLI tables, TOML reference, mobile examples, and protocol page at desktop and narrow widths. Check navigation, search, tables, code scrolling, and code-copy behavior. A strict build checks links and rendering, not the correctness of remote API behavior or app builds.

The [Gradle guide](../gradle/index.md) records the settings DSL, overrides, and supported versions. Recheck the `plugins/sentiary/` plugin tests when the plugin or CLI changes. Keep release procedures in the contributor guides; do not commit local validation transcripts.

## Hosting

Coolify Cloud hosts a **Documentation** application in each Sentiary environment.
Use repository `MateeDevs/sentiary-tools`, branch `main` for production or `staging` for staging, and base directory `/`,
Dockerfile `/docs/Dockerfile`, build target `docs`, exposed port `80`, and health check `/`.
Use the existing server and destination for the selected environment. Set the build-time variable `DOCS_ENVIRONMENT` to `production` or `staging`; production is the default. Deploy through the Coolify CLI and verify
HTTPS, navigation, search, assets, and reference pages after the build finishes.
The MCP runtime uses `/go/Dockerfile` with the repository root as its build context.
The staging CLI workflow only uploads build artifacts; it does not create releases.

Build the documentation image from the repository root:

```sh
docker build -f docs/Dockerfile --build-arg DOCS_ENVIRONMENT=production -t sentiary-docs .
```

For MCP container setup, see [self-hosting](../mcp/index.md#self-host-with-docker).

### API reference

The [API section](../api/index.md) links to Scalar API Reference, using the pinned `@scalar/api-reference` 1.76.0 standalone browser bundle.
The CDN script has a SHA-384 integrity hash; update the version and hash together when upgrading.
`hideTestRequestButton` and `hideClientButton` keep the reference read-only. Developer tools and external font loading are disabled.
The reference loads `/api/openapi.yaml` from the docs domain. `docs/nginx.conf`
proxies only that exact path to the selected backend's `/swagger/documentation.yaml`, verifies
upstream TLS, and strips credentials and cookies. The backend resource
`backend/app/src/main/resources/openapi/documentation.yaml` remains the authoritative schema;
the docs repository keeps no duplicate. A backend deployment updates the reference automatically. The schema server URL is `https://api.sentiary.com`; operation paths already include `/api/v1`.

`zensical serve --config-file docs/zensical.toml` previews the static reference page but does not run this nginx proxy. Preview
the schema integration through the `docs` Docker target or the deployed docs application.
Verify the proxied schema matches the backend schema and the reference assets load after deployment.

## Build from source

Use the Go version required by `go/go.mod` (currently `go 1.25.5`) or a compatible newer toolchain:

```sh
git clone https://github.com/MateeDevs/sentiary-tools.git
cd sentiary-tools/go
go build -o ../bin/sentiary-cli ./cmd/sentiary-cli
../bin/sentiary-cli version --json
# Or install to GOBIN, or GOPATH/bin when GOBIN is unset:
go install ./cmd/sentiary-cli
```

On Windows the built executable needs a `.exe` suffix. Add `GOBIN` or the Go workspace's `bin` directory to PATH when using `go install`. A local build normally reports `dev`; release builds report their release tag.

## Repository references

- [CLI integration protocol](../integration/protocol.md)
- [Download and cache internals](../configuration/internals.md)
- [Gradle release operations](releases.md)
