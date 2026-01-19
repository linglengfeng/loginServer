@echo off
REM Build backend binary for Linux (amd64)

REM Switch to repo root (one level up from deploy)
cd /d "%~dp0.."

REM Set Go env
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

REM Download deps
go mod download
IF %ERRORLEVEL% NEQ 0 (
    echo go mod download failed.
    exit /b 1
)

REM Clean old binary before build (if exists)
if exist "%~dp0loginServer" (
    echo Removing old binary: %~dp0loginServer
    del /f /q "%~dp0loginServer"
    if %ERRORLEVEL% NEQ 0 (
        echo Warning: Failed to remove old binary, continuing anyway...
    )
)

REM Build (output to script dir -> deploy/loginServer)
go build -o "%~dp0loginServer" .
IF %ERRORLEVEL% NEQ 0 (
    echo Build failed.
    exit /b 1
)

echo Build success. Output: %~dp0loginServer (Linux amd64)