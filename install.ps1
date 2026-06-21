#!/usr/bin/env pwsh
$Repo = "n1xcode/n1x"
$Binary = "n1x"

Write-Host "Installing N1X Code - Terminal AI Coding Agent" -ForegroundColor Blue

$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$Url = "https://github.com/$Repo/releases/latest/download/${Binary}_windows_${Arch}.tar.gz"

$InstallDir = "$env:USERPROFILE\.n1x\bin"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

Write-Host "Downloading from: $Url"
$TmpFile = [System.IO.Path]::GetTempFileName()
Invoke-WebRequest -Uri $Url -OutFile $TmpFile

tar -xzf $TmpFile -C $InstallDir
Remove-Item $TmpFile

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
  [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
  $env:Path = "$env:Path;$InstallDir"
}

Write-Host "N1X Code installed!" -ForegroundColor Green
Write-Host "Run 'nix config' to set up your API keys"
Write-Host "Run 'nix run ""your prompt""' to start coding"
