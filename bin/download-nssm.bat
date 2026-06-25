@echo off
setlocal EnableDelayedExpansion

set SCRIPT_DIR=%~dp0
if "%SCRIPT_DIR:~-1%"=="\" set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%

set NSSM=%SCRIPT_DIR%\nssm.exe
if exist "%NSSM%" (
    echo nssm.exe already exists at "%NSSM%"
    exit /b 0
)

echo Downloading NSSM...

powershell -NoProfile -ExecutionPolicy Bypass -Command ^
  "$ErrorActionPreference='Stop';" ^
  "$dest='%NSSM%';" ^
  "$urls=@('https://nssm.cc/ci/nssm-2.24-101-g897c7ad.zip','https://nssm.cc/release/nssm-2.24.zip');" ^
  "$ok=$false;" ^
  "foreach($url in $urls){" ^
  "  try {" ^
  "    $zip=Join-Path $env:TEMP ('nssm-' + [guid]::NewGuid().ToString() + '.zip');" ^
  "    $extract=Join-Path $env:TEMP ('nssm-extract-' + [guid]::NewGuid().ToString());" ^
  "    Write-Host ('Trying ' + $url);" ^
  "    Invoke-WebRequest -Uri $url -OutFile $zip;" ^
  "    New-Item -ItemType Directory -Force -Path $extract | Out-Null;" ^
  "    Expand-Archive -Path $zip -DestinationPath $extract -Force;" ^
  "    $candidate=Get-ChildItem -Path $extract -Recurse -Filter nssm.exe | Select-Object -First 1;" ^
  "    if(-not $candidate){ throw 'nssm.exe not found in archive' };" ^
  "    Copy-Item $candidate.FullName -Destination $dest -Force;" ^
  "    $ok=$true; break" ^
  "  } catch { Write-Host ('Failed: ' + $_.Exception.Message) }" ^
  "};" ^
  "if(-not $ok){" ^
  "  if(Get-Command winget -ErrorAction SilentlyContinue){" ^
  "    Write-Host 'Trying winget install NSSM.NSSM...';" ^
  "    winget install NSSM.NSSM --accept-package-agreements --accept-source-agreements | Out-Null;" ^
  "    $wingetExe=Get-ChildItem -Path (Join-Path $env:LOCALAPPDATA 'Microsoft\WinGet') -Recurse -Filter nssm.exe -ErrorAction SilentlyContinue | Select-Object -First 1;" ^
  "    if($wingetExe){ Copy-Item $wingetExe.FullName -Destination $dest -Force; $ok=$true }" ^
  "  }" ^
  "};" ^
  "if(-not $ok){ exit 1 }"

if errorlevel 1 (
    echo [ERROR] Failed to download nssm.exe automatically.
    echo Try manually:
    echo   winget install NSSM.NSSM
    echo   copy from WinGet folder to "%NSSM%"
    echo Or download from https://nssm.cc/ci/nssm-2.24-101-g897c7ad.zip
    timeout /t 8 >nul
    exit /b 1
)

echo nssm.exe ready at "%NSSM%"
exit /b 0
