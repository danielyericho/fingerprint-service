@echo off
:: -------------------- Auto-elevate --------------------
net session >nul 2>&1
if %errorlevel% neq 0 (
    powershell -Command "Start-Process '%~f0' -Verb runAs"
    exit /b
)

:: -------------------- Config --------------------------
set SERVICE_NAME=FingerprintQubuService
set SERVICE_DISPLAY=Fingerprint Qubu Service

set SCRIPT_DIR=%~dp0
if "%SCRIPT_DIR:~-1%"=="\" set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%

set NSSM=%SCRIPT_DIR%\nssm.exe
set EXE_PATH=%SCRIPT_DIR%\fingerprint-service.exe
set LOG_DIR=%SCRIPT_DIR%\logs

:: -------------------- Pre-checks ----------------------
if not exist "%NSSM%" (
    echo nssm.exe not found. Attempting automatic download...
    call "%SCRIPT_DIR%\download-nssm.bat"
    if %errorlevel% neq 0 exit /b 1
)
if not exist "%NSSM%" (
    echo [ERROR] nssm.exe was not found at "%NSSM%".
    echo Run download-nssm.bat or copy nssm.exe manually.
    timeout /t 8 >nul
    exit /b 1
)
if not exist "%EXE_PATH%" (
    echo [ERROR] fingerprint-service.exe was not found at "%EXE_PATH%".
    echo Build it with: go build -o bin\fingerprint-service.exe ./cmd/server
    timeout /t 5 >nul
    exit /b 1
)

if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"

echo Installing service "%SERVICE_NAME%" from "%EXE_PATH%"
echo Logs will be written to "%LOG_DIR%"

:: -------------------- Install Logic -------------------
sc.exe query "%SERVICE_NAME%" >nul 2>&1
if %errorlevel% equ 0 (
    echo Service "%SERVICE_NAME%" already exists. Starting...
    "%NSSM%" start "%SERVICE_NAME%" >nul 2>&1
    exit /b 0
)

"%NSSM%" install "%SERVICE_NAME%" "%EXE_PATH%"
if %errorlevel% neq 0 goto :fail_create

"%NSSM%" set "%SERVICE_NAME%" AppDirectory "%SCRIPT_DIR%"
"%NSSM%" set "%SERVICE_NAME%" DisplayName "%SERVICE_DISPLAY%"
"%NSSM%" set "%SERVICE_NAME%" Description "Fingerprint Qubu HTTP API service (managed by NSSM)."
"%NSSM%" set "%SERVICE_NAME%" Start SERVICE_AUTO_START

"%NSSM%" set "%SERVICE_NAME%" AppStdout "%LOG_DIR%\fingerprint.out.log"
"%NSSM%" set "%SERVICE_NAME%" AppStderr "%LOG_DIR%\fingerprint.err.log"
"%NSSM%" set "%SERVICE_NAME%" AppStdoutCreationDisposition 4
"%NSSM%" set "%SERVICE_NAME%" AppStderrCreationDisposition 4

"%NSSM%" set "%SERVICE_NAME%" AppRotateFiles 1
"%NSSM%" set "%SERVICE_NAME%" AppRotateOnline 1
"%NSSM%" set "%SERVICE_NAME%" AppRotateSeconds 86400
"%NSSM%" set "%SERVICE_NAME%" AppRotateBytes 10485760

"%NSSM%" set "%SERVICE_NAME%" AppStopMethodSkip 0
"%NSSM%" set "%SERVICE_NAME%" AppStopMethodConsole 5000
"%NSSM%" set "%SERVICE_NAME%" AppStopMethodWindow 5000
"%NSSM%" set "%SERVICE_NAME%" AppStopMethodThreads 5000

"%NSSM%" set "%SERVICE_NAME%" AppExit Default Restart
"%NSSM%" set "%SERVICE_NAME%" AppRestartDelay 5000
"%NSSM%" set "%SERVICE_NAME%" AppThrottle 10000

:: -------------------- Start ---------------------------
"%NSSM%" start "%SERVICE_NAME%" >nul 2>&1
if %errorlevel% neq 0 goto :fail_start

echo Service "%SERVICE_NAME%" installed and started successfully.
echo Logs: "%LOG_DIR%\fingerprint.err.log" (errors / app log) and "%LOG_DIR%\fingerprint.out.log".
exit /b 0

:fail_create
echo Failed to install service "%SERVICE_NAME%".
timeout /t 5 >nul
exit /b 1

:fail_start
echo Service "%SERVICE_NAME%" was created but failed to start.
echo Check "%LOG_DIR%\fingerprint.err.log" for details.
timeout /t 5 >nul
exit /b 1
