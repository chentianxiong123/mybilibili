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

Start-MavenService "Gateway" "mybilibili-gateway" "8080"
Start-Sleep -Seconds 3
Start-MavenService "User" "mybilibili-user" "8081"
Start-MavenService "Video" "mybilibili-video" "8082"
Start-MavenService "Danmaku" "mybilibili-danmaku" "8083"
Start-MavenService "Search" "mybilibili-search" "8084"
Start-MavenService "Comment" "mybilibili-comment" "8085"
Start-MavenService "Interaction" "mybilibili-interaction" "8086"
Start-MavenService "Message" "mybilibili-message" "8087"
Start-MavenService "AI" "mybilibili-ai" "8088" "-Dspring-boot.run.profiles=dev"
Start-MavenService "Live" "mybilibili-live" "8089"
Start-MavenService "Analytics" "mybilibili-analytics" "8090"

Start-NodeService "User Frontend" "mybilibili-web" "5173"
Start-NodeService "Admin Frontend" "mybilibili-admin-web" "3002"
Start-NodeService "WAP Frontend" "mybilibili-wap" "5174"

Write-Host ""
Write-Host "========================================"
Write-Host "  Start commands issued"
Write-Host "========================================"
Write-Host ""
Write-Host "Backend Services:"
Write-Host "  - Gateway:      http://localhost:8080"
Write-Host "  - User:         http://localhost:8081"
Write-Host "  - Video:        http://localhost:8082"
Write-Host "  - Danmaku:      http://localhost:8083"
Write-Host "  - Search:       http://localhost:8084"
Write-Host "  - Comment:      http://localhost:8085"
Write-Host "  - Interaction:  http://localhost:8086"
Write-Host "  - Message:      http://localhost:8087"
Write-Host "  - AI:           http://localhost:8088"
Write-Host "  - Live:         http://localhost:8089"
Write-Host "  - Analytics:    http://localhost:8090"
Write-Host ""
Write-Host "Frontend:"
Write-Host "  - User:         http://localhost:5173"
Write-Host "  - Admin:        http://localhost:3002"
Write-Host "  - WAP:          http://localhost:5174"
