@echo off
rem Runs the simulation tests. No window needed.
setlocal
call "%~dp0.tools\env.cmd"
cd /d "%~dp0"
go test ./internal/... %*
