@echo off
setlocal enabledelayedexpansion
set "PROJECT_DIR=%~dp0"

echo ============================================
echo   SimpleVoiceChanger - Build Script
echo ============================================
echo.

echo [1/8] Checking Go installation...
where go >nul 2>nul
set GO_ERR=!errorlevel!
if !GO_ERR! neq 0 (
    echo ERROR: Go is not installed or not in PATH.
    echo Please install Go 1.23 or later.
    pause
    exit /b 1
)
echo Go found.

echo [2/8] Checking C compiler - needed for CGO/Fyne...
where gcc >nul 2>nul
set GCC_ERR=!errorlevel!
set GCC_PATH=
if !GCC_ERR! equ 0 (
    for /f "delims=" %%g in ('where gcc') do set GCC_PATH=%%g
)
if not defined GCC_PATH (
    if exist "C:\TDM-GCC-64\bin\gcc.exe" (
        set GCC_PATH=C:\TDM-GCC-64\bin\gcc.exe
    ) else if exist "C:\TDM-GCC-32\bin\gcc.exe" (
        set GCC_PATH=C:\TDM-GCC-32\bin\gcc.exe
    ) else if exist "C:\Program Files\Git\mingw64\bin\gcc.exe" (
        set GCC_PATH=C:\Program Files\Git\mingw64\bin\gcc.exe
    )
)
if not defined GCC_PATH (
    echo ERROR: gcc MinGW not found.
    echo Install TDM-GCC from https://sourceforge.net/projects/tdm-gcc/
    echo or Git for Windows and add it to PATH.
    pause
    exit /b 1
)
echo C compiler: !GCC_PATH!

echo [3/8] Preparing ASCII build directory...
set "BUILD_DIR=C:\svc-build"
if exist "!BUILD_DIR!" rmdir /s /q "!BUILD_DIR!" 2>nul
mkdir "!BUILD_DIR!"

echo Copying source files...
powershell -Command "$dirs = @('audio','core','effects','gui','i18n','internal','tests','assets','local'); foreach ($d in $dirs) { if (Test-Path ($env:PROJECT_DIR + $d)) { Copy-Item -Path ($env:PROJECT_DIR + $d) -Destination ($env:BUILD_DIR + '\' + $d) -Recurse -Force } }"
powershell -Command "Copy-Item -Path ($env:PROJECT_DIR + 'go.mod') -Destination ($env:BUILD_DIR + '\go.mod') -Force"
powershell -Command "Copy-Item -Path ($env:PROJECT_DIR + 'go.sum') -Destination ($env:BUILD_DIR + '\go.sum') -Force"
powershell -Command "Copy-Item -Path ($env:PROJECT_DIR + 'main.go') -Destination ($env:BUILD_DIR + '\main.go') -Force"
powershell -Command "Copy-Item -Path ($env:PROJECT_DIR + 'main_android.go') -Destination ($env:BUILD_DIR + '\main_android.go') -Force"
powershell -Command "Copy-Item -Path ($env:PROJECT_DIR + 'README.md') -Destination ($env:BUILD_DIR + '\README.md') -Force"
powershell -Command "Copy-Item -Path ($env:PROJECT_DIR + 'LICENSE') -Destination ($env:BUILD_DIR + '\LICENSE') -Force"

mkdir "!BUILD_DIR!\manifest"
echo ^<?xml version="1.0" encoding="UTF-8" standalone="yes"?^> > "!BUILD_DIR!\manifest\app.manifest"
echo ^<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0"^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^<trustInfo xmlns="urn:schemas-microsoft-com:asm.v3"^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^<security^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^<requestedPrivileges^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^<requestedExecutionLevel level="asInvoker" uiAccess="false"/^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^</requestedPrivileges^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^</security^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^</trustInfo^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^<compatibility xmlns="urn:schemas-microsoft-com:compatibility.v1"^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^<application^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^<supportedOS Id="{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}"/^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^</application^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^</compatibility^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^<application xmlns="urn:schemas-microsoft-com:asm.v3"^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^<windowsSettings^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^<dpiAware xmlns="http://schemas.microsoft.com/SMI/2005/WindowsSettings"^>true/pm^</dpiAware^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^<dpiAwareness xmlns="http://schemas.microsoft.com/SMI/2016/WindowsSettings"^>PerMonitorV2^</dpiAwareness^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^</windowsSettings^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^</application^> >> "!BUILD_DIR!\manifest\app.manifest"
echo ^</assembly^> >> "!BUILD_DIR!\manifest\app.manifest"
echo CREATEPROCESS_MANIFEST_RESOURCE_ID RT_MANIFEST "app.manifest" > "!BUILD_DIR!\manifest\app.rc"

echo Build directory prepared at !BUILD_DIR!

echo [4/8] Building manifest resource...
if exist "!BUILD_DIR!\app.syso" del /q "!BUILD_DIR!\app.syso" 2>nul
"C:\TDM-GCC-64\bin\windres.exe" -i "!BUILD_DIR!\manifest\app.rc" -o "!BUILD_DIR!\app.syso" -O coff
    set WINDRES_ERR=!errorlevel!
    if !WINDRES_ERR! neq 0 (
        echo WARNING: windres failed - manifest will not be embedded
    ) else (
        echo Manifest resource built: OK
    )

echo [5/8] Updating dependencies...
cd /d "!BUILD_DIR!"
set "GOMODCACHE=!BUILD_DIR!\modcache"
set "GOPATH=!BUILD_DIR!\gopath"
set CGO_ENABLED=1
set "CC=!GCC_PATH!"
go get -u ./...
set GET_ERR=!errorlevel!
if !GET_ERR! neq 0 (
    echo WARNING: go get failed - continuing with existing dependencies
)
echo Dependencies updated.

