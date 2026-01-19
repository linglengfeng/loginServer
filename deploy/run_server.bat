@echo off
chcp 65001 >nul
echo ========================================
echo Starting Login Server (Foreground)
echo ========================================
echo.

REM 切到脚本所在目录，确保使用本目录下的二进制
cd /d "%~dp0"
if %errorlevel% neq 0 (
    echo Failed to change to script directory!
    pause
    exit /b 1
)

echo [1/2] Initializing database...
mysql -hlocalhost -P3306 -uroot -p123456 -e "source ..\\sql\\server.sql"
if %errorlevel% neq 0 (
    echo Database initialization failed!
    pause
    exit /b 1
)
echo Database initialization successful!
echo.

echo [2/2] Starting server (foreground)...
REM 切换到 loginServer 目录
cd /d "%~dp0.."
if %errorlevel% neq 0 (
    echo Failed to change to loginServer directory!
    pause
    exit /b 1
)

REM 前台运行服务器
go run .
if %errorlevel% neq 0 (
    echo Server exited with error code: %errorlevel%
    pause
    exit /b %errorlevel%
)

pause
