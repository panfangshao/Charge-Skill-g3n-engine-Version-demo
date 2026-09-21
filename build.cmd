@echo off
rem Builds the game into build\chargedemo.exe. Run setup.cmd first (once).
setlocal
call "%~dp0.tools\env.cmd"
if not exist "%~dp0.tools\go\bin\go.exe" (
  echo The Go toolchain is missing. Run setup.cmd first.
  exit /b 1
)
cd /d "%~dp0"
go build -o build\chargedemo.exe . || exit /b 1
