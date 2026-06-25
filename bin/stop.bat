@echo off
net session >nul 2>&1
if %errorlevel% neq 0 (
    powershell -Command "Start-Process '%~f0' -Verb runAs"
    exit /b
)

set SERVICE_NAME=FingerprintQubuService

set SCRIPT_DIR=%~dp0
if "%SCRIPT_DIR:~-1%"=="\" set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%
set NSSM=%SCRIPT_DIR%\nssm.exe

sc.exe query "%SERVICE_NAME%" >nul 2>&1
if %errorlevel% neq 0 exit /b 0

if exist "%NSSM%" (
    "%NSSM%" stop "%SERVICE_NAME%" >nul 2>&1
) else (
    sc.exe stop "%SERVICE_NAME%" >nul 2>&1
)

timeout /t 2 >nul
exit /b 0
