param([string]$Binary = "loom-mario-term")
$ErrorActionPreference = "Stop"
$Repo       = "ujjwalvivek/loom"
$InstallDir = "$env:LOCALAPPDATA\Programs"
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "x86" }
$Os = "windows"
Write-Host "Fetching latest release..." -ForegroundColor Cyan
$releases = "https://api.github.com/repos/$Repo/releases/latest"
try { $release = Invoke-RestMethod -Uri $releases -Headers @{ "Accept" = "application/vnd.github.v3+json" } }
catch {
  Write-Host "ERROR: Could not find any releases. Has a release been published on GitHub?" -ForegroundColor Red
  exit 1
}
$tag = $release.tag_name
$archive = "${Binary}_${Os}_${Arch}.tar.gz"
$url     = "https://github.com/$Repo/releases/download/$tag/$archive"
$tmp     = "$env:TEMP\$archive"
Write-Host "Downloading $url ..." -ForegroundColor Cyan
try {
  Invoke-WebRequest -Uri $url -OutFile $tmp
} catch {
  Remove-Item -Path $tmp -Force -ErrorAction SilentlyContinue
  Write-Host "ERROR: Failed to download $url" -ForegroundColor Red
  Write-Host "The binary '$Binary' may not exist for this platform in release $tag." -ForegroundColor Red
  exit 1
}
Write-Host "Extracting ..." -ForegroundColor Cyan
if (Get-Command tar -ErrorAction SilentlyContinue) { tar -xzf $tmp -C "$env:TEMP" }
else {
  Add-Type -AssemblyName System.IO.Compression.FileSystem
  [System.IO.Compression.TarFile]::ExtractToDirectory($tmp, $InstallDir)
}
$exePath = Join-Path $InstallDir "$Binary.exe"
$tmpBin  = Join-Path "$env:TEMP" $Binary
if (Test-Path $tmpBin) { Move-Item -Path $tmpBin -Destination $exePath -Force }
if (-not (Test-Path $exePath)) {
  Write-Host "ERROR: Binary not found after extraction." -ForegroundColor Red
  exit 1
}
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$InstallDir*") {
  [Environment]::SetEnvironmentVariable("PATH", "$userPath;$InstallDir", "User")
  Write-Host "Added $InstallDir to your PATH. Restart your terminal for it to take effect." -ForegroundColor Yellow
}
Remove-Item -Path $tmp -Force -ErrorAction SilentlyContinue
Write-Host ""
Write-Host "$Binary $tag installed to $InstallDir" -ForegroundColor Green
Write-Host "Run '$Binary' to start." -ForegroundColor Green
