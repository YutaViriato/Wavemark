# Installs the wavemark binary on Windows.
#
#   irm https://wavemark.viriato.site/install.ps1 | iex
#
# Environment:
#   WAVEMARK_VERSION       version to install, e.g. v0.1.0 (default: latest)
#   WAVEMARK_INSTALL_DIR   where to put the binary (default: %LOCALAPPDATA%\wavemark)
#   WAVEMARK_DOWNLOAD_URL  base URL of the release files (default: GitHub Releases)
$ErrorActionPreference = 'Stop'

$Repo = 'YutaViriato/Wavemark'
$Version = if ($env:WAVEMARK_VERSION) { $env:WAVEMARK_VERSION } else { 'latest' }
$InstallDir = if ($env:WAVEMARK_INSTALL_DIR) { $env:WAVEMARK_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'wavemark' }

$Arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'Arm64') { 'arm64' } else { 'amd64' }
$Asset = "wavemark_windows_$Arch.zip"
$Base = if ($env:WAVEMARK_DOWNLOAD_URL) { $env:WAVEMARK_DOWNLOAD_URL.TrimEnd('/') }
        elseif ($Version -eq 'latest') { "https://github.com/$Repo/releases/latest/download" }
        else { "https://github.com/$Repo/releases/download/$Version" }

$Tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("wavemark-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $Tmp | Out-Null
try {
    Write-Host "wavemark: downloading $Asset ($Version)"
    Invoke-WebRequest "$Base/$Asset" -OutFile (Join-Path $Tmp $Asset) -UseBasicParsing
    Invoke-WebRequest "$Base/checksums.txt" -OutFile (Join-Path $Tmp 'checksums.txt') -UseBasicParsing

    $Line = Get-Content (Join-Path $Tmp 'checksums.txt') | Where-Object { $_ -match " $([regex]::Escape($Asset))$" }
    if (-not $Line) { throw "$Asset is not listed in checksums.txt" }
    $Expected = ($Line -split ' ')[0]
    $Actual = (Get-FileHash (Join-Path $Tmp $Asset) -Algorithm SHA256).Hash.ToLower()
    if ($Expected -ne $Actual) { throw "checksum mismatch for $Asset" }

    Expand-Archive (Join-Path $Tmp $Asset) -DestinationPath $Tmp -Force
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Move-Item (Join-Path $Tmp 'wavemark.exe') (Join-Path $InstallDir 'wavemark.exe') -Force

    $UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (($UserPath -split ';') -notcontains $InstallDir) {
        [Environment]::SetEnvironmentVariable('Path', "$UserPath;$InstallDir", 'User')
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "wavemark: added $InstallDir to your PATH (new terminals pick it up)"
    }
    Write-Host "wavemark: installed $(& (Join-Path $InstallDir 'wavemark.exe') version)"
    Write-Host "wavemark: next: wavemark init my-site; wavemark dev my-site"
}
finally {
    Remove-Item $Tmp -Recurse -Force -ErrorAction SilentlyContinue
}
