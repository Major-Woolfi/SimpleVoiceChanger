@echo off
setlocal enabledelayedexpansion

echo ========================================
echo SimpleVoiceChanger Build Script (Windows + Android)
echo ========================================
echo.

set BUILD_DIR=C:\svc-build
set OUTPUT_DIR=%~dp0BUILD
set APP_NAME=SimpleVoiceChanger

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

echo [1/8] Checking Go installation...
go version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Go not found in PATH. Install Go 1.21+ and restart terminal.
    exit /b 1
)
echo OK

echo [2/8] Checking C compiler (gcc for CGO/Fyne)...
gcc --version >nul 2>&1
if errorlevel 1 (
    echo ERROR: gcc not found. Install MinGW-w64 (e.g. via MSYS2: pacman -S mingw-w64-x86_64-gcc)
    exit /b 1
)
echo OK

echo [3/8] Preparing ASCII build directory...
if exist "%BUILD_DIR%" rmdir /s /q "%BUILD_DIR%"
mkdir "%BUILD_DIR%"
xcopy /E /I /Q /Y "%~dp0*" "%BUILD_DIR%\" >nul
echo OK

echo [4/8] Building Windows manifest resource...
cd /d "%BUILD_DIR%"
windres app.manifest -O coff -o app.syso 2>nul
if errorlevel 1 (
    echo WARNING: windres failed, continuing without manifest
) else (
    echo OK
)

echo [5/8] Downloading dependencies...
go mod download
if errorlevel 1 (
    echo ERROR: go mod download failed
    exit /b 1
)
echo OK

echo [6/8] Tidying modules...
go mod tidy
if errorlevel 1 (
    echo ERROR: go mod tidy failed
    exit /b 1
)
echo OK

echo [7/8] Building Windows executable (this may take ~2 minutes)... 
go build -ldflags="-s -w -H=windowsgui" -o "%APP_NAME%.exe" .
if errorlevel 1 (
    echo ERROR: Windows build failed
    exit /b 1
)
echo OK

echo [8/8] Building Android APK...
set ANDROID_SDK_ROOT=%ANDROID_HOME%
if "%ANDROID_HOME%"=="" (
    echo WARNING: ANDROID_HOME not set, skipping Android build
    goto :copy_output
)

go install fyne.io/tools/cmd/fyne@latest 2>nul
if errorlevel 1 (
    echo WARNING: fyne CLI install failed, skipping Android build
    goto :copy_output
)

if exist icon.png (
    fyne package -os android --appID github.com.MajorWoolfi.SimpleVoiceChanger --icon icon.png . 2>&1
) else (
    fyne package -os android --appID github.com.MajorWoolfi.SimpleVoiceChanger --icon assets/icons/mic.svg . 2>&1
)
if errorlevel 1 (
    echo WARNING: fyne package failed, skipping Android build
) else (
    echo Android APK built
)

:copy_output
echo.
echo [Final] Copying binaries to output directory...
copy /Y "%BUILD_DIR%\%APP_NAME%.exe" "%OUTPUT_DIR%\%APP_NAME%.exe" >nul
if errorlevel 1 (
    echo ERROR: Copy exe failed
    exit /b 1
)

if exist "%BUILD_DIR%\%APP_NAME%.apk" (
    copy /Y "%BUILD_DIR%\%APP_NAME%.apk" "%OUTPUT_DIR%\%APP_NAME%.apk" >nul
    echo Android APK copied
)

echo.
echo ========================================
echo Build complete!
echo Output directory: %OUTPUT_DIR%
echo Files:
dir /b "%OUTPUT_DIR%"
echo ========================================