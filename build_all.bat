@echo off
setlocal

rem Ensure toolchain directories are in PATH for the current session
set "MINGW_BIN=%LOCALAPPDATA%\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin"
set "PATH=%LOCALAPPDATA%\Programs\Go\bin;%MINGW_BIN%;%LOCALAPPDATA%\Microsoft\WinGet\Links;%APPDATA%\npm;%PATH%"

if not exist "node_modules" (
    echo [1/3] Installing Node.js dependencies...
    call yarn install --ignore-engines
    if errorlevel 1 (
        echo Failed to install Node.js dependencies.
        exit /b %errorlevel%
    )
) else (
    echo [1/3] node_modules already exists, skipping yarn install.
)

echo [2/3] Building frontend (vbuild.bat)...
call vbuild.bat
if errorlevel 1 (
    echo Frontend build failed.
    exit /b %errorlevel%
)

echo [3/3] Building xbvr.exe (build.bat)...
call build.bat
if errorlevel 1 (
    echo Go build failed.
    exit /b %errorlevel%
)

echo Build completed successfully: xbvr.exe
endlocal
