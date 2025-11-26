@echo off
REM Show live logs

setlocal enabledelayedexpansion

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║        HobitsApplication - Live Logs (Press Ctrl+C to exit)  ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

docker-compose logs -f

endlocal