echo [6/8] Building Windows executable...
if exist SimpleVoiceChanger.exe del /q SimpleVoiceChanger.exe 2>nul
go build -ldflags="-H windowsgui" -v -o SimpleVoiceChanger.exe .
set BUILD_ERR=!errorlevel!
if !BUILD_ERR! neq 0 (
    echo.
    echo ERROR: Windows EXE build failed with error !BUILD_ERR!
    pause
    exit /b 1
)
echo Build Windows EXE: OK

echo [7/8] Building Android APK - optional...
where gomobile >nul 2>nul
set GOMO_FIND=!errorlevel!
if !GOMO_FIND! neq 0 (
    echo gomobile not found - APK build skipped
) else (
    set "GOMO_BIN=%USERPROFILE%\go\bin\gomobile.exe"
    if not exist "!GOMO_BIN!" set "GOMO_BIN=%USERPROFILE%\go\bin\gomobile"
    if not exist "!GOMO_BIN!" (
        echo gomobile binary not found - APK build skipped
    ) else (
        set "ANDROID_SDK="
        if defined ANDROID_HOME set "ANDROID_SDK=!ANDROID_HOME!"
        if defined ANDROID_SDK_ROOT set "ANDROID_SDK=!ANDROID_SDK_ROOT!"
        if not defined ANDROID_SDK (
            if exist "!USERPROFILE!\AppData\Local\Android\Sdk" set "ANDROID_SDK=!USERPROFILE!\AppData\Local\Android\Sdk"
        )
        if not defined ANDROID_SDK (
            echo Android SDK not found - APK build skipped
        ) else (
            echo Android SDK: !ANDROID_SDK!
            set "ANDROID_NDK="
            if defined ANDROID_NDK_HOME set "ANDROID_NDK=!ANDROID_NDK_HOME!"
            if not defined ANDROID_NDK (
                if exist "C:\ndk\25.0.8775105" set "ANDROID_NDK=C:\ndk\25.0.8775105"
            )
            if defined ANDROID_NDK (
                echo Android NDK: !ANDROID_NDK!
                if exist "!BUILD_DIR!\app.syso" del /q "!BUILD_DIR!\app.syso" 2>nul
                set "GOTMPDIR=!BUILD_DIR!\gomobile-tmp"
                mkdir "!BUILD_DIR!\gomobile-tmp" 2>nul
                cd /d "!BUILD_DIR!"
                set "CGO_ENABLED=1"
                set "CC=!ANDROID_NDK!\toolchains\llvm\prebuilt\windows-x86_64\bin\clang.exe"
                set "CXX=!ANDROID_NDK!\toolchains\llvm\prebuilt\windows-x86_64\bin\clang++.exe"
                set "ANDROID_HOME=!ANDROID_SDK!"
                set "ANDROID_NDK_HOME=!ANDROID_NDK!"
                set "PATH=!ANDROID_NDK!\toolchains\llvm\prebuilt\windows-x86_64\bin;!PATH!"
                "!GOMO_BIN!" init
                if !errorlevel! equ 0 (
                    if exist SimpleVoiceChanger.apk del /q SimpleVoiceChanger.apk 2>nul
                    "!GOMO_BIN!" build -target android/arm64 -androidapi 30 -v -o SimpleVoiceChanger.apk .
                    set APK_ERR=!errorlevel!
                    if !APK_ERR! neq 0 (
                        echo WARNING: Android APK build failed
                    ) else (
                        echo Build Android APK: OK
                    )
                ) else (
                    echo WARNING: gomobile init failed - APK build skipped
                )
            ) else (
                echo Android NDK not found - APK build skipped
            )
        )
    )
)

echo [8/8] Copying result and cleaning up...
powershell -Command "if (!(Test-Path ($env:PROJECT_DIR + '\BUILD'))) { New-Item -Path ($env:PROJECT_DIR + '\BUILD') -ItemType Directory -Force }"
powershell -Command "if (Test-Path ($env:PROJECT_DIR + '\BUILD\SimpleVoiceChanger.exe')) { Remove-Item ($env:PROJECT_DIR + '\BUILD\SimpleVoiceChanger.exe') }"
powershell -Command "Copy-Item -Path ($env:BUILD_DIR + '\SimpleVoiceChanger.exe') -Destination ($env:PROJECT_DIR + '\BUILD\SimpleVoiceChanger.exe') -Force"
if exist "!BUILD_DIR!\SimpleVoiceChanger.apk" (
    powershell -Command "if (Test-Path ($env:PROJECT_DIR + '\BUILD\SimpleVoiceChanger.apk')) { Remove-Item ($env:PROJECT_DIR + '\BUILD\SimpleVoiceChanger.apk') }"
    powershell -Command "Copy-Item -Path ($env:BUILD_DIR + '\SimpleVoiceChanger.apk') -Destination ($env:PROJECT_DIR + '\BUILD\SimpleVoiceChanger.apk') -Force"
)
rmdir /s /q "!BUILD_DIR!" 2>nul
powershell -command "Get-Item '!PROJECT_DIR!\BUILD\SimpleVoiceChanger.exe' -Stream Zone.Identifier -ErrorAction SilentlyContinue | Remove-Item" >nul 2>&1
echo Temp directory cleaned.
echo Output: BUILD\SimpleVoiceChanger.exe
echo.
echo ============================================
echo   Build complete!
echo ============================================
echo Windows: BUILD\SimpleVoiceChanger.exe
echo Android: BUILD\SimpleVoiceChanger.apk (if built)
echo Note: Embedded Windows manifest for compatibility
echo Note: Build performed in ASCII path (C:\svc-build) for Cyrillic safety
echo.
pause
