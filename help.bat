@echo off
REM Show help

setlocal enabledelayedexpansion

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║          HobitsApplication - Quick Start Guide                ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

echo 📋 QUICK START (3 STEPS)
echo ═══════════════════════════════════════════════════════════════
echo.
echo Step 1: Setup (first time only)
echo   ► Double-click: setup.bat
echo     OR run in PowerShell:
echo     copy .env.example .env
echo.
echo Step 2: Edit .env file
echo   ► Open .env with notepad
echo   ► Set TELEGRAM_BOT_TOKEN=your_token_here
echo   ► Save and close
echo.
echo Step 3: Start services
echo   ► Double-click: start.bat
echo     OR run in PowerShell:
echo     docker-compose up -d
echo.

echo.
echo 🔧 AVAILABLE COMMANDS
echo ═══════════════════════════════════════════════════════════════
echo.
echo setup.bat           - Create .env file (first time only)
echo start.bat           - Start all services
echo stop.bat            - Stop all services
echo status.bat          - Show container status
echo logs.bat            - View all logs (live)
echo logs-bot.bat        - View bot logs (live)
echo logs-service.bat    - View backend logs (live)
echo clean.bat           - Remove all containers and data (⚠️ be careful!)
echo help.bat            - Show this help (current)
echo.

echo.
echo 📊 SERVICES AFTER START
echo ═══════════════════════════════════════════════════════════════
echo.
echo Grafana Dashboard   http://localhost:3000     admin / admin
echo Prometheus          http://localhost:9090
echo RabbitMQ Manager    http://localhost:15672    guest / guest
echo PostgreSQL          localhost:5432
echo HobitsService gRPC  localhost:50051
echo Backend REST API    http://localhost:8080
echo.

echo.
echo 🤖 TEST THE BOT
echo ═══════════════════════════════════════════════════════════════
echo.
echo 1. After start.bat finishes, wait 30 seconds
echo 2. Open Telegram
echo 3. Find your bot (by name from BotFather)
echo 4. Send: /start
echo.

echo.
echo 📚 MORE DOCUMENTATION
echo ═════════════════════════════════════════════════════════════════
echo.
echo README.md           - Full project documentation
echo QUICKSTART.md       - Quick start guide
echo DEPLOYMENT.md       - Production deployment
echo PROJECT_STRUCTURE.md - Project structure overview
echo.

echo.
echo 🆘 TROUBLESHOOTING
echo ═══════════════════════════════════════════════════════════════
echo.
echo Docker not installed?
echo   → Install Docker Desktop: https://www.docker.com/products/docker-desktop
echo.
echo Bot token not set?
echo   → Edit .env and set TELEGRAM_BOT_TOKEN
echo   → Get token from @BotFather in Telegram
echo.
echo Services won't start?
echo   → Run: status.bat
echo   → Check logs: logs.bat
echo   → Ensure Docker is running
echo.
echo Need to restart?
echo   → Run: stop.bat
echo   → Then: start.bat
echo.

echo.
pause
endlocal
