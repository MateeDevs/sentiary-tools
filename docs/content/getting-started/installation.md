---
tags:
  - CLI
---

# Installation

## Install

=== "macOS and Linux"

    Download and execute the installer:

    ```sh
    curl -fsSL https://raw.githubusercontent.com/MateeDevs/sentiary-tools/main/go/install.sh | sh
    ```

    It needs `curl` or `wget`, `tar`, and `sha256sum` or `shasum`. It detects OS and architecture, downloads the archive and `SHA256SUMS`, and checks the archive before extraction and installation.

    The default directory is `/usr/local/bin` when writable, otherwise `$HOME/.local/bin`. Override it with `INSTALL_DIR`. The installer prints a PATH reminder; it does not edit shell startup files.

    ```sh
    curl -fsSL https://raw.githubusercontent.com/MateeDevs/sentiary-tools/main/go/install.sh | INSTALL_DIR="$HOME/.local/bin" sh
    export PATH="$HOME/.local/bin:$PATH"
    sentiary-cli --version
    ```

    Add the PATH export to your shell startup file if needed. To pin a release, replace `vX.Y.Z` with an existing release tag:

    ```sh
    curl -fsSL https://raw.githubusercontent.com/MateeDevs/sentiary-tools/main/go/install.sh | SENTIARY_VERSION=vX.Y.Z sh
    # Alternatively, use the first argument:
    curl -fsSL https://raw.githubusercontent.com/MateeDevs/sentiary-tools/main/go/install.sh | sh -s -- vX.Y.Z
    ```

    `SENTIARY_VERSION` takes precedence over the first argument. With neither, the installer uses GitHub's latest release. `INSTALL_DIR` and the version controls are installer inputs, not CLI flags.

=== "Windows"

    ```powershell
    & ([scriptblock]::Create((Invoke-WebRequest -UseBasicParsing https://raw.githubusercontent.com/MateeDevs/sentiary-tools/main/go/install.ps1).Content))
    sentiary-cli --version
    ```

    This executes the downloaded script in the current PowerShell session. The default directory is `$env:LOCALAPPDATA\Programs\Sentiary\bin`; `INSTALL_DIR` overrides it. The installer adds that directory to both the user PATH and the current process PATH. It checks the downloaded executable using `Get-FileHash -Algorithm SHA256` before replacing the installed binary.

    ```powershell
    $env:INSTALL_DIR = "$env:LOCALAPPDATA\Programs\Sentiary\bin"
    & ([scriptblock]::Create((Invoke-WebRequest -UseBasicParsing https://raw.githubusercontent.com/MateeDevs/sentiary-tools/main/go/install.ps1).Content)) -Version vX.Y.Z
    ```

    `-Version` defaults to `SENTIARY_VERSION`; an explicit parameter wins. Without either, it uses the latest release. Both AMD64 and ARM64 executables are available.

## Platforms and releases

These one-liners execute a downloaded script with your current user's permissions. Only run them if you trust the source. To inspect first, download the script to a file and read it before running it. The installer's checksum verification covers the release binary or archive, not the installer script itself. It does not independently authenticate the release manifest.

The release workflow builds AMD64 and ARM64 for macOS (`darwin`), Linux (`linux`), and Windows (`windows`). Intel/x86-64 maps to `amd64`; Apple Silicon/aarch64 maps to `arm64`. The installers reject other architectures. There is no 32-bit installer target.

Download assets from [Sentiary Tools releases](https://github.com/MateeDevs/sentiary-tools/releases). Unix assets are `sentiary-cli-<os>-<arch>.tar.gz`; Windows assets are `sentiary-cli-windows-<arch>.exe`. `SHA256SUMS` accompanies releases.

## Manual checksum verification

Download the asset and `SHA256SUMS` from the **same release**. On Linux, use `sha256sum <asset>`; on macOS use `shasum -a 256 <asset>`; on Windows use `Get-FileHash <asset> -Algorithm SHA256`. Compare the complete hash with the exact asset's line in `SHA256SUMS`. Reject missing, ambiguous, or mismatched entries. For Unix, verify the archive before extracting it. The manifest is fetched from the release; it is not an independently signed attestation.

## Upgrade and inspect

Use `sentiary-cli update` to check immediately and approve an update interactively. `sentiary-cli update --check` only reports availability; `sentiary-cli update --yes` explicitly approves installation in scripts or CI. Declining succeeds; check/download/install failures return a nonzero exit status. Explicit updates ignore automatic opt-outs and skipped releases.

After successful interactive `init` or online `pull` commands, the CLI consults a per-user cache. Other commands emit JSON or raw export streams and do not trigger automatic updates. It starts a detached background check at most once every 24 hours; ordinary commands never wait on that network check. A cached newer version offers **Update now**, **Remind me later** (24 hours), or **Skip this version** (that exact release). The prompt can delay command exit. EOF means remind later. An approved update failure does not change the completed command's exit status. Failed automatic checks are silent, retain cached availability, and retry in the next daily window. Concurrent invocations coordinate checks, prompts, and replacement.

Automatic checking and prompts are disabled in CI, noninteractive terminals, JSON output, help/version/update commands, offline pulls, and when `SENTIARY_NO_UPDATE=1`. Development and staging builds never self-update. Self-replacement requires a stable official binary with a matching installer-created ownership record. Package-manager/source builds and legacy unmarked installs cannot self-replace; reinstall with the official installer to establish ownership. No update requests elevation. Release downloads are checked against the same release's `SHA256SUMS`; replacement is prepared before touching the old executable.

You can also rerun the installer to replace the binary with the latest release, or specify an existing tag for a reproducible upgrade/downgrade. Neither method migrates your TOML, generated files, or export cache. Check which binary your shell resolves (`command -v sentiary-cli` or `Get-Command sentiary-cli`) if the reported version is unexpected.

`version`, `--version`, and `-v` print the version. `version --json` prints the version and supported integration protocols. See [protocol negotiation](../integration/protocol.md#version-handshake).

For local development, see [building the CLI from source](../contributing/documentation.md#build-from-source).
