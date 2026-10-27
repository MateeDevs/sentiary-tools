---
tags:
  - CLI
---

# Installation

## Install

=== "macOS and Linux"

    Download the installer for inspection, then run it:

    ```sh
    curl -fsSL https://raw.githubusercontent.com/MateeDevs/sentiary-tools/main/install.sh -o install-sentiary.sh
    sh install-sentiary.sh
    ```

    It needs `curl` or `wget`, `tar`, and `sha256sum` or `shasum`. It detects OS and architecture, downloads the archive and `SHA256SUMS`, and checks the archive before extraction and installation.

    The default directory is `/usr/local/bin` when writable, otherwise `$HOME/.local/bin`. Override it with `INSTALL_DIR`. The installer prints a PATH reminder; it does not edit shell startup files.

    ```sh
    INSTALL_DIR="$HOME/.local/bin" sh install-sentiary.sh
    export PATH="$HOME/.local/bin:$PATH"
    sentiary-cli --version
    ```

    Add the PATH export to your shell startup file if needed. To pin a release, replace `vX.Y.Z` with an existing release tag:

    ```sh
    SENTIARY_VERSION=vX.Y.Z sh install-sentiary.sh
    # Alternatively, use the first argument:
    sh install-sentiary.sh vX.Y.Z
    ```

    `SENTIARY_VERSION` takes precedence over the first argument. With neither, the installer uses GitHub's latest release. `INSTALL_DIR` and the version controls are installer inputs, not CLI flags.

=== "Windows"

    ```powershell
    Invoke-WebRequest https://raw.githubusercontent.com/MateeDevs/sentiary-tools/main/install.ps1 -OutFile install-sentiary.ps1
    .\install-sentiary.ps1
    sentiary-cli --version
    ```

    PowerShell execution policy must permit running the downloaded script. The default directory is `$env:LOCALAPPDATA\Programs\Sentiary\bin`; `INSTALL_DIR` overrides it. The installer adds that directory to both the user PATH and the current process PATH. It checks the downloaded executable using `Get-FileHash -Algorithm SHA256` before replacing the installed binary.

    ```powershell
    $env:INSTALL_DIR = "$env:LOCALAPPDATA\Programs\Sentiary\bin"
    .\install-sentiary.ps1 -Version vX.Y.Z
    ```

    `-Version` defaults to `SENTIARY_VERSION`; an explicit parameter wins. Without either, it uses the latest release. Both AMD64 and ARM64 executables are available.

## Platforms and releases

The release workflow builds AMD64 and ARM64 for macOS (`darwin`), Linux (`linux`), and Windows (`windows`). Intel/x86-64 maps to `amd64`; Apple Silicon/aarch64 maps to `arm64`. The installers reject other architectures. There is no 32-bit installer target.

Download assets from [Sentiary Tools releases](https://github.com/MateeDevs/sentiary-tools/releases). Unix assets are `sentiary-cli-<os>-<arch>.tar.gz`; Windows assets are `sentiary-cli-windows-<arch>.exe`. `SHA256SUMS` accompanies releases.

## Manual checksum verification

Download the asset and `SHA256SUMS` from the **same release**. On Linux, use `sha256sum <asset>`; on macOS use `shasum -a 256 <asset>`; on Windows use `Get-FileHash <asset> -Algorithm SHA256`. Compare the complete hash with the exact asset's line in `SHA256SUMS`. Reject missing, ambiguous, or mismatched entries. For Unix, verify the archive before extracting it. The manifest is fetched from the release; it is not an independently signed attestation.

## Upgrade and inspect

Rerun the installer to replace the binary with the latest release, or specify an existing tag for a reproducible upgrade/downgrade. It does not migrate your TOML, generated files, or cache. Check which binary your shell resolves (`command -v sentiary-cli` or `Get-Command sentiary-cli`) if the reported version is unexpected.

`version`, `--version`, and `-v` print the version. `version --json` prints the version and supported integration protocols. See [protocol negotiation](https://github.com/MateeDevs/sentiary-tools/blob/staging/repository-docs/integration-protocol.md#version-handshake).

For local development, see [building the CLI from source](https://github.com/MateeDevs/sentiary-tools/blob/staging/repository-docs/maintenance.md#build-from-source).
