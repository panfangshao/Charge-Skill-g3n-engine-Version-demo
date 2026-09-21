# Copies the Godot project to a scratch directory and registers capture.gd as
# an autoload there, so the reference capture never touches ../aaa.
#
# Called by capture-godot.cmd.
param(
    [Parameter(Mandatory = $true)][string]$Source,
    [Parameter(Mandatory = $true)][string]$Work,
    [Parameter(Mandatory = $true)][string]$CaptureScript
)

$ErrorActionPreference = 'Stop'

if (-not (Test-Path (Join-Path $Source 'project.godot'))) {
    throw "Cannot find a Godot project at $Source"
}

if (Test-Path $Work) { Remove-Item -Recurse -Force $Work }
New-Item -ItemType Directory -Force -Path $Work | Out-Null

foreach ($name in 'project.godot', 'icon.svg', 'icon.svg.import') {
    $path = Join-Path $Source $name
    if (Test-Path $path) { Copy-Item $path $Work }
}
foreach ($dir in 'scenes', 'scripts') {
    $path = Join-Path $Source $dir
    if (Test-Path $path) { Copy-Item -Recurse $path (Join-Path $Work $dir) }
}
Copy-Item $CaptureScript (Join-Path $Work 'capture.gd')

# Drop the editor addon -- it is tooling, not part of the demo -- and register
# the capture autoload in its place.
$projectFile = Join-Path $Work 'project.godot'
$text = Get-Content -Raw -Encoding UTF8 $projectFile
$text = $text -replace '(?m)^_mcp_game_helper=.*\r?\n', ''
$text = $text -replace '(?m)^enabled=PackedStringArray\("res://addons/.*\r?\n', ''
$text = $text -replace '(?m)^\[editor_plugins\]\r?\n(\r?\n)*', ''
$text = $text -replace '(?m)^\[autoload\]\r?\n(\r?\n)*', ''
$text = $text.TrimEnd() + "`r`n`r`n[autoload]`r`n`r`n_capture=`"*res://capture.gd`"`r`n"
Set-Content -Encoding UTF8 $projectFile $text
