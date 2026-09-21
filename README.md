# 冲锋训练场 — g3n port

A [g3n](https://github.com/g3n/engine) reimplementation of the Godot project in
`../aaa`, built to look and behave the same.

```
setup.cmd   :: once: fetches a portable Go toolchain and MinGW-w64 into .tools
run.cmd     :: builds the game and plays it
test.cmd    :: runs the simulation tests, no window needed
```

Controls are unchanged: **WASD / arrow keys** to move, **Space** or the **冲锋**
button to charge, **R** to reset, **Esc** to quit.

## Requirements

`git` and `curl` on PATH, and nothing else. `setup.cmd` puts Go and MinGW-w64
in `.tools/`; delete that directory and the machine is exactly as it was.

MinGW is not optional. g3n binds GLFW and OpenGL through cgo, and cgo on Windows
needs a GCC-compatible compiler — the LLVM that ships on many dev machines
targets the MSVC ABI, which cgo cannot use.

## How close is it?

Both builds were run at 1280x720 and screenshotted on the same physics tick,
then compared pixel by pixel (`tools/compare.py`).

| Region | idle | mid-charge | after the hit |
| --- | --- | --- | --- |
| Whole frame | 2.70 | 2.26 | 2.65 |
| Arena (3D) | 2.26 | 2.33 | 2.25 |
| Bottom panel | 7.49 | 4.03 | 7.15 |
| Title / instructions | 14.73 | 14.73 | 14.73 |

(mean per-channel error, 0-255)

The background, the panel fill and the button fill come out **exactly** equal
(`9,14,23` / `9,14,23` / `19,21,24` in both). The panel border is within 0.2/255
and the ground within 4/255. The capsules carry most of the arena's remaining
error, about 6.5 over their bounding boxes, in the specular highlight and the
shadow falloff.

The two text regions carry the rest, and it is mostly typeface: Godot embeds
Open Sans SemiBold for Latin and falls back to the system CJK font, while this
port sets everything in Microsoft YaHei. The 3D labels are a little thinner than
Godot's, because Godot dilates the glyph outline properly and this port stamps
the text around a circle instead.

Simulation agreement is exact rather than approximate. Driving both engines from
the same tape — press Space on physics tick 10:

| Tick | Godot | g3n |
| --- | --- | --- |
| 20 (mid-charge) | player x = -1.0000, cooldown 2.8333 | player x = -1.0000, cooldown 2.8333 |
| 90 (after the hit) | player x = 3.0977, target x = 7.7000, hits 1, cooldown 1.6667 | player x = 3.0990, target x = 7.7000, hits 1, cooldown 1.6667 |

The 1.3 mm left on the player after a 10 m charge is the collision response:
`internal/sim/physics.go` solves the capsule-vs-capsule sweep in closed form,
Godot ran it through Jolt.

## Layout

| File | Counterpart in `../aaa` |
| --- | --- |
| `internal/config/config.go` | every constant from `charge_demo.tscn` + `charge_demo.gd` |
| `internal/config/strings.go` | every user-visible string |
| `internal/sim/game.go` | `charge_demo.gd` (`_physics_process`, `start_charge`, `hit_target`, …) |
| `internal/sim/physics.go` | `CharacterBody3D.move_and_slide` |
| `internal/world/` | the `Node3D` tree: Camera3D, Ground, grid, actors |
| `internal/hud/` | the `CanvasLayer` subtree, and `StyleBoxFlat` |
| `internal/labels/` | the two `Label3D` nodes |
| `internal/effects/` | `spawn_effect()` and its Tween |
| `internal/meshes/` | `BoxMesh`, `CapsuleMesh`, `SphereMesh` |
| `internal/colors/` | Godot's sRGB / tonemap curves, on the CPU |
| `internal/shaders/` | the `gl_compatibility` scene shader |
| `internal/fontutil/` | loading the system CJK font |

`main.go` assembles the application.

## The parts worth knowing about

**Lighting.** g3n's own shading is Blinn-Phong in camera space, which cannot be
talked into matching Godot's GGX-plus-filmic-tonemap output. `internal/shaders`
registers two extra programs with the renderer and bypasses g3n's material types
entirely. Godot's compatibility renderer tonemaps and encodes sRGB *per
fragment* — there is no HDR resolve pass — so the shader does both, which is
also why the translucent effect puffs and the HUD blend against sRGB values the
way they do in Godot. Unlit surfaces still go through Godot's tonemapper, so
their colours are baked through it on the CPU in `colors.Unshaded`.

**Shadows.** g3n has no shadow mapping at all. But this scene has exactly two
shadow casters and both are upright capsules, so the shadow test is a
closed-form clamped closest-point between the sun ray and a vertical segment,
evaluated in the fragment shader. That is cheaper than a shadow map, has no bias
or acne to tune, and gives a silhouette that matches Godot's rather than an
approximation of it. A capsule skips itself, so it is shaded by N·L alone.

**Physics.** g3n has no physics either, and Godot ran this scene on Jolt, so
nothing off the shelf would reproduce Godot's contact solver. Both actors are
capsules of equal radius pinned to `y = 0.9` every frame, which reduces the
whole simulation to circle-vs-circle sweeping in the XZ plane — solved exactly
in `internal/sim/physics.go`. That is what makes the two builds agree to four
decimal places.

**No `app` package.** g3n's `app.App()` opens an audio device, which links
`OpenAL32.dll` and `libvorbis.dll` at load time. Neither ships with Windows, so
an executable built on top of it fails to start with `0xc0000135`. This demo has
no sound; `main.go` makes the same handful of calls `App()` does, minus the
audio.

**Fonts.** Microsoft YaHei ships as `msyh.ttc`, a TrueType *collection*, and the
parser underneath g3n only understands a plain sfnt file.
`internal/fontutil` rebuilds face 0 of the collection into a standalone
in-memory font. Nothing is written to disk and no font is redistributed — the
bytes come from the machine's own font directory at startup. g3n's `text.Font`
defaults to 72 DPI, so Godot's pixel `font_size` values carry over unchanged.

**Coordinates.** g3n shares Godot's handedness — Y up, −Z forward — so every
transform in `charge_demo.tscn` carries over with no conversion at all.

**StyleBoxFlat.** g3n's gui has no corner radius, and its panel shader
composites a texture over the panel colour and then divides by the result alpha,
which is a division by zero wherever the texture is fully transparent. The panel
and the button are therefore rasterised at their exact pixel size with the
background baked into the corners (`internal/hud/stylebox.go`), and re-rasterised
on resize — which the HUD has to do anyway, since g3n's gui has no global scale
and Godot's `canvas_items` stretch mode is applied by hand.

## Development

```
test.cmd
```

12 tests over the simulation, driven from scripted input tapes and compared
against the numbers above. No window needed: `sim.State.Step` takes a plain
delta and an input vector.

To re-run the visual comparison:

```
tools\capture-godot.cmd <godot_console.exe> godot.png 90 space 10
build\chargedemo.exe -screenshot g3n.png -ticks 90 -press space -press-at 10
python tools\compare.py godot.png g3n.png
```

Both harnesses count physics ticks rather than rendered frames — that is the
clock the two engines share — and both freeze the simulation before grabbing the
frame, so the image matches the `STATE` line printed beside it. `../aaa` is never
modified: `tools/prepare-godot-ref.ps1` copies the project to a scratch
directory and adds `tools/capture.gd` there as an autoload.
