param(
  [string]$Version = $(if ($env:MONOLOGUE_VERSION) { $env:MONOLOGUE_VERSION } else { "latest" }),
  [string]$InstallDir = $(if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { Join-Path $HOME "AppData\Local\Programs\monologue-toolkit\bin" }),
  [string]$Repo = $(if ($env:MONOLOGUE_REPO) { $env:MONOLOGUE_REPO } else { "EveryInc/monologue-toolkit" })
)

$ErrorActionPreference = "Stop"

function Get-AssetNames {
  $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
  switch ($arch) {
    "x64" { $normalizedArch = "amd64" }
    "arm64" { $normalizedArch = "arm64" }
    default { throw "Unsupported architecture: $arch" }
  }

  return @{
    Archive = "monologue_windows_${normalizedArch}.zip"
    Checksums = "checksums.txt"
  }
}

$assetNames = Get-AssetNames
$releasePath = if ($Version -eq "latest") { "latest/download" } else { "download/$Version" }
$archiveUrl = "https://github.com/$Repo/releases/$releasePath/$($assetNames.Archive)"
$checksumsUrl = "https://github.com/$Repo/releases/$releasePath/$($assetNames.Checksums)"

$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("monologue-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmpDir | Out-Null

try {
  $archivePath = Join-Path $tmpDir $assetNames.Archive
  $checksumsPath = Join-Path $tmpDir $assetNames.Checksums

  Write-Host "Downloading $archiveUrl"
  Invoke-WebRequest -Uri $archiveUrl -OutFile $archivePath

  try {
    Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath | Out-Null
    $checksums = Get-Content $checksumsPath
    $line = $checksums | Where-Object { $_ -match [regex]::Escape($assetNames.Archive) + '$' } | Select-Object -First 1
    if ($line) {
      $expected = ($line -split '\s+')[0]
      $actual = (Get-FileHash -Algorithm SHA256 -Path $archivePath).Hash.ToLowerInvariant()
      if ($actual -ne $expected.ToLowerInvariant()) {
        throw "Checksum verification failed for $($assetNames.Archive)"
      }
    }
  } catch {
    Write-Host "Skipping checksum verification because checksums could not be read."
  }

  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
  Expand-Archive -Path $archivePath -DestinationPath $tmpDir -Force
  Copy-Item -Path (Join-Path $tmpDir "monologue.exe") -Destination (Join-Path $InstallDir "monologue.exe") -Force

  Write-Host ""
  Write-Host "Installed monologue.exe to $(Join-Path $InstallDir 'monologue.exe')"
  Write-Host ""
  Write-Host "Next steps:"
  Write-Host "  1. Add $InstallDir to your PATH if needed"
  Write-Host "  2. Run: monologue onboarding"
  Write-Host ""
  Write-Host "To update later, run: monologue update"
} finally {
  Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
}
