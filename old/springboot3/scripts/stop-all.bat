@echo off
chcp 65001 >nul
echo ========================================
echo   停止 MyBilibili 服务
echo ========================================
echo.

setlocal enabledelayedexpansion

set PORTS=8080 8081 8082 8083 8084 8085 8086 8087 8088 8089 5173 5174 3002
set FOUND=0

for %%p in (%PORTS%) do (
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%%p " ^| findstr "LISTENING" 2^>nul') do (
        echo 停止端口 %%p - PID: %%a
        taskkill /f /pid %%a >nul 2>&1 && set FOUND=1
    )
)

echo.
if %FOUND% equ 1 (
    echo ========================================
    echo   MyBilibili 服务已全部停止
    echo ========================================
) else (
    echo 未检测到 MyBilibili 相关服务在运行
)
echo.
pause