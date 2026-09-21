@echo off
rem Builds the game, then runs it.
setlocal
call "%~dp0build.cmd" || exit /b 1
"%~dp0build\chargedemo.exe" %*
