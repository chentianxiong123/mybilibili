# Start all services script
$ErrorActionPreference = "Continue"

Write-Host "========================================"
Write-Host "  Starting All Services"
Write-Host "========================================"
Write-Host ""

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

Write-Host "[1/11] Starting Gateway (port 8080) ..."
Start-Process powershell -ArgumentList "-NoExit -Command `"cd '$scriptDir\mybilibili-gateway'; mvn spring-boot:run`""

Start-Sleep -Seconds 3

Write-Host "[2/11] Starting User (port 8081) ..."
Start-Process powershell -ArgumentList "-NoExit -Command `"cd '$scriptDir\mybilibili-user'; mvn spring-boot:run`""

Write-Host "[3/11] Starting Video (port 8082) ..."
Start-Process powershell -ArgumentList "-NoExit -Command `"cd '$scriptDir\mybilibili-video'; mvn spring-boot:run`""

Write-Host "[4/11] Starting Danmaku (port 8083) ..."
Start-Process powershell -ArgumentList "-NoExit -Command `"cd '$scriptDir\mybilibili-danmaku'; mvn spring-boot:run`""

Write-Host "[5/11] Starting Search (port 8084) ..."
Start-Process powershell -ArgumentList "-NoExit -Command `"cd '$scriptDir\mybilibili-search'; mvn spring-boot:run`""

Write-Host "[6/11] Starting Comment (port 8085) ..."
Start-Process powershell -ArgumentList "-NoExit -Command `"cd '$scriptDir\mybilibili-comment'; mvn spring-boot:run`""

Write-Host "[7/11] Starting Interaction (port 8086) ..."
Start-Process powershell -ArgumentList "-NoExit -Command `"cd '$scriptDir\mybilibili-interaction'; mvn spring-boot:run`""

Write-Host "[8/11] Starting Message (port 8087) ..."
Start-Process powershell -ArgumentList "-NoExit -Command `"cd '$scriptDir\mybilibili-message'; mvn spring-boot:run`""

Write-Host "[9/11] Starting AI (port 8088) ..."
Start-Process powershell -ArgumentList "-NoExit -Command `"cd '$scriptDir\mybilibili-ai'; mvn spring-boot:run -Dspring-boot.run.profiles=dev`""

Write-Host "[10/11] Starting User Frontend (port 5173) ..."
Start-Process powershell -ArgumentList "-NoExit -Command `"cd '$scriptDir\mybilibili-web'; npm run dev`""

Write-Host "[11/11] Starting Admin Frontend (port 3002) ..."
Start-Process powershell -ArgumentList "-NoExit -Command `"cd '$scriptDir\mybilibili-admin-web'; npm run dev`""

Write-Host ""
Write-Host "========================================"
Write-Host "  All Services Started!"
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
Write-Host ""
Write-Host "Frontend:"
Write-Host "  - User:         http://localhost:5173"
Write-Host "  - Admin:        http://localhost:3002"
Write-Host ""
Write-Host "Note: Make sure MySQL, Redis, Nacos, RocketMQ, ElasticSearch are running first"
Write-Host ""
