# CLI and hosted MCP

The Go module provides `sentiary-cli` for terminal workflows and `sentiary-mcp` for AI assistants.
Both frontends share the API client in `internal/sentiary`.

See the [CLI documentation](../docs/content/cli/reference.md),
[installation guide](../docs/content/getting-started/installation.md), and
[development guide](../contributing/development.md).

## API keys and languages

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

## Docker

From the repository root, build and run the hosted frontend:

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
