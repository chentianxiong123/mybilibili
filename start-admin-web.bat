@echo off
chcp 65001 >nul
echo ==========================================
echo    MyBilibili Admin Web - Start Script
echo ==========================================
echo.

cd /d "%~dp0mybilibili-admin-web"

echo [1/2] Checking dependencies...
if not exist "node_modules" (
    echo [Info] node_modules not found, installing dependencies...
    echo.
    call npm install
    if errorlevel 1 (
        echo [Error] Failed to install dependencies!
        pause
        exit /b 1
    )
    echo [Success] Dependencies installed!
) else (
    echo [Success] Dependencies exist
)

echo.
echo [2/2] Starting Admin Web dev server...
echo.
echo ------------------------------------------
echo  Project: mybilibili-admin-web
echo  URL: http://localhost:5173
echo  Press Ctrl+C to stop
echo ------------------------------------------
echo.

npm run dev

pause
