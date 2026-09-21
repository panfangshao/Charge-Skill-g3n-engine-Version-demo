@echo off
rem Captures a reference frame (and the simulation state) from the original
rem Godot project, for diffing against this port.
rem
rem   tools\capture-godot.cmd <godot.exe> <out.png> [ticks] [press] [press-at]
rem
rem e.g.  tools\capture-godot.cmd ..\Godot*\Godot_..._console.exe ref.png 90 space 10
rem
rem The arguments are positional rather than --flag=value because cmd.exe
rem splits arguments on '=' and the values would arrive shredded.
rem
rem ..\aaa is never modified: prepare-godot-ref.ps1 copies the project to a
rem scratch directory and adds tools\capture.gd there as an autoload. Use the
rem *console* Godot binary, or the STATE line goes nowhere.
setlocal
cd /d "%~dp0.."

if "%~2"=="" (
  echo usage: tools\capture-godot.cmd ^<godot.exe^> ^<out.png^> [ticks] [press] [press-at]
  exit /b 2
)

set "GODOT=%~1"
set "OUT=%~f2"
set "TICKS=%~3"
set "PRESS=%~4"
set "PRESS_AT=%~5"
if "%TICKS%"=="" set "TICKS=30"
if "%PRESS_AT%"=="" set "PRESS_AT=0"

set "WORK=%TEMP%\charge-demo-godot-ref"

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0prepare-godot-ref.ps1" ^
  -Source "%~dp0..\..\aaa" -Work "%WORK%" -CaptureScript "%~dp0capture.gd" || exit /b 1

if "%PRESS%"=="" (
  "%GODOT%" --path "%WORK%" --resolution 1280x720 --rendering-driver opengl3 -- ^
    "--screenshot=%OUT%" "--ticks=%TICKS%"
) else (
  "%GODOT%" --path "%WORK%" --resolution 1280x720 --rendering-driver opengl3 -- ^
    "--screenshot=%OUT%" "--ticks=%TICKS%" "--press=%PRESS%" "--press-at=%PRESS_AT%"
)
