# Install the refuse CLI from the latest GitHub release (Windows).
#
# Usage:  irm https://raw.githubusercontent.com/RefuseHQ/refuse-cli/main/scripts/install.ps1 | iex
#
# Honours $env:REFUSE_INSTALL_DIR (default $env:USERPROFILE\.refuse\bin) and
# $env:REFUSE_VERSION (default latest). Verifies a sha256 checksum from the
# release.

$ErrorActionPreference = 'Stop'

$Repo       = 'RefuseHQ/refuse-cli'
$InstallDir = if ($env:REFUSE_INSTALL_DIR) { $env:REFUSE_INSTALL_DIR } else { Join-Path $env:USERPROFILE '.refuse\bin' }
$Version    = if ($env:REFUSE_VERSION)     { $env:REFUSE_VERSION }     else { 'latest' }

function Get-Arch {
    switch -Regex ($env:PROCESSOR_ARCHITECTURE) {
        '^(AMD64|x86_64)$' { return 'x86_64' }
        '^ARM64$'          { return 'arm64' }
        '^(x86|i.86)$'     { return 'i386' }
        default            { throw "unsupported arch: $env:PROCESSOR_ARCHITECTURE" }
    }
}

$Arch = Get-Arch

if ($Version -eq 'latest') {
    $resp = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
    if (-not $resp.tag_name) { throw 'could not resolve latest version' }
    $Version = $resp.tag_name -replace '^v', ''
}

$Archive  = "refuse_windows_${Arch}.zip"
$Url      = "https://github.com/$Repo/releases/download/v$Version/$Archive"
$SumsUrl  = "https://github.com/$Repo/releases/download/v$Version/checksums.txt"

$Tmp = Join-Path $env:TEMP "refuse-install-$([guid]::NewGuid())"
New-Item -ItemType Directory -Path $Tmp -Force | Out-Null

try {
    Write-Host "refuse: downloading $Url"
    $zipPath = Join-Path $Tmp $Archive
    Invoke-WebRequest -Uri $Url -OutFile $zipPath -UseBasicParsing

    Write-Host 'refuse: verifying checksum'
    $sumsPath = Join-Path $Tmp 'checksums.txt'
    Invoke-WebRequest -Uri $SumsUrl -OutFile $sumsPath -UseBasicParsing

    $expected = (Get-Content $sumsPath | Where-Object { $_ -match "\s$([regex]::Escape($Archive))$" } | Select-Object -First 1) -replace '\s.*$', ''
    if (-not $expected) { throw "checksum line for $Archive not found" }
    $actual = (Get-FileHash -Algorithm SHA256 -Path $zipPath).Hash.ToLower()
    if ($expected.ToLower() -ne $actual) {
        throw "checksum mismatch: expected $expected got $actual"
    }

    Write-Host "refuse: extracting to $InstallDir"
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Expand-Archive -Path $zipPath -DestinationPath $Tmp -Force
    Move-Item -Path (Join-Path $Tmp 'refuse.exe') -Destination (Join-Path $InstallDir 'refuse.exe') -Force
}
finally {
    Remove-Item -Path $Tmp -Recurse -Force -ErrorAction SilentlyContinue
}

# Add to user PATH if not already present.
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (-not ($userPath -split ';' | Where-Object { $_ -ieq $InstallDir })) {
    [Environment]::SetEnvironmentVariable('Path', "$userPath;$InstallDir", 'User')
    Write-Host "refuse: added $InstallDir to user PATH (open a new terminal to pick it up)"
}

Write-Host "refuse: installed $InstallDir\refuse.exe"
Write-Host 'refuse: try `refuse --version` then `refuse init` (in a new terminal)'
