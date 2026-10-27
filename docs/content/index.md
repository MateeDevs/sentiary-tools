# Sentiary documentation

Manage translations in Sentiary and deliver them to your application with CLI and Gradle 2.0.

<div class="grid cards" markdown>

-   :material-console: **Use the CLI**

    Install the CLI, create `sentiary.toml`, and download your first translations.

    [Get started](getting-started/installation.md)

-   :material-puzzle: **Connect Gradle**

    Generate and stage Android and Compose resources as part of your build.

    [Set up Gradle](gradle/index.md)

-   :material-robot: **Connect an AI assistant**

    Manage projects, languages, and translations through hosted MCP.

    [Set up MCP](mcp/index.md)

-   :material-api: **Explore the API**

    Find authentication requirements, endpoints, and request and response schemas.

    [Open the API guide](api/index.md)

</div>

```sh
sentiary-cli init
sentiary-cli pull --dry-run
sentiary-cli pull
```

Start with an existing Sentiary project and its language IDs. Follow the [first download guide](getting-started/quickstart.md) for credentials and generated-file ownership.

[Contribute](contributing/development/index.html) · [Browse by topic](topics/index.html) · [Migrate from Gradle 1.x](migration.md) · [Troubleshooting](troubleshooting.md)

For AI tools: [llms.txt](llms.txt) · [Full documentation](llms-full.txt)
