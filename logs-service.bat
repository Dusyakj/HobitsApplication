@echo off
REM Show service logs

setlocal enabledelayedexpansion

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║    HobitsService Logs (Press Ctrl+C to exit)                 ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

docker-compose logs -f hobits-service

endlocal
