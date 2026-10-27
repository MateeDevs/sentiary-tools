# Sentiary Tools

This repository contains the Sentiary CLI, hosted MCP service, and Gradle settings plugin.

- `sentiary-cli` provides direct terminal commands.
- `sentiary-mcp` provides a hosted streamable HTTP MCP service.
- `plugins/sentiary/` provides the `com.sentiary.gradle` plugin and delegates downloads to the CLI.

The CLI and MCP frontends use the API contract and client in `go/internal/sentiary`.

## CLI documentation

See the [documentation site sources](docs/content/index.md) for [installation](docs/content/getting-started/installation.md), the [complete command reference](docs/content/cli/reference.md), [configured downloads](docs/content/configuration/reference.md), and [mobile downloads](docs/content/cli/mobile.md).

Preview and build instructions are in [documentation maintenance](contributing/maintenance.md). The site uses Zensical 0.0.72 and builds a static CI artifact. [Gradle plugin 2.0 documentation](docs/content/gradle/index.md) covers the current settings plugin.

The CLI sends account keys in the `Authorization: Bearer <key>` header. Project operations
require membership and role permissions; project keys cannot manage projects or languages.
Older API deployments accept account keys only on batch and string routes. Deploy the
account-key backend fix and its Firestore index before using management commands.

Use language IDs from `list-languages`, such as `en-US`, `cs-CZ`, and `sk-SK`.
Add each language to the project before import and use its exact ID. An absent project
language returns 404; underscore forms such as `cs_CZ` are not interchangeable.


## Hosted MCP

The hosted frontend reads the user API key when it creates an MCP session. Each session gets a separate API client.

The hosted frontend reads the key from the first available source:

1. `X-Sentiary-User-Api-Key: <key>`
2. `Authorization: Bearer <key>`
3. `SENTIARY_USER_API_KEY` in the service environment

The `Authorization` header can also contain a bare API key.

Example MCP client configuration:

```json
{
  "mcpServers": {
    "sentiary": {
      "type": "http",
      "url": "https://mcp.example.com/",
      "headers": {
        "Authorization": "Bearer user-api-key"
      }
    }
  }
}
```

The hosted frontend uses port `8080` by default. Set `PORT` to use a different port.

The service exposes MCP at `/` and its health check at `/healthz`.

## Build

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

## Docker

Build and run the hosted frontend:

```sh
docker build -f go/Dockerfile -t sentiary-mcp .
docker run --rm -p 8080:8080 sentiary-mcp
```

Clients can send their own API keys. The container does not need a shared API key.

## MCP tools

Project tools:

- `list_projects`
- `get_project`
- `create_project`
- `edit_project`
- `remove_project`

Language tools:

- `list_supported_languages`
- `list_project_languages`
- `add_project_language`
- `remove_project_language`

Member tools:

- `list_project_members`
- `set_project_member_role`
- `remove_project_member`

Invitation tools:

- `list_invitations`
- `create_invitation`
- `accept_invitation`
- `decline_invitation`
- `list_project_invitations`
- `remove_project_invitation`

String tools:

- `list_project_strings`
- `search_project_strings`
- `get_project_string`
- `get_project_string_by_key`
- `add_project_string`
- `edit_project_string`
- `remove_project_string`
- `set_project_string_translation`
- `remove_project_string_translation`

Batch tools:

- `get_project_strings_info`
- `export_project_strings`
- `import_project_strings`

The batch tools support `json`, `android`, `apple`, and `compose` formats.

## Agent skill

The repository includes two agent skills:

- [`sentiary-cli`](https://github.com/MateeDevs/sentiary-mcp-server/tree/main/.agents/skills/sentiary-cli) manages Sentiary through the CLI.
- [`sentiary-mcp`](https://github.com/MateeDevs/sentiary-mcp-server/tree/main/.agents/skills/sentiary-mcp) manages Sentiary through MCP.
