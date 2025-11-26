@echo off
REM Stop all services

setlocal enabledelayedexpansion

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║              Stopping HobitsApplication                       ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

echo [INFO] Stopping all services...
docker-compose down

if !errorlevel! equ 0 (
    echo.
    echo ✓ All services stopped successfully!
    echo.
    echo 📋 To remove all data and containers:
    echo    - Run: docker-compose down -v
    echo.
) else (
    echo [ERROR] Failed to stop services
)

echo.
pause
endlocal
