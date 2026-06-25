@echo off
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo Requesting administrator privileges...
    powershell -Command "Start-Process '%~f0' -Verb runAs"
    exit /b
)

set SERVICE_NAME=FingerprintQubuService

set SCRIPT_DIR=%~dp0
if "%SCRIPT_DIR:~-1%"=="\" set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%
set NSSM=%SCRIPT_DIR%\nssm.exe

echo Starting service "%SERVICE_NAME%"...

if exist "%NSSM%" (
    "%NSSM%" start "%SERVICE_NAME%" >nul 2>&1
) else (
    sc.exe start "%SERVICE_NAME%" >nul 2>&1
)

if %errorlevel% neq 0 (
    echo Failed to start service "%SERVICE_NAME%". Check "%SCRIPT_DIR%\logs\fingerprint.err.log".
    timeout /t 3 >nul
    exit /b
)

echo Service "%SERVICE_NAME%" started successfully.
exit /b
