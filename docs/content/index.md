# Sentiary documentation

Manage translations in Sentiary and deliver them to your application with the CLI and Gradle plugin.

<div class="grid cards" markdown>

-   :material-console: **Use the CLI**

    Install the CLI, create `sentiary.toml`, and download your first translations.

    [Get started](getting-started/installation.md)

-   :material-puzzle: **Connect Gradle**

    Generate and stage Android and Compose resources as part of your build.

    [Set up Gradle](gradle/index.md)

-   :material-robot: **Connect an AI assistant**

    Manage projects, languages, and translations with the CLI.

    [CLI command reference](cli/reference.md)

-   :material-api: **Explore the API**

    Find authentication requirements, endpoints, and request and response schemas.

    [Open the API guide](api/index.md)

</div>

The MCP server has been removed. Existing MCP clients should migrate to the [CLI](getting-started/installation.md). AI assistants can use the [Sentiary CLI skill](https://github.com/MateeDevs/sentiary-tools/tree/main/.agents/skills/sentiary-cli).

```sh
sentiary-cli init
sentiary-cli pull --dry-run
sentiary-cli pull
```

Start with an existing Sentiary project and its language IDs. Follow the [first download guide](getting-started/quickstart.md) for credentials and generated-file ownership.

[Contribute](contributing/development/index.html) · [Browse by topic](topics/index.html) · [Troubleshooting](troubleshooting.md)

For AI tools: [llms.txt](llms.txt) · [Full documentation](llms-full.txt)
