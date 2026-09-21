@echo off
rem One-time bootstrap: fetches a portable Go toolchain and MinGW-w64 into
rem .tools, then downloads the Go module dependencies. Takes a few minutes,
rem almost all of it download time.
rem
rem MinGW is not optional: g3n binds GLFW and OpenGL through cgo, and cgo on
rem Windows needs a GCC-compatible compiler. Nothing is installed system-wide --
rem delete .tools and the machine is exactly as it was.
setlocal
cd /d "%~dp0"

set "GO_VERSION=1.27.1"
set "MINGW_TAG=16.2.0posix-14.0.0-ucrt-r1"
set "MINGW_ZIP=winlibs-x86_64-posix-seh-gcc-16.2.0-mingw-w64ucrt-14.0.0-r1.zip"
set "TOOLS=%~dp0.tools"

if not exist "%TOOLS%" mkdir "%TOOLS%"

if not exist "%TOOLS%\go\bin\go.exe" (
  echo === Downloading Go %GO_VERSION%
  curl -sSLf -o "%TOOLS%\go.zip" "https://go.dev/dl/go%GO_VERSION%.windows-amd64.zip" || exit /b 1
  tar -xf "%TOOLS%\go.zip" -C "%TOOLS%" || exit /b 1
  del "%TOOLS%\go.zip"
)

if not exist "%TOOLS%\mingw64\bin\gcc.exe" (
  echo === Downloading MinGW-w64
  curl -sSLf -o "%TOOLS%\mingw.zip" ^
    "https://github.com/brechtsanders/winlibs_mingw/releases/download/%MINGW_TAG%/%MINGW_ZIP%" || exit /b 1
  tar -xf "%TOOLS%\mingw.zip" -C "%TOOLS%" || exit /b 1
  del "%TOOLS%\mingw.zip"
)

echo === Downloading Go modules
call "%TOOLS%\env.cmd"
go mod download || exit /b 1

echo.
echo Done. Now run build.cmd, or run.cmd to build and play.
