#!/usr/bin/env pwsh
$Repo = "N1XNAC/nix-code"
$Binary = "n1x"

Write-Host "Installing N1X Code - Terminal AI Coding Agent" -ForegroundColor Blue

$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$Url = "https://github.com/$Repo/releases/latest/download/${Binary}_windows_${Arch}.tar.gz"

$InstallDir = "$env:USERPROFILE\.n1x\bin"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

Write-Host "Downloading from: $Url"
$TmpFile = [System.IO.Path]::GetTempFileName()
Invoke-WebRequest -Uri $Url -OutFile $TmpFile

tar -xzf $TmpFile -C $InstallDir --strip-components=1
Remove-Item $TmpFile

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
  [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
  $env:Path = "$env:Path;$InstallDir"
}

Write-Host "N1X Code installed!" -ForegroundColor Green
Write-Host "Run 'n1x config' to set up your API keys"
Write-Host "Run 'n1x run ""your prompt""' to start coding"
