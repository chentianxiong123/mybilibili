@echo off
chcp 65001 >nul
echo ========================================
echo   启动所有服务
echo ========================================
echo.

cd /d "%~dp0.."

set "JAVA_HOME=D:\Program Files (x86)\corretto-17"
set "PATH=%JAVA_HOME%\bin;%PATH%"

echo [1/11] 启动 Gateway (端口 8080) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-gateway && mvn spring-boot:run"

timeout /t 3 /nobreak >nul

echo [2/11] 启动 User (端口 8081) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-user && mvn spring-boot:run"

echo [3/11] 启动 Video (端口 8082) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-video && mvn spring-boot:run"

echo [4/11] 启动 Danmaku (端口 8083) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-danmaku && mvn spring-boot:run"

echo [5/11] 启动 Search (端口 8084) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-search && mvn spring-boot:run"

echo [6/11] 启动 Comment (端口 8085) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-comment && mvn spring-boot:run"

echo [7/11] 启动 Interaction (端口 8086) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-interaction && mvn spring-boot:run"

echo [8/11] 启动 Message (端口 8087) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-message && mvn spring-boot:run"

echo [9/11] 启动 AI (端口 8088) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-ai && mvn spring-boot:run -Dspring-boot.run.profiles=dev"

echo [10/14] 启动 Live (端口 8089) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-live && mvn spring-boot:run"

echo [11/14] 启动 Analytics (端口 8090) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-analytics && mvn spring-boot:run"

echo [12/14] 启动用户前端 (端口 5173) ...
start cmd /k "cd /d %~dp0..\mybilibili-web && pnpm run dev"

echo [13/14] 启动管理员前端 (端口 3002) ...
start cmd /k "cd /d %~dp0..\mybilibili-admin-web && pnpm run dev"

echo.
echo ========================================
echo   所有服务已启动！
echo ========================================
echo.
echo 后端服务:
echo   - Gateway:      http://localhost:8080
echo   - User:         http://localhost:8081
echo   - Video:        http://localhost:8082
echo   - Danmaku:      http://localhost:8083
echo   - Search:       http://localhost:8084
echo   - Comment:      http://localhost:8085
echo   - Interaction:  http://localhost:8086
echo   - Message:      http://localhost:8087
echo   - AI:           http://localhost:8088
echo   - Live:         http://localhost:8089
echo   - Analytics:    http://localhost:8090
echo.
echo 前端:
echo   - 用户端:       http://localhost:5173
echo   - 管理端:       http://localhost:3002
echo.
echo 注意: 需要先启动 MySQL, Redis, Nacos, RocketMQ, ElasticSearch 等基础设施
echo.
pause
