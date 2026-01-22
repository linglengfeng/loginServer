@echo off
chcp 65001 >nul
echo ========================================
echo Starting Login Server
echo ========================================
echo.

echo [1/3] Initializing database...
mysql -hlocalhost -P3306 -uroot -p123456 -e "source ..\\sql\\server.sql"
if %errorlevel% neq 0 (
    echo Database initialization failed!
    pause
    exit /b 1
)
echo Database initialization successful!
echo.

echo [2/3] Running go mod tidy...
cd ..
go mod tidy
if %errorlevel% neq 0 (
    echo go mod tidy failed!
    pause
    exit /b 1
)
echo go mod tidy successful!
echo.

echo [3/3] Starting server...
go run . shell
if %errorlevel% neq 0 (
    echo Server startup failed!
    pause
    exit /b 1
)

pause
