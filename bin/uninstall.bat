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
    timeout /t 2 >nul
    "%NSSM%" remove "%SERVICE_NAME%" confirm
    if %errorlevel% neq 0 goto :fail_delete
    echo Service "%SERVICE_NAME%" removed successfully.
    exit /b 0
)

sc.exe stop "%SERVICE_NAME%" >nul 2>&1
timeout /t 2 >nul
sc.exe delete "%SERVICE_NAME%"
if %errorlevel% neq 0 goto :fail_delete

echo Service "%SERVICE_NAME%" removed successfully.
exit /b 0

:fail_delete
echo Failed to remove service "%SERVICE_NAME%".
timeout /t 5 >nul
exit /b 1
