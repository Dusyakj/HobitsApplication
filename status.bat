@echo off
REM Show container status

setlocal enabledelayedexpansion

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║          HobitsApplication - Container Status                ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

docker-compose ps

echo.
echo Legend:
echo   running  = ✓ Service is active
echo   exited   = ✗ Service is stopped
echo   unhealthy = ⚠ Service has issues
echo.

pause
endlocal
