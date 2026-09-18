@echo off
echo ============================================
echo   SimpleVoiceChanger - Build Script
echo ============================================
echo.

echo [1/3] Checking Go installation...
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Go is not installed or not in PATH.
    echo Please install Go 1.23 or later.
    pause
    exit /b 1
)

echo [2/3] Building Windows executable...
if not exist "BUILD" mkdir BUILD
go build -ldflags="-s -w" -o "BUILD\SimpleVoiceChanger.exe" .
if %errorlevel% neq 0 (
    echo ERROR: Build failed.
    pause
    exit /b 1
)
echo Build Windows: OK

echo [3/3] Build complete!
echo.
echo Output files:
echo   BUILD\SimpleVoiceChanger.exe
echo.
echo NOTE: Android APK requires Android SDK and gomobile.
echo Run: go install mobile/... && gomobile bind -target android
echo.
pause
