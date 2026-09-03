# HotStack - Instalador para Windows
# Uso: irm https://github.com/abraa/hotstack/releases/latest/download/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "abraaosala/hotstack"
$InstallDir = if ($env:LOCALAPPDATA) { "$env:LOCALAPPDATA\HotStack" } else { "$env:USERPROFILE\.hotstack\bin" }

Write-Host "HotStack - Instalador" -ForegroundColor Cyan
Write-Host "====================="
Write-Host ""

# Detetar arquitetura
$arch = if ([System.Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Host "Arquitetura não suportada: 32-bit" -ForegroundColor Red
    exit 1
}

$platform = "windows-$arch"

# Obter versão mais recente
Write-Host "A detetar versão mais recente..."
try {
    $release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
    $version = $release.tag_name
} catch {
    Write-Host "Erro ao obter versão: $_" -ForegroundColor Red
    exit 1
}

Write-Host "Plataforma: $platform"
Write-Host "Versão:     $version"
Write-Host ""

# Descarregar
$url = "https://github.com/$Repo/releases/download/$version/hotstack-$platform.exe"
$dest = "$InstallDir\hotstack.exe"

Write-Host "A descargar de: $url"
Write-Host "Para: $dest"
Write-Host ""

try {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Invoke-WebRequest -Uri $url -OutFile $dest -UseBasicParsing
} catch {
    Write-Host "Erro ao descargar: $_" -ForegroundColor Red
    exit 1
}

# Adicionar ao PATH
$currentUserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($currentUserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$currentUserPath;$InstallDir", "User")
    Write-Host "✓ Adicionado ao PATH do utilizador" -ForegroundColor Green
    Write-Host "  Reinicia o terminal para aplicar as alterações"
}

Write-Host ""
Write-Host "✓ HotStack instalado em: $dest" -ForegroundColor Green
Write-Host ""
Write-Host "Testa com:" -ForegroundColor Yellow
Write-Host "  hotstack --help"
