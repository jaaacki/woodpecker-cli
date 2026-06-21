$ErrorActionPreference = "Stop"

$Repo = if ($env:WPCI_REPO) { $env:WPCI_REPO } else { "jaaacki/woodpecker-cli" }
$Version = if ($env:WPCI_VERSION) { $env:WPCI_VERSION } else { "latest" }
$InstallDir = if ($env:WPCI_INSTALL_DIR) { $env:WPCI_INSTALL_DIR } else { Join-Path $HOME ".local\bin" }
$BinName = if ($env:WPCI_BIN_NAME) { $env:WPCI_BIN_NAME } else { "wpci.exe" }

$arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
  "X64" { "amd64" }
  "Arm64" { "arm64" }
  default { throw "unsupported architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
}

if ($Version -eq "latest") {
  $release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
  $tag = $release.tag_name
} else {
  $tag = $Version
}

if (-not $tag) {
  throw "could not determine release version for $Repo"
}

$asset = "wpci-windows-$arch.exe"
$base = "https://github.com/$Repo/releases/download/$tag"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("wpci-install-" + [System.Guid]::NewGuid())
New-Item -ItemType Directory -Force -Path $tmp | Out-Null

try {
  Write-Host "Installing $Repo $tag for windows/$arch"
  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
  $download = Join-Path $tmp $asset
  Invoke-WebRequest "$base/$asset" -OutFile $download

  try {
    $checksums = Join-Path $tmp "checksums.txt"
    Invoke-WebRequest "$base/checksums.txt" -OutFile $checksums
    $line = Get-Content $checksums | Where-Object { $_ -match "\s+$([regex]::Escape($asset))$" } | Select-Object -First 1
    if ($line) {
      $expected = ($line -split "\s+")[0]
      $actual = (Get-FileHash -Algorithm SHA256 $download).Hash.ToLowerInvariant()
      if ($actual -ne $expected.ToLowerInvariant()) {
        throw "checksum mismatch"
      }
    }
  } catch {
    if ($_.Exception.Message -eq "checksum mismatch") { throw }
  }

  $target = Join-Path $InstallDir $BinName
  Move-Item -Force $download $target
  Write-Host "Installed: $target"
  Write-Host "Next:"
  Write-Host "  wpci account add home --server https://ci.example.com"
  Write-Host "  wpci account token set home"
  Write-Host "  wpci home doctor"
} finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

