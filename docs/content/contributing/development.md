# Development

Run commands from the `sentiary-tools` repository root. Use the Go version in
[`go/go.mod`](https://github.com/MateeDevs/sentiary-tools/blob/main/go/go.mod) and JDK 21 for the Gradle plugin.
Plugin integration fixtures also need Python 3 and Android SDK platform 36.

## Repository layout

| Directory | Contents |
| --- | --- |
| `go/` | CLI, hosted MCP service, shared API client, and installers |
| `plugins/sentiary/` | Gradle settings plugin, available as `:sentiary` |
| `docs/` | Documentation site configuration and public guides in `content/` |
| `docs/content/contributing/` | Development, maintenance, and release guides |

## Build and check

Build both frontends:

```sh
mkdir -p bin
go -C go build -o ../bin/sentiary-cli ./cmd/sentiary-cli
go -C go build -o ../bin/sentiary-mcp ./cmd/sentiary-mcp
```

Run Go checks and plugin integration fixtures from the repository root:

```sh
go -C go test ./...
go -C go vet ./...
SENTIARY_TEST_CLI="$PWD/bin/sentiary-cli" ./gradlew :sentiary:test :sentiary:ktlintCheck :sentiary:validatePlugins
```

`SENTIARY_TEST_CLI` enables the actual-CLI integration fixture. Build the CLI from the
same checkout before running the plugin checks.

## Further reading

- [CLI and hosted MCP](../mcp/index.md): authentication, server configuration, Docker, and tool reference.
- [Documentation maintenance](documentation.md): preview, strict build, Markdown exports, and hosting.
- [Gradle development and releases](releases.md): integration matrix, consumer builds, and publication.
- [CLI integration protocol](../integration/protocol.md): the Gradle-to-CLI machine interface.
- [Download internals](../configuration/internals.md): caching, ownership, and concurrency.
