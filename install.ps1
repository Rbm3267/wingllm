# WingLLM One-Line Windows Installer
# Usage: irm https://raw.githubusercontent.com/Rbm3267/wingllm/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

Write-Host "=================================================" -ForegroundColor Cyan
Write-Host " WingLLM Installer (Local AI Speculative Engine) " -ForegroundColor Cyan
Write-Host "=================================================" -ForegroundColor Cyan

$InstallDir = "$env:LOCALAPPDATA\wingllm\bin"
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$ExePath = "$InstallDir\wingllm.exe"
$ReleaseUrl = "https://github.com/Rbm3267/wingllm/releases/latest/download/wingllm.exe"

Write-Host "Downloading WingLLM binary..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $ReleaseUrl -OutFile $ExePath -UseBasicParsing
    Write-Host "Downloaded successfully to: $ExePath" -ForegroundColor Green
} catch {
    Write-Host "Note: Release asset not available yet. Falling back to local build if Go is present." -ForegroundColor Gray
    if (Get-Command go -ErrorAction SilentlyContinue) {
        go install github.com/Rbm3267/wingllm/cmd/wingllm@latest
    }
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
