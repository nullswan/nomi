@echo off
setlocal enabledelayedexpansion

rem Define variables
set REPO=nullswan/golem
for /f "tokens=2 delims=: " %%i in ('curl -s https://api.github.com/repos/%REPO%/releases/latest ^| findstr "tag_name"') do set LATEST_RELEASE=%%i
set PLATFORM=windows
set ARCH=amd64

rem Construct the download URL
set DOWNLOAD_URL=https://github.com/%REPO%/releases/download/%LATEST_RELEASE%/golem-cli_%LATEST_RELEASE%_%PLATFORM%_%ARCH%.zip

rem Download and extract
echo Downloading !DOWNLOAD_URL!...
curl -L -o golem-cli.zip !DOWNLOAD_URL!
PowerShell -Command "Expand-Archive -Path golem-cli.zip -DestinationPath ."
move golem-cli.exe C:\Program Files\golem-cli\

rem Clean up
del golem-cli.zip

echo golem-cli installed successfully!
