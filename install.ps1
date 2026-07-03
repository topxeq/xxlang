# Xxlang Installation / Update Script for Windows
# Downloads and installs the latest version of Xxlang from GitHub Releases.
# If Xxlang is already installed, compares versions and skips the download
# when the installed version matches the latest release (use -Force to
# reinstall anyway).
#
# Usage:
#   iwr -useb https://raw.githubusercontent.com/topxeq/xxlang/master/install.ps1 | iex
#
# Or with PowerShell 7+:
#   irm https://raw.githubusercontent.com/topxeq/xxlang/master/install.ps1 | iex
#
# Force reinstall:
#   iwr -useb https://raw.githubusercontent.com/topxeq/xxlang/master/install.ps1 | iex -ArgumentList "-Force"
#

param(
    [string]$InstallDir = "",
    [switch]$NoPathUpdate,
    [switch]$Force
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
    # Try .NET RuntimeInformation first (requires .NET 4.7.1+)
    try {
        $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
        if ($arch) {
            switch ($arch) {
                "X64" { return "amd64" }
                "X86" { return "386" }
                "Arm64" { return "arm64" }
                "Arm" { return "arm" }
                default { return $arch.ToString().ToLower() }
            }
        }
    } catch {
        # Fall through to WMI method
    }

    # Fallback: Use environment variable (works on all Windows versions)
    $cpuArch = $env:PROCESSOR_ARCHITECTURE
    switch ($cpuArch) {
        "AMD64" { return "amd64" }
        "x86" { return "386" }
        "ARM64" { return "arm64" }
        "ARM" { return "arm" }
        default {
            # Last resort: Use WMI
            try {
                $proc = Get-CimInstance -ClassName Win32_Processor -ErrorAction SilentlyContinue
                if ($proc) {
                    $arch = $proc.AddressWidth
                    switch ($arch) {
                        64 { return "amd64" }
                        32 { return "386" }
                    }
                }
            } catch {
                # Ignore errors
            }
            # Default to amd64 as most common
            return "amd64"
        }
    }
}

# Compare two semantic version strings of the form X.Y.Z.
# Returns 0 if equal, -1 if $a < $b, 1 if $a > $b.
function Compare-Version {
    param([string]$a, [string]$b)
    if ($a -eq $b) { return 0 }
    $aParts = $a -split '\.' | ForEach-Object { [int]($_) }
    $bParts = $b -split '\.' | ForEach-Object { [int]($_) }
    # Pad to equal length
    while ($aParts.Count -lt 3) { $aParts += 0 }
    while ($bParts.Count -lt 3) { $bParts += 0 }
    for ($i = 0; $i -lt 3; $i++) {
        if ($aParts[$i] -lt $bParts[$i]) { return -1 }
        if ($aParts[$i] -gt $bParts[$i]) { return 1 }
    }
    return 0
}

# Get the version of the currently installed xxl.exe, if any.
# Returns the version string (e.g. "0.9.10") on success, $null on failure.
function Get-InstalledVersion {
    param([string]$BinPath)

    if ([string]::IsNullOrEmpty($BinPath) -or -not (Test-Path $BinPath)) {
        return $null
    }
    try {
        # `xxl version` prints "Xxlang v0.9.10". Extract the X.Y.Z part.
        $output = & $BinPath version 2>$null
        if ($output -match 'v(\d+\.\d+\.\d+)') {
            return $matches[1]
        }
    } catch {
        # Binary might be wrong arch, corrupted, etc. Treat as not installed.
    }
    return $null
}

