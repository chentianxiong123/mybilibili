# Start all services script
$ErrorActionPreference = "Continue"

Write-Host "========================================"
Write-Host "  Starting All Services"
Write-Host "========================================"
Write-Host ""

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoDir = Split-Path -Parent $scriptDir
Set-Location $repoDir

$jdkHome = "D:\Program Files (x86)\corretto-17"

function Start-MavenService {
    param(
        [string]$Name,
        [string]$Module,
        [string]$Port,
        [string]$ExtraArgs = ""
    )

    Write-Host "Starting $Name (port $Port) ..."
    $moduleDir = Join-Path $repoDir $Module
    $cmd = "set `"JAVA_HOME=$jdkHome`" && set `"PATH=%JAVA_HOME%\bin;%PATH%`" && cd /d `"$moduleDir`" && mvn spring-boot:run $ExtraArgs"
    Start-Process cmd.exe -ArgumentList "/k", $cmd
}

function Start-NodeService {
    param(
        [string]$Name,
        [string]$Module,
        [string]$Port
    )

    Write-Host "Starting $Name (port $Port) ..."
    $moduleDir = Join-Path $repoDir $Module
    $cmd = "cd /d `"$moduleDir`" && pnpm run dev"
    Start-Process cmd.exe -ArgumentList "/k", $cmd
}

Start-MavenService "Account Social" "mybilibili-account-social" "8081"
Start-MavenService "Video Media" "mybilibili-video-media" "8082"
Start-MavenService "Search Recommend" "mybilibili-search-recommend" "8084"
Start-MavenService "Content Interaction" "mybilibili-content-interaction" "8085"
Start-MavenService "AI" "mybilibili-ai" "8088" "-Dspring-boot.run.profiles=dev"
Start-Sleep -Seconds 3
Start-MavenService "Gateway" "mybilibili-gateway" "8080"

Start-NodeService "User Frontend" "mybilibili-web" "5173"
Start-NodeService "Admin Frontend" "mybilibili-admin-web" "3002"
Start-NodeService "WAP Frontend" "mybilibili-wap" "5174"

Write-Host ""
Write-Host "========================================"
Write-Host "  Start commands issued"
Write-Host "========================================"
Write-Host ""
Write-Host "Backend Services:"
Write-Host "  - Gateway:              http://localhost:8080"
Write-Host "  - Account Social:       http://localhost:8081"
Write-Host "  - Video Media:          http://localhost:8082"
Write-Host "  - Search Recommend:     http://localhost:8084"
Write-Host "  - Content Interaction:  http://localhost:8085"
Write-Host "  - AI:                   http://localhost:8088"
Write-Host ""
Write-Host "Frontend:"
Write-Host "  - User:         http://localhost:5173"
Write-Host "  - Admin:        http://localhost:3002"
Write-Host "  - WAP:          http://localhost:5174"
