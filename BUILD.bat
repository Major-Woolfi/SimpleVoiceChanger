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
    echo ERROR: gcc not found. Install MinGW-w64 via MSYS2: pacman -S mingw-w64-x86_64-gcc
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
set "PORTAUDIO_DIR=C:\svc-portaudio"
set "PORTAUDIO_TAG="
set "PORTAUDIO_DLL="
if exist "%PORTAUDIO_DIR%\lib\pkgconfig\portaudio-2.0.pc" (
    set "PKG_CONFIG_PATH=%PORTAUDIO_DIR%\lib\pkgconfig;%PKG_CONFIG_PATH%"
    set "PATH=%PORTAUDIO_DIR%\bin;%PATH%"
)
pkg-config --exists portaudio-2.0 2>nul
if not errorlevel 1 (
    set "PORTAUDIO_TAG=-tags portaudio"
    set "PORTAUDIO_DLL=%PORTAUDIO_DIR%\bin\libportaudio.dll"
    echo portaudio found: enabling real audio backend
) else (
    echo WARNING: pkg-config/portaudio not found, building with mock audio backend
    echo          Install pkg-config + PortAudio to enable real microphone/playback
)
go build %PORTAUDIO_TAG% -ldflags="-s -w -H=windowsgui" -o "%APP_NAME%.exe" .
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

if "%ANDROID_NDK_HOME%"=="" (
    for /f "delims=" %%d in ('dir /b /ad /o-n "%ANDROID_HOME%\ndk" 2^>nul') do (
        if not defined ANDROID_NDK_HOME set "ANDROID_NDK_HOME=%ANDROID_HOME%\ndk\%%d"
    )
)
if not "%ANDROID_NDK_HOME%"=="" echo Using NDK: %ANDROID_NDK_HOME%

go install fyne.io/tools/cmd/fyne@latest 2>nul
if errorlevel 1 (
    echo WARNING: fyne CLI install failed, skipping Android build
    goto :copy_output
)

if exist icon.png (
    fyne package -os android --appID com.majorwoolfi.simplevoicechanger --icon icon.png . 2>&1
) else (
    fyne package -os android --appID com.majorwoolfi.simplevoicechanger --icon assets/icons/mic.svg . 2>&1
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

if not "%PORTAUDIO_DLL%"=="" (
    if exist "%PORTAUDIO_DLL%" (
        copy /Y "%PORTAUDIO_DLL%" "%OUTPUT_DIR%\libportaudio.dll" >nul
        echo libportaudio.dll copied ^(required for real audio^)
    )
)

echo.
echo ========================================
echo Build complete!
echo Output directory: %OUTPUT_DIR%
echo Files:
dir /b "%OUTPUT_DIR%"
echo ========================================