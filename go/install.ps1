param([string]$Version = $env:SENTIARY_VERSION)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$repository = "MateeDevs/sentiary-tools"
$binary = "sentiary-cli"

$processorArchitecture = if ($env:PROCESSOR_ARCHITEW6432) {
    $env:PROCESSOR_ARCHITEW6432
} else {
    $env:PROCESSOR_ARCHITECTURE
}

$architecture = switch ($processorArchitecture.ToUpperInvariant()) {
    "AMD64" { "amd64" }
    "X86_64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "Unsupported architecture: $processorArchitecture" }
}

$asset = "$binary-windows-$architecture.exe"
if ($Version -and $Version -notmatch '^[a-zA-Z0-9][a-zA-Z0-9._-]*$') { throw "Invalid version" }
$baseUrl = if ($Version) { "https://github.com/$repository/releases/download/$Version" } else { "https://github.com/$repository/releases/latest/download" }
$url = "$baseUrl/$asset"
$installDirectory = if ($env:INSTALL_DIR) {
    $env:INSTALL_DIR
} else {
    Join-Path $env:LOCALAPPDATA "Programs\Sentiary\bin"
}
$installPath = Join-Path $installDirectory "$binary.exe"
$temporaryPath = Join-Path ([IO.Path]::GetTempPath()) "$([Guid]::NewGuid()).exe"

try {
    New-Item -ItemType Directory -Force -Path $installDirectory | Out-Null
    Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile $temporaryPath
    $manifest = (Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/SHA256SUMS").Content
    $checksumLines = @($manifest -split "`n" | Where-Object { $_ -match ("^[a-fA-F0-9]{64}\s+" + [regex]::Escape($asset) + "\s*$") })
    if ($checksumLines.Count -ne 1) { throw "Missing or ambiguous checksum for $asset" }
    $expected = ($checksumLines[0] -split '\s+')[0]
    $actual = (Get-FileHash -Algorithm SHA256 -Path $temporaryPath).Hash
    if ($actual -ine $expected) { throw "Binary checksum mismatch" }
    Move-Item -Force -Path $temporaryPath -Destination $installPath
} finally {
    if (Test-Path $temporaryPath) {
        Remove-Item -Force $temporaryPath
    }
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$pathEntries = @($userPath -split ";" | Where-Object { $_ })
$hasInstallDirectory = $pathEntries | Where-Object {
    $_.TrimEnd("\") -ieq $installDirectory.TrimEnd("\")
}

if (-not $hasInstallDirectory) {
    $newUserPath = (@($pathEntries) + $installDirectory) -join ";"
    [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
}

$currentPathEntries = @($env:Path -split ";")
if ($currentPathEntries -notcontains $installDirectory) {
    $env:Path = "$installDirectory;$env:Path"
}

$version = & $installPath --version
Write-Output "Installed $binary to $installPath"
if ($version) {
    Write-Output "Version: $version"
}
