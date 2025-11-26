@echo off
REM Start all services

setlocal enabledelayedexpansion

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║              Starting HobitsApplication                       ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

REM Check if .env exists
if not exist .env (
    echo [ERROR] .env file not found!
    echo Please run setup.bat first
    echo.
    pause
    exit /b 1
)

REM Check if docker-compose is available
docker-compose --version >nul 2>&1
if !errorlevel! neq 0 (
    echo [ERROR] Docker Compose is not installed or not in PATH
    echo Please install Docker Desktop: https://www.docker.com/products/docker-desktop
    echo.
    pause
    exit /b 1
)

echo [INFO] Starting all services with docker-compose...
echo.

docker-compose up -d

if !errorlevel! equ 0 (
    echo.
    echo ✓ All services started successfully!
    echo.
    echo 📊 Available services:
    echo    - Grafana Dashboard:  http://localhost:3000 (admin/admin)
    echo    - Prometheus:         http://localhost:9090
    echo    - RabbitMQ Manager:   http://localhost:15672 (guest/guest)
    echo    - PostgreSQL:         localhost:5432
    echo    - HobitsService:      localhost:50051 (gRPC)
    echo.
    echo 🤖 Next steps:
    echo    1. Wait 30 seconds for services to fully start
    echo    2. Open Telegram and find your bot
    echo    3. Send /start command
    echo.
    echo 📋 Useful commands:
    echo    - status.bat      : Show container status
    echo    - logs.bat        : View live logs
    echo    - stop.bat        : Stop all services
    echo.
) else (
    echo [ERROR] Failed to start services
    echo Run: docker-compose up -d
    echo To see detailed errors
)

echo.
pause
endlocal
