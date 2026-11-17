# WingLLM One-Line Windows Installer
# Usage: irm https://huggingface.co/bennett-xdlabs/wingllm/raw/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

Write-Host "=================================================" -ForegroundColor Cyan
Write-Host " WingLLM Installer (Local AI Speculative Engine) " -ForegroundColor Cyan
Write-Host "=================================================" -ForegroundColor Cyan

$InstallDir = "$env:LOCALAPPDATA\wingllm\bin"
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$ExePath = "$InstallDir\wingllm.exe"
$ReleaseUrl = "https://huggingface.co/bennett-xdlabs/wingllm/resolve/main/wingllm.exe"

Write-Host "Downloading WingLLM binary from Hugging Face..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $ReleaseUrl -OutFile $ExePath -UseBasicParsing
    Write-Host "Downloaded successfully to: $ExePath" -ForegroundColor Green
} catch {
    Write-Host "Error downloading binary: $_" -ForegroundColor Red
    exit 1
}

# Add to user PATH if not present
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "Adding $InstallDir to User PATH..." -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:PATH = "$InstallDir;$env:PATH"
}

Write-Host "`n WingLLM installed successfully!" -ForegroundColor Green
Write-Host "Run 'wingllm list' to inspect your local models." -ForegroundColor Cyan
Write-Host "Run 'wingllm serve <model>' to start your 2x local AI server.`n" -ForegroundColor Cyan
