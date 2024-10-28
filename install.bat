@echo off
setlocal enabledelayedexpansion

rem Define variables
set REPO=nullswan/nomi
for /f "tokens=2 delims=: " %%i in ('curl -s https://api.github.com/repos/%REPO%/releases/latest ^| findstr "tag_name"') do set LATEST_RELEASE=%%i
set PLATFORM=windows
set ARCH=amd64

rem Construct the download URL
set DOWNLOAD_URL=https://github.com/%REPO%/releases/download/%LATEST_RELEASE%/nomi-cli_%LATEST_RELEASE%_%PLATFORM%_%ARCH%.zip

rem Download and extract
echo Downloading !DOWNLOAD_URL!...
curl -L -o --progress-bar nomi-cli.zip !DOWNLOAD_URL!
PowerShell -Command "Expand-Archive -Path nomi-cli.zip -DestinationPath ."
move nomi-cli.exe C:\Program Files\nomi\nomi.exe

rem Clean up
del nomi-cli.zip

echo nomi-cli installed successfully!