# Get latest version from GitHub
function Get-LatestVersion {
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing -ErrorAction Stop
        if ($release -and $release.tag_name) {
            $version = $release.tag_name -replace "^v", ""
            if ($version) {
                return $version
            }
        }
    } catch {
        Write-Warn "Could not fetch from API, trying HTML fallback..."
    }

    # Fallback: parse HTML
    try {
        $html = Invoke-WebRequest -Uri "https://github.com/$Repo/releases/latest" -UseBasicParsing -ErrorAction Stop
        if ($html.Content -match "tag/v(\d+\.\d+\.\d+)") {
            return $matches[1]
        }
    } catch {
        Write-Err "Failed to get latest version from GitHub: $_"
        exit 1
    }

    Write-Err "Could not determine latest version"
    exit 1
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

    # Determine install directory first, so we can check the currently
    # installed version before hitting the network.
    if ([string]::IsNullOrEmpty($InstallDir)) {
        # Default to user's local bin or AppData
        $localAppData = [Environment]::GetFolderPath("LocalApplicationData")
        if ([string]::IsNullOrEmpty($localAppData)) {
            $localAppData = $env:LOCALAPPDATA
        }
        if ([string]::IsNullOrEmpty($localAppData)) {
            $localAppData = Join-Path $env:USERPROFILE "AppData\Local"
        }
        $InstallDir = Join-Path $localAppData "xxlang"
    }

    # Create install directory if needed
    if (-not (Test-Path $InstallDir)) {
        Write-Info "Creating install directory: $InstallDir"
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $installPath = Join-Path $InstallDir $BinaryName

    # Get latest version
    Write-Info "Fetching latest version..."
    $version = Get-LatestVersion
    Write-Info "Latest version: $version"

    # Check the currently installed version (if any) and skip the download
    # when it already matches the latest release.
    $installedVersion = Get-InstalledVersion -BinPath $installPath
    # Also try `xxl` from PATH in case installPath doesn't point at it.
    if ([string]::IsNullOrEmpty($installedVersion)) {
        $xxlInPath = Get-Command xxl -ErrorAction SilentlyContinue
        if ($xxlInPath) {
            $installedVersion = Get-InstalledVersion -BinPath $xxlInPath.Source
        }
    }

    if (-not [string]::IsNullOrEmpty($installedVersion)) {
        Write-Info "Installed version: $installedVersion"
        if (-not $Force) {
            $cmp = Compare-Version -a $installedVersion -b $version
            if ($cmp -eq 0) {
                Write-Success "Already up to date (v$installedVersion). Nothing to do."
                Write-Success "Use -Force to reinstall."
                exit 0
            } elseif ($cmp -gt 0) {
                # Installed is newer than the latest release tag — happens for
                # local builds ahead of a release. Don't downgrade.
                Write-Warn "Installed version ($installedVersion) is newer than the latest release ($version). Not downgrading."
                exit 0
            }
            Write-Info "Update available: $installedVersion -> $version"
        } else {
            Write-Warn "Force reinstall requested — ignoring installed version $installedVersion."
        }
    } else {
        Write-Info "No previous installation detected — performing fresh install."
    }

    # Build download URL - using zip archive
    # Format: xxlang-windows-{arch}.zip
    $archiveName = "xxlang-windows-$arch.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/v$version/$archiveName"

    Write-Info "Download URL: $downloadUrl"

    # Download archive
    Write-Info "Downloading Xxlang v$version..."
    try {
        $tempArchive = Join-Path $env:TEMP $archiveName
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempArchive -UseBasicParsing
    } catch {
        Write-Err "Failed to download Xxlang: $_"
        exit 1
    }

    # Extract archive
    Write-Info "Extracting..."
    try {
        $extractDir = Join-Path $env:TEMP "xxlang-extract-$(Get-Random)"
        New-Item -ItemType Directory -Path $extractDir -Force | Out-Null

        # Use .NET's ZipFile for extraction
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        [System.IO.Compression.ZipFile]::ExtractToDirectory($tempArchive, $extractDir)

        # Find the extracted binary
        $extractedBinary = Join-Path $extractDir $BinaryName
        if (-not (Test-Path $extractedBinary)) {
            Write-Err "Binary not found in archive: $BinaryName"
            Remove-Item $tempArchive -Force
            Remove-Item $extractDir -Recurse -Force
            exit 1
        }
    } catch {
        Write-Err "Failed to extract archive: $_"
        Remove-Item $tempArchive -Force -ErrorAction SilentlyContinue
        exit 1
    }

    # Replace the existing binary. On Windows the running exe cannot be
    # removed or overwritten while it is mapped, but renaming it aside first
    # is allowed. This matches the strategy used by `xxl update` (v0.9.9+):
    #   1. Copy the new binary to installPath + ".new" (sibling, so same drive).
    #   2. Rename the current exe to ".old".
    #   3. Rename ".new" into place.
    #   4. Best-effort remove ".old".
    # Each step tolerates the previous step having been skipped, so the
    # script is idempotent across failed runs.
    if (Test-Path $installPath) {
        Write-Info "Replacing existing installation..."
        $newPath = "$installPath.new"
        $oldPath = "$installPath.old"

        # Stage the new binary next to the target.
        Remove-Item $newPath -Force -ErrorAction SilentlyContinue
        Copy-Item $extractedBinary $newPath -Force

        # Move the current executable aside (Windows allows renaming a
        # running exe; only creating a new file with the same name while
        # the old one is mapped is restricted).
        Remove-Item $oldPath -Force -ErrorAction SilentlyContinue
        if (Test-Path $installPath) {
            Move-Item $installPath $oldPath -Force
        }

        # Move the new binary into place.
        Move-Item $newPath $installPath -Force

        # Best-effort cleanup of the old executable. If the old binary is
        # still mapped (e.g. an xxl process is still running), this fails —
        # that's harmless and it will be removable after that process exits.
        Remove-Item $oldPath -Force -ErrorAction SilentlyContinue
    } else {
        Write-Info "Installing to $installPath..."
        Move-Item $extractedBinary $installPath -Force
    }

    # Cleanup
    Remove-Item $tempArchive -Force -ErrorAction SilentlyContinue
    Remove-Item $extractDir -Recurse -Force -ErrorAction SilentlyContinue

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
