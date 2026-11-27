@echo off
setlocal
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

where x86_64-w64-mingw32-gcc >nul 2>&1
if %errorlevel% equ 0 (
    set CC=x86_64-w64-mingw32-gcc
    set CXX=x86_64-w64-mingw32-g++
) else (
    set CC=gcc
    set CXX=g++
)

go build -ldflags "-H windowsgui" -tags=json1 -o xbvr.exe pkg/tray/main.go
set BUILD_ERR=%errorlevel%
endlocal & exit /b %BUILD_ERR%
