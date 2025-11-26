@echo off
REM Setup script for HobitsApplication

setlocal enabledelayedexpansion

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║         HobitsApplication - Initial Setup                    ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

REM Check if .env exists
if exist .env (
    echo [INFO] .env file already exists
    echo.
    echo To reconfigure, please edit .env manually or delete it first
    echo.
    pause
    exit /b 0
) else (
    echo [INFO] Creating .env file from .env.example...
    copy .env.example .env
    echo.
    echo ✓ .env file created successfully!
    echo.
    echo ⚠️  IMPORTANT: Please edit .env and set:
    echo    - TELEGRAM_BOT_TOKEN=your_token_here
    echo.
    echo How to get Telegram Bot Token:
    echo   1. Open Telegram
    echo   2. Find @BotFather
    echo   3. Send /newbot
    echo   4. Follow instructions
    echo   5. Copy your token
    echo.
    echo Other settings have defaults and are optional.
    echo.
    echo Next step: Edit .env file and then run start.bat
    echo.
    pause
)

endlocal
