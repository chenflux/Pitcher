@echo off
setlocal enabledelayedexpansion

echo ============================================================
echo   HoneyWatch Cross-Platform Build Script
echo ============================================================

if not exist VERSION (
    echo ERROR: VERSION file not found
    exit /b 1
)

set /p VERSION= < VERSION
if "!VERSION!"=="" (
    echo ERROR: VERSION file is empty
    exit /b 1
)

set VDATE=%DATE:~0,4%%DATE:~5,2%%DATE:~8,2%

if not exist dist mkdir dist
if not exist releases mkdir releases

echo.
echo [Build Server] windows amd64
go build -ldflags "-s" -o dist\honeywatch-server-!VERSION!-windows-amd64.exe .\cmd\server\
if errorlevel 1 goto :fail

echo [Build Frontend]
if not exist web\node_modules (
    cd web & call npm install & cd ..
)
cd web
call npm run build
if errorlevel 1 goto :fail
cd ..

echo.
echo ============================================================
echo   Building Agent: all platforms
echo ============================================================

set BUILD_LOG=build_agent.log
echo. > %BUILD_LOG%

set AGENT_ARCHS=amd64 arm64 386 arm
set AGENT_PLATS=windows linux darwin

for %%p in (%AGENT_PLATS%) do (
    for %%a in (%AGENT_ARCHS%) do (
        set SKIP_BUILD=0
        if %%p==darwin if %%a==386 (
            echo [SKIP] darwin/386 - not supported
            set SKIP_BUILD=1
        )
        if %%p==windows if %%a==arm (
            echo [SKIP] windows/arm - not supported
            set SKIP_BUILD=1
        )
        if %%p==darwin if %%a==arm (
            echo [SKIP] darwin/arm - not supported
            set SKIP_BUILD=1
        )

        if !SKIP_BUILD!==1 (
            echo   [SKIP] %%p/%%a
        ) else (
            set OUTFILE=dist\honeywatch-agent-!VERSION!-%%p-%%a.exe
            if %%p==linux set OUTFILE=dist\honeywatch-agent-!VERSION!-linux-%%a
            if %%p==darwin set OUTFILE=dist\honeywatch-agent-!VERSION!-darwin-%%a

            echo [Build] %%p/%%a - !OUTFILE!

            pushd agent
            set GOOS=%%p
            set GOARCH=%%a
            go build -ldflags "-s" -o ..\!OUTFILE! .\cmd\
            set BUILD_CODE=!errorlevel!
            popd
            if !BUILD_CODE! neq 0 (
                echo   [FAIL] %%p/%%a >> ..\%BUILD_LOG%
            ) else (
                echo   [OK] !OUTFILE!
            )
        )
    )
)

echo.
echo ============================================================
echo   Generating releases manifest
echo ============================================================

echo { > releases\manifest.json
echo   "version": "!VERSION!", >> releases\manifest.json
echo   "build_date": "%VDATE%", >> releases\manifest.json
echo   "files": [ >> releases\manifest.json

set FIRST=1
for %%f in (dist\honeywatch-agent-!VERSION!-*) do (
    if not defined FIRST (
        echo     ,>> releases\manifest.json
    )
    set FIRST=

    set FN=%%~nxf
    set SIZE=0
    for %%s in ("%%f") do set SIZE=%%~zs

    for %%a in (amd64 arm64 386 arm) do (
        echo !FN! | find "-linux-%%a" >nul && (
            set OS=linux
            set ARCH=%%a
            set EXT=
        )
        echo !FN! | find "-darwin-%%a" >nul && (
            set OS=darwin
            set ARCH=%%a
            set EXT=
        )
        echo !FN! | find "-windows-%%a.exe" >nul && (
            set OS=windows
            set ARCH=%%a
            set EXT=.exe
        )
    )

    echo     { >> releases\manifest.json
    echo       "filename": "!FN!", >> releases\manifest.json
    echo       "os": "!OS!", >> releases\manifest.json
    echo       "arch": "!ARCH!", >> releases\manifest.json
    echo       "url": "/api/downloads/agent/!FN!", >> releases\manifest.json
    echo       "size": !SIZE! >> releases\manifest.json
    echo     } >> releases\manifest.json
)

echo   ] >> releases\manifest.json
echo } >> releases\manifest.json

echo.
echo ============================================================
echo   All builds complete!
echo ============================================================
echo Server:     dist\honeywatch-server-!VERSION!-windows-amd64.exe
echo Agent binaries:
dir /b dist\honeywatch-agent-!VERSION!-*
echo.
echo Releases manifest: releases\manifest.json
goto :end

:fail
echo.
echo Build failed!
exit /b 1

:end
endlocal