# Xxlang Installation Script for Windows
# Downloads and installs the latest version of Xxlang from GitHub Releases
#
# Usage:
#   iwr -useb https://raw.githubusercontent.com/topxeq/xxlang/master/install.ps1 | iex
#
# Or with PowerShell 7+:
#   irm https://raw.githubusercontent.com/topxeq/xxlang/master/install.ps1 | iex
#

param(
    [string]$InstallDir = "",
    [switch]$NoPathUpdate
)

$ErrorActionPreference = "Stop"
$Repo = "topxeq/xxlang"
$BinaryName = "xxl.exe"

# Print functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] " -ForegroundColor Blue -NoNewline
    Write-Host $Message
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
}

function Write-Err {
    param([string]$Message)
    Write-Host "[ERROR] " -ForegroundColor Red -NoNewline
    Write-Host $Message
}

# Detect architecture
function Get-Architecture {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch) {
        "X64" { return "amd64" }
        "X86" { return "386" }
        "Arm64" { return "arm64" }
        "Arm" { return "arm" }
        default { return $arch.ToString().ToLower() }
    }
}

# Get latest version from GitHub
function Get-LatestVersion {
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
        $version = $release.tag_name -replace "^v", ""
        return $version
    } catch {
        # Fallback: parse HTML
        try {
            $html = Invoke-WebRequest -Uri "https://github.com/$Repo/releases/latest" -UseBasicParsing
            if ($html.Content -match "tag/v(\d+\.\d+\.\d+)") {
                return $matches[1]
            }
        } catch {
            Write-Err "Failed to get latest version from GitHub"
            exit 1
        }
    }
}

# Main installation
function Main {
    Write-Host ""
    Write-Host "======================================" -ForegroundColor Green
    Write-Host "    Xxlang Installation Script       " -ForegroundColor Green
    Write-Host "======================================" -ForegroundColor Green
    Write-Host ""

    # Detect architecture
    $arch = Get-Architecture
    Write-Info "Detected Architecture: $arch"

    # Get latest version
    Write-Info "Fetching latest version..."
    $version = Get-LatestVersion
    Write-Info "Latest version: $version"

    # Build download URL
    $assetName = "xxlang-$version-windows-$arch.exe"
    $downloadUrl = "https://github.com/$Repo/releases/download/v$version/$assetName"

    Write-Info "Download URL: $downloadUrl"

    # Determine install directory
    if ([string]::IsNullOrEmpty($InstallDir)) {
        # Default to user's local bin or AppData
        $localAppData = [Environment]::GetFolderPath("LocalApplicationData")
        $InstallDir = Join-Path $localAppData "xxlang"
    }

    # Create install directory if needed
    if (-not (Test-Path $InstallDir)) {
        Write-Info "Creating install directory: $InstallDir"
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $installPath = Join-Path $InstallDir $BinaryName

    # Download
    Write-Info "Downloading Xxlang v$version..."
    try {
        $tempFile = Join-Path $env:TEMP $BinaryName
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
    } catch {
        Write-Err "Failed to download Xxlang: $_"
        exit 1
    }

    # Remove existing installation
    if (Test-Path $installPath) {
        Write-Info "Removing existing installation..."
        Remove-Item $installPath -Force
    }

    # Install
    Write-Info "Installing to $installPath..."
    Move-Item $tempFile $installPath -Force

    # Verify installation
    if (Test-Path $installPath) {
        Write-Success "Xxlang v$version installed successfully!"
        Write-Host ""
        & $installPath version
        Write-Host ""
        Write-Success "Installation path: $installPath"

        # Add to PATH if requested
        if (-not $NoPathUpdate) {
            $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
            if ($userPath -notlike "*$InstallDir*") {
                Write-Info "Adding $InstallDir to user PATH..."
                [Environment]::SetEnvironmentVariable("Path", "$InstallDir;$userPath", "User")
                $env:Path = "$InstallDir;$env:Path"
                Write-Success "PATH updated. You may need to restart your terminal."
            }
        }

        Write-Host ""
        Write-Host "Quick Start:" -ForegroundColor Green
        Write-Host "    xxl                    # Start REPL"
        Write-Host "    xxl run script.xxl     # Run a script"
        Write-Host "    xxl update             # Update to latest version"
        Write-Host "    xxl help               # Show help"
        Write-Host ""
    } else {
        Write-Err "Installation failed"
        exit 1
    }
}

# Run main
Main
