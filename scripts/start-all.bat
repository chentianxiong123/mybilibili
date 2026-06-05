@echo off
chcp 65001 >nul
echo ========================================
echo   启动所有服务
echo ========================================
echo.

cd /d "%~dp0.."

set "JAVA_HOME=D:\Program Files (x86)\corretto-17"
set "PATH=%JAVA_HOME%\bin;%PATH%"

echo [1/9] 启动 Account Social (端口 8081) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-account-social && mvn spring-boot:run"

echo [2/9] 启动 Video Media (端口 8082) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-video-media && mvn spring-boot:run"

echo [3/9] 启动 Search Recommend (端口 8084) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-search-recommend && mvn spring-boot:run"

echo [4/9] 启动 Content Interaction (端口 8085) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-content-interaction && mvn spring-boot:run"

echo [5/9] 启动 AI (端口 8088) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-ai && mvn spring-boot:run -Dspring-boot.run.profiles=dev"

timeout /t 3 /nobreak >nul

echo [6/9] 启动 Gateway (端口 8080) ...
start cmd /k "set JAVA_HOME=D:\Program Files (x86)\corretto-17 && set PATH=%JAVA_HOME%\bin;%PATH% && cd /d %~dp0..\mybilibili-gateway && mvn spring-boot:run"

echo [7/9] 启动用户前端 (端口 5173) ...
start cmd /k "cd /d %~dp0..\mybilibili-web && pnpm run dev"

echo [8/9] 启动管理员前端 (端口 3002) ...
start cmd /k "cd /d %~dp0..\mybilibili-admin-web && pnpm run dev"

echo [9/9] 启动移动端前端 (端口 5174) ...
start cmd /k "cd /d %~dp0..\mybilibili-wap && pnpm run dev"

echo.
echo ========================================
echo   所有服务已启动！
echo ========================================
echo.
echo 后端服务:
echo   - Gateway:              http://localhost:8080
echo   - Account Social:       http://localhost:8081
echo   - Video Media:          http://localhost:8082
echo   - Search Recommend:     http://localhost:8084
echo   - Content Interaction:  http://localhost:8085
echo   - AI:                   http://localhost:8088
echo.
echo 前端:
echo   - 用户端:       http://localhost:5173
echo   - 管理端:       http://localhost:3002
echo   - 移动端:       http://localhost:5174
echo.
echo 注意: 需要先启动 MySQL, Redis, Nacos, RocketMQ, ElasticSearch 等基础设施
echo.
pause
