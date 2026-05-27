@echo off
chcp 65001 >nul
echo ========================================
echo   编译所有微服务
echo ========================================
echo.

cd /d "%~dp0"

echo 正在编译 mybilibili-cloud ...
echo 耗时较长，请耐心等待
echo.

call mvn clean install -DskipTests

if %errorlevel% neq 0 (
    echo.
    echo [错误] 编译失败，请检查错误信息
    pause
    exit /b 1
)

echo.
echo ========================================
echo   编译成功！
echo ========================================
echo.
echo 所有模块 JAR 包已生成至各模块 target/ 目录
echo 现在可以运行 start-all.bat 启动服务
echo.
pause