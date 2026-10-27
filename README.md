# Sentiary Tools

This project provides two frontends for the Sentiary REST API.

- `sentiary-cli` provides direct terminal commands.
- `sentiary-mcp` provides a hosted streamable HTTP MCP service.

Both frontends use the API contract and client in `internal/sentiary`.

## CLI documentation

See the [documentation site sources](docs/index.md) for [installation](docs/getting-started/installation.md), the [complete command reference](docs/cli/reference.md), [configured downloads](docs/configuration/reference.md), and [examples](docs/examples/index.md).

Preview and build instructions are in [documentation maintenance](docs/maintenance.md). The site uses Zensical 0.0.72 and builds a static CI artifact. [Gradle plugin 2.0 documentation](docs/gradle/index.md) remains marked unreleased.

## Hosted MCP

The hosted frontend reads the user API key when it creates an MCP session. Each session gets a separate API client.

The hosted frontend reads the key from the first available source:

1. `X-Sentiary-User-Api-Key: <key>`
2. `Authorization: Bearer <key>`
3. `Authorization: Ribbon <key>`
4. `SENTIARY_USER_API_KEY` in the service environment

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
go build -o bin/sentiary-cli ./cmd/sentiary-cli
go build -o bin/sentiary-mcp ./cmd/sentiary-mcp
```

## Docker

Build and run the hosted frontend:

```sh
docker build -t sentiary-mcp .
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
