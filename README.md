# Sentiary Tools

This project provides two frontends for the Sentiary REST API.

- `sentiary-cli` provides direct terminal commands.
- `sentiary-mcp` provides a hosted streamable HTTP MCP service.

Both frontends use the API contract and client in `internal/sentiary`.

## Install the CLI

Install the latest release:

```sh
curl -fsSL https://raw.githubusercontent.com/MateeDevs/sentiary-mcp-server/main/install.sh | sh
```

On Windows, run this command in PowerShell:

```powershell
irm https://raw.githubusercontent.com/MateeDevs/sentiary-mcp-server/main/install.ps1 | iex
```

The installers support AMD64 and ARM64 on macOS, Linux, and Windows.

You can also install the CLI from source:

```sh
git clone https://github.com/MateeDevs/sentiary-mcp-server.git
cd sentiary-mcp-server
go install ./cmd/sentiary-cli
```

Set the user API key before you run a command:

```sh
export SENTIARY_USER_API_KEY="user-api-key"
```

You can set `SENTIARY_PROJECT_ID` as the default project ID. A command-level `--project-id` value takes precedence.

The CLI sends the key in the `Authorization: Ribbon <key>` header.
User keys identify an account. Project operations still require membership and the required role permissions.
Project keys cannot manage projects or languages.

Older API deployments accept user keys only on batch and string routes.
A batch request can succeed while project or language commands return 401.
Deploy the account-key backend fix and its Firestore index before you use these commands.
No CLI authentication update is required.

## CLI commands

Export Czech strings as JSON to standard output:

```sh
sentiary-cli export \
  --project-id="asdf" \
  --format="json" \
  --language="cs-CZ"
```

Write the export to a file:

```sh
sentiary-cli export \
  --project-id="asdf" \
  --format="json" \
  --language="cs-CZ" \
  --output="strings.json"
```

Use language IDs from `list-languages`, such as `en-US`, `cs-CZ`, and `sk-SK`.
Add each language with `add-project-language` before an import.
The import language must match a project language exactly. An absent project language returns 404.
Do not substitute `en` for `en-US` or use underscore forms such as `cs_CZ`.

Import a file:

```sh
sentiary-cli import \
  --project-id="asdf" \
  --format="json" \
  --language="cs-CZ" \
  --input="strings.json"
```

Use `-` as the input or output path for a standard stream.

Available commands:

- Projects: `list-projects`, `get-project`, `create-project`, `edit-project`, `remove-project`
- Languages: `list-languages`, `list-project-languages`, `add-project-language`, `remove-project-language`
- Members: `list-project-members`, `set-project-member-role`, `remove-project-member`
- Invitations: `list-invitations`, `create-invitation`, `accept-invitation`, `decline-invitation`, `list-project-invitations`, `remove-project-invitation`
- `list`
- `search`
- `get`
- `add`
- `edit`
- `remove`
- `set-translation`
- `remove-translation`
- `info`
- `export`
- `import`

Run `sentiary-cli <command> --help` to see the command options.

Create a project and add Czech:

```sh
sentiary-cli create-project --name="Mobile app"
sentiary-cli add-project-language \
  --project-id="asdf" \
  --language="cs-CZ"
```

Invite a developer:

```sh
sentiary-cli create-invitation \
  --project-id="asdf" \
  --user-email="developer@example.com" \
  --role="developer"
```

Member and invitation roles are `administrator`, `developer`, and `translator`.

Search keys, translations, descriptions, and context:

```sh
sentiary-cli search \
  --project-id="asdf" \
  --query="Ahoj"
```

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
