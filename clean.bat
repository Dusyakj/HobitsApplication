@echo off
REM Clean up all containers and volumes

setlocal enabledelayedexpansion

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║     HobitsApplication - Full Cleanup                          ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

echo ⚠️  WARNING: This will remove ALL containers, volumes, and data!
echo.
echo Press Y to confirm, or any other key to cancel
set /p choice=Your choice (Y/N):

if /i "%choice%"=="Y" (
    echo.
    echo [INFO] Removing all containers and volumes...
    docker-compose down -v

    if !errorlevel! equ 0 (
        echo.
        echo ✓ Cleanup completed successfully!
        echo.
        echo All containers and data volumes have been removed.
        echo.
    ) else (
        echo [ERROR] Cleanup failed
    )
) else (
    echo [INFO] Cleanup cancelled
)

echo.
pause
endlocal
