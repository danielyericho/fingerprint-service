@echo off
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo Requesting administrator privileges...
    powershell -NoProfile -ExecutionPolicy Bypass -Command "Start-Process -FilePath '%~f0' -Verb runAs -Wait"
    exit /b
)

set SERVICE_NAME=FingerprintQubuService

set SCRIPT_DIR=%~dp0
if "%SCRIPT_DIR:~-1%"=="\" set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%
set NSSM=%SCRIPT_DIR%\nssm.exe

sc.exe query "%SERVICE_NAME%" | findstr /I "RUNNING" >nul 2>&1
if %errorlevel% equ 0 (
    echo Service "%SERVICE_NAME%" is already running.
    goto :success
)

echo Starting service "%SERVICE_NAME%"...

if exist "%NSSM%" (
    "%NSSM%" start "%SERVICE_NAME%" >nul 2>&1
) else (
    sc.exe start "%SERVICE_NAME%" >nul 2>&1
)

sc.exe query "%SERVICE_NAME%" | findstr /I "RUNNING" >nul 2>&1
if %errorlevel% neq 0 (
    echo Failed to start service "%SERVICE_NAME%".
    echo Check "%SCRIPT_DIR%\logs\fingerprint.err.log".
    goto :fail
)

echo Service "%SERVICE_NAME%" started successfully.

goto :success

:success
echo.
pause
exit /b 0

:fail
echo.
pause
exit /b 1