# Android, KMP Compose, and Xcode

Download the [mobile configuration](sentiary.toml) as `sentiary.toml`. It defines one project, explicit `en`/`cs` languages, and three named outputs. Replace its project ID and project-key placeholder, or remove `api_key` and provide `SENTIARY_USER_API_KEY` instead.

For automatic staging with the unreleased 2.0 settings plugin, use the [Gradle walkthrough](gradle.md) with this same TOML. The instructions below describe manual CLI orchestration.

## Android

The `android` output writes `android/app/build/generated/sentiary/values/strings.xml` and `values-cs/strings.xml`. The map sends `en` to the unqualified default `values` folder and `cs` to `values-cs`; the API still receives the original IDs.

```sh
sentiary-cli pull --output android --dry-run
sentiary-cli pull --output android
```

For manual integration, register the generated resource root in the Android module's `build.gradle.kts`:

```kotlin
android {
    sourceSets.getByName("main").res.srcDir("build/generated/sentiary")
}
```

Run the CLI before Gradle in your local/CI build sequence. This snippet registers a directory; it does not add a download task dependency. After generation, build your app and check that its resource merge consumes the generated files. Do not register both generated and manually copied definitions of the same keys.

## KMP Compose

The `compose` output writes under `shared/build/generated/sentiary/composeResources`. It uses the same explicit locale map, with `values/strings.xml` and `values-cs/strings.xml`.

```sh
sentiary-cli pull --output compose --dry-run
sentiary-cli pull --output compose
```

For Compose Multiplatform versions exposing `customDirectory`, configure the shared module:

```kotlin
compose.resources {
    customDirectory(
        sourceSetName = "commonMain",
        directoryProvider = provider {
            layout.buildDirectory.dir("generated/sentiary/composeResources").get()
        },
    )
}
```

Run the CLI before the Compose resource preparation/build tasks. Confirm this API against your installed Compose version; this manual snippet is not a declaration of Gradle plugin 2.0 compatibility. The [Compose resource documentation](https://www.jetbrains.com/help/kotlin-multiplatform-dev/compose-multiplatform-resources-setup.html) describes custom resource directories.

## Combined Xcode catalog

The `ios` output uses `xcstrings`, `combined`, and a fixed `ios/Resources/Localizable.xcstrings` path. It inherits `source_language = "en"` and both project languages. It creates one catalog, not one catalog per language.

```sh
sentiary-cli pull --output ios --dry-run
sentiary-cli pull --output ios
```

Add the generated catalog to your Xcode target's resources. Ensure the file exists before the Xcode build. Set a source-language override on the output only when that language is included in the selection. `include_empty = true` writes empty localizations with state `new`; no fallback text is inserted. Do not edit generated text in Xcode expecting a merge back to Sentiary.

Pull several outputs together with `--output android --output compose --output ios`, or omit selection for all. Offline restoration follows the same selection rules. The backend must support combined export and export/state protocol 1.
