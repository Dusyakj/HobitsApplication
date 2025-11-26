@echo off
REM Show bot logs

setlocal enabledelayedexpansion

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║      HobitsBot Logs (Press Ctrl+C to exit)                   ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

docker-compose logs -f hobitsbot

endlocal
