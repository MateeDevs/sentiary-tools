# CI caching and offline recovery

Choose a cache path explicitly for portable CI. Cache the CLI's exports directory separately from generated `build/` directories. Do not treat a restored cache as proof of current authorization or server freshness; an online pull still authenticates and checks revisions.

A GitHub Actions sequence, after installing the CLI and checking out your application's config:

```yaml
- uses: actions/cache@v4
  with:
    path: .ci-cache/sentiary
    key: sentiary-${{ runner.os }}-${{ hashFiles('sentiary.toml') }}-${{ github.run_id }}
    restore-keys: |
      sentiary-${{ runner.os }}-${{ hashFiles('sentiary.toml') }}-
      sentiary-${{ runner.os }}-
- name: Download translations
  env:
    SENTIARY_USER_API_KEY: ${{ secrets.SENTIARY_USER_API_KEY }}
    SENTIARY_CACHE_DIR: .ci-cache/sentiary
  run: sentiary-cli pull
```

The run-specific key permits saving refreshed sets; restore prefixes reuse previous content. Restrict access to cached translations as appropriate for your project. Do not include credentials or `sentiary.local.toml` in cache paths. Ignore `.ci-cache/` locally if using this convention. Pin the CLI release separately.

## Restore deleted generated files

With the mobile config, prime Android and Compose together:

```sh
sentiary-cli pull --cache-dir .ci-cache/sentiary --output android --output compose
```

Delete only the generated Android and Compose directories using your normal clean command or file manager. Keep `.sentiary/` and `.ci-cache/sentiary/`. Then restore:

```sh
sentiary-cli pull --offline --cache-dir .ci-cache/sentiary --output android --output compose
```

The command restores missing files without a key or network. Verify the resulting paths with your app build. It prints the completed set's timestamp and age to stderr. A cache primed only for Android cannot restore Compose. Separate completed sets are not combined; prime the joint selection online first.

For a planned disconnected CI build, use the same explicit offline invocation after restoring a trusted complete cache. Do not automatically retry a failed online pull offline: doing so can conceal revoked access or service errors. The 2.0 settings plugin forwards Gradle's `--offline` to the CLI; this is verified locally. Standalone CLI invocations still need the explicit `--offline` flag.
