@echo off
setlocal

cd /d "%~dp0"

if /i "%1"=="build" (
  go build -o htmlfmt-server.exe main.go
  exit /b %errorlevel%
)

if /i "%1"=="run" (
  go run main.go
  exit /b %errorlevel%
)

if /i "%1"=="help" goto help

go build -o htmlfmt-server.exe main.go
if errorlevel 1 exit /b %errorlevel%
.\htmlfmt-server.exe
exit /b %errorlevel%

:help
echo Usage: run.bat [build^|run^|help]
echo   build - build htmlfmt-server.exe
echo   run   - run via go run main.go
echo   (no arg) build then run exe
exit /b 0
