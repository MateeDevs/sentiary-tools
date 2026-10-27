# CLI integration protocol

This machine interface is separate from ordinary CLI commands and the backend export/state protocol. The [Gradle settings plugin](../docs/gradle/index.md) uses it for resource bindings.

Download the [JSON Schema bundle](https://github.com/MateeDevs/sentiary-tools/blob/staging/docs/integration/schemas.json) for handshake, request, result, and preview shapes. Use a reference such as `schemas.json#/$defs/request`, `#/$defs/result`, `#/$defs/preview`, or `#/$defs/handshake` to select a shape from the bundle. The bundle describes emitted/accepted structure; semantic restrictions below also apply. Paths and digests in examples are illustrative.

## Version handshake

```sh
sentiary-cli version --json
```

```json
{"version":"dev","integrationProtocols":[1]}
```

`version` is a string; `integrationProtocols` is an array of supported integer versions. The current executable advertises `[1]`. Negotiate support before running a request; do not infer it solely from a release version. `--version --json` and `-v --json` also work. Extra version arguments return usage exit 2.

## Request

```json
{
  "protocolVersion": 1,
  "bindings": [
    {"output":"android","format":"android","resourceRoot":"/workspace/android/app/build/generated/sentiary"}
  ]
}
```

Run `sentiary-cli pull --config /workspace/sentiary.toml --integration-request request.json --json`.

| Field | Type | Contract |
| --- | --- | --- |
| `protocolVersion` | integer | Required, must equal `1` |
| `bindings` | array of objects | Selects named outputs; absent/null/empty currently selects all outputs with no root validation |
| `bindings[].output` | string | Existing output name; repeated names deduplicate selection |
| `bindings[].format` | string | Must equal output format, and output layout must be `per_language` |
| `bindings[].resourceRoot` | string | Absolute path; every generated file for that binding must be strictly beneath it |

Unknown fields, malformed JSON, a second JSON value, unsupported versions, unknown outputs, nonabsolute roots, mismatched format/layout, and files outside the root fail. Repeated bindings are each validated. `--output` and `--integration-request` cannot be combined. Request paths resolve from the working directory. Bindings do not rewrite destinations or define source sets; the TOML determines paths. Resolve symlinks in the checkout/config root before constructing `resourceRoot`, since the config root is canonicalized and containment uses path comparison. An empty request is accepted by the current decoder but integrations should send explicit nonempty bindings.

## Result

On successful non-preview `pull --json`, stdout is:

```json
{"protocolVersion":1,"files":[{"output":"android","path":"/workspace/android/app/build/generated/sentiary/values/strings.xml","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}]}
```

`protocolVersion` is integer `1`; `files` is an array (empty when no files), each with output name, absolute destination path, and lowercase 64-character SHA-256 of the written bytes. This shape is also emitted without an integration request. Failures return nonzero and a diagnostic on stderr, not a result object. Offline age diagnostics remain on stderr.

## Preview

`pull --dry-run` emits an array, **not** a protocol-versioned result, even when `--json` is also supplied:

```json
[{"output":"android","project":"mobile","format":"android","layout":"per_language","languageIds":["en"],"includeEmptyTranslations":false,"path":"/workspace/android/app/build/generated/sentiary/values/strings.xml"}]
```

Each item has string `output`, project **alias** `project`, string `format`, `layout` (`per_language` or `combined`), string-array `languageIds`, boolean `includeEmptyTranslations`, and absolute string `path`. `sourceLanguageId` is emitted only when nonempty (combined catalogs). Aliases report resolved source IDs in `languageIds`, while the path reflects the target locale.

Preview does not write destinations or publish an online completed set. Wildcards/aliases can cause language lookup requests. `--offline --dry-run` resolves from a covering completed set. Root validation applies to both previews and actual results, but a preview is not a check of ownership or future filesystem success.
