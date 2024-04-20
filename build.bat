@echo off
set CGO_ENABLED=0
set GOARCH=amd64
set GOOS=linux

REM Git Version
for /f "tokens=3" %%i in ('git --version') do set GIT_VERSION=%%i
echo Git version is %GIT_VERSION%
REM GO Version
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
echo Git version is %GO_VERSION%
REM 获取当前日期
for /f "tokens=1-3 delims=/ " %%a in ("%DATE%") do (
    set "DAY=%%a"
    set "MONTH=%%b"
    set "YEAR=%%c"
)
REM 格式化日期为 YYYY-MM-DD
set "FORMATTED_DATE=%YEAR%-%MONTH%-%DAY%"
echo Formatted date: %FORMATTED_DATE%

set TJ_VERSION=2.15.3
echo TJ version is %TJ_VERSION%
go build -ldflags "-w -s -X 'trojan/trojan.MVersion=%TJ_VERSION%' -X 'trojan/trojan.BuildDate=%FORMATTED_DATE%' -X 'trojan/trojan.GoVersion=%GO_VERSION%' -X 'trojan/trojan.GitVersion=%GIT_VERSION%'" -o "result/trojan" .
rem go build -ldflags "-w -s -X 'trojan/trojan.MVersion=%version%' -X 'trojan/trojan.BuildDate=%DATE%' -X 'trojan/trojan.GoVersion=%goversion%' -X 'trojan/trojan.GitVersion=%gitversion%'" -o "result/trojan-linux-amd64" .
