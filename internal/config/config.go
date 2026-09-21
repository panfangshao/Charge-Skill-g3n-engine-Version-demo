// Package config holds every constant in this port, transcribed from the Godot
// project (../aaa).
//
// Each value names its counterpart in scenes/charge_demo.tscn or
// scripts/charge_demo.gd so the two projects can be diffed by eye.
package config

import "github.com/g3n/engine/math32"

// ------------------------------------------------------------------ gameplay
// scripts/charge_demo.gd constants
const (
	MoveSpeed   = 5.0
	ChargeSpeed = 24.0
	ChargeRange = 14.0
	Cooldown    = 3.0

	KnockbackSpeed = 12.0  // hit_target()
	KnockbackDecel = 20.0  // move_toward(Vector3.ZERO, 20.0 * delta)
	TrailInterval  = 0.035 // trail_timer reset
	EffectDuration = 0.3   // tween duration in spawn_effect()
)

// clamp_to_arena()
const (
	ArenaX = 11.0
	ArenaZ = 8.0
	ActorY = 0.9
)

// Godot runs _physics_process at a fixed 60 Hz by default.
const (
	PhysicsStep     = 1.0 / 60.0
	MaxPhysicsSteps = 8 // physics/common/max_physics_steps_per_frame
)

// --------------------------------------------------------------------- scene
// Spawn transforms (charge_demo.tscn / reset_demo())
var (
	PlayerStart = math32.Vector3{X: -5, Y: ActorY, Z: 0}
	TargetStart = math32.Vector3{X: 4, Y: ActorY, Z: 0}
)

// CapsuleMesh / CapsuleShape3D. Godot's height spans the whole capsule.
const (
	CapsuleRadius = 0.45
	CapsuleHeight = 1.8
)

// Ground: StaticBody3D at (0, -0.2, 0), MeshInstance3D child at (0, 0, -0.31181765)
var (
	GroundSize   = math32.Vector3{X: 24, Y: 0.4, Z: 18}
	GroundMeshAt = math32.Vector3{X: 0, Y: -0.2, Z: -0.31181765}
)

// Camera3D: orthographic, size 25.5, KEEP_HEIGHT aspect.
// The Transform3D basis works out to a rotation of -24.9537 deg about X
// (its sine is the 0.42195368 in the scene file). g3n shares Godot's
// handedness -- Y up, -Z forward -- so the transform carries over unchanged.
var CameraPos = math32.Vector3{X: 0, Y: 5.3208942, Z: 14.248308}

const (
	CameraPitchDeg    = -24.9537
	CameraOrthoHeight = 25.5
)

// DirectionalLight3D "Sun": rotation (x = -55, y = -30), energy 1.5.
// SunTravel is the -Z column of its basis, i.e. the direction the light
// travels; SunDir points back towards the light, which is what shading wants.
var (
	SunTravel = math32.Vector3{X: 0.28678823, Y: -0.81915206, Z: -0.49673176}
	SunColor  = math32.Color{R: 1, G: 0.9, B: 0.78}
)

const SunEnergy = 1.5

// Environment
var (
	BgColor      = math32.Color{R: 0.035, G: 0.055, B: 0.09}
	AmbientColor = math32.Color{R: 0.65, G: 0.78, B: 1}
)

const AmbientEnergy = 0.65

// StandardMaterial3D resources
var GroundAlbedo = math32.Color{R: 0.12, G: 0.19, B: 0.24}

const (
	GroundMetallic  = 0.0
	GroundRoughness = 0.9
)

var BlueAlbedo = math32.Color{R: 0.1, G: 0.65, B: 1}

const (
	BlueMetallic  = 0.25
	BlueRoughness = 0.3
)

var RedAlbedo = math32.Color{R: 1, G: 0.25, B: 0.2}

const (
	RedMetallic  = 0.15
	RedRoughness = 0.35
)

// add_grid() / add_grid_line()
var (
	GridColor = math32.Color{R: 0.22, G: 0.34, B: 0.4}
	GridXSize = math32.Vector3{X: 0.025, Y: 0.015, Z: 18}
	GridZSize = math32.Vector3{X: 24, Y: 0.015, Z: 0.025}
)

const GridY = 0.012

// spawn_effect()
var TrailColor = math32.Color4{R: 0.1, G: 0.7, B: 1, A: 0.5}

const (
	TrailRadius   = 0.3
	TrailEndScale = 0.1
)

var ImpactColor = math32.Color4{R: 1, G: 0.65, B: 0.15, A: 0.8}

const (
	ImpactRadius   = 0.5
	ImpactEndScale = 3.5
)

// Label3D above each actor
const (
	LabelOffsetY   = 1.65
	LabelFontSize  = 40.0
	LabelPixelSize = 0.005 // Godot Label3D default
	LabelOutline   = 8
)

var (
	PlayerLabelColor = math32.Color{R: 0.4, G: 0.85, B: 1}
	TargetLabelColor = math32.Color{R: 1, G: 0.55, B: 0.4}
)

// ----------------------------------------------------------------------- HUD
// g3n's text.Font defaults to 72 DPI, so a point is exactly a pixel and
// Godot's font_size values carry over unchanged.
const (
	BaseWidth  = 1280.0 // project.godot viewport size
	BaseHeight = 720.0
)

var TitlePos = math32.Vector2{X: 32, Y: 24}

const TitleFontSize = 30.0

var InstructionsPos = math32.Vector2{X: 34, Y: 72}

const InstructionsFontSize = 18.0

var InstructionsColor = math32.Color{R: 0.65, G: 0.78, B: 0.88}

const (
	PanelMarginX          = 24.0
	PanelTopFromBottom    = 122.0
	PanelBottomFromBottom = 24.0
	PanelHeight           = PanelTopFromBottom - PanelBottomFromBottom
	PanelBorderWidth      = 1
	PanelCornerRadius     = 16
)

var (
	PanelBg     = math32.Color4{R: 0.035, G: 0.055, B: 0.09, A: 0.94}
	PanelBorder = math32.Color{R: 0.2, G: 0.4, B: 0.52}
)

var StatusPos = math32.Vector2{X: 24, Y: 17}

const StatusFontSize = 22.0

var DetailsPos = math32.Vector2{X: 24, Y: 55}

const DetailsFontSize = 16.0

var DetailsColor = math32.Color{R: 0.6, G: 0.73, B: 0.82}

const (
	ButtonRightInset   = 20.0  // offset_right relative to the panel
	ButtonWidth        = 222.0 // 242 - 20
	ButtonTop          = 15.0
	ButtonHeight       = 68.0 // 83 - 15
	ButtonFontSize     = 24.0
	ButtonCornerRadius = 3
)

// Godot 4 default theme (scene/theme/default_theme.cpp)
var (
	ControlFontColor        = math32.Color{R: 0.875, G: 0.875, B: 0.875}
	ButtonNormalBg          = math32.Color4{R: 0.1, G: 0.1, B: 0.1, A: 0.6}
	ButtonHoverBg           = math32.Color4{R: 0.225, G: 0.225, B: 0.225, A: 0.6}
	ButtonPressedBg         = math32.Color4{R: 0, G: 0, B: 0, A: 0.6}
	ButtonDisabledBg        = math32.Color4{R: 0.1, G: 0.1, B: 0.1, A: 0.3}
	ButtonFontDisabledColor = math32.Color4{R: 0.875, G: 0.875, B: 0.875, A: 0.5}
)

// ------------------------------------------------------------ light calibration
// Godot's gl_compatibility renderer folds a few constants into the light
// uniforms it uploads (notably the pi that cancels the Lambert 1/pi). These two
// gains were fitted by least squares against screenshots of the Godot build.
const (
	AmbientGain = 2.4172
	LightGain   = 1.8278
)

// -------------------------------------------------------------------- window
// Godot falls back to the system CJK font; name the same family explicitly.
// The .ttc entries are font collections, which the TrueType parser underneath
// g3n cannot read directly -- see internal/fontutil.
var FontCandidates = []string{
	`C:\Windows\Fonts\msyh.ttc`, // Microsoft YaHei
	`C:\Windows\Fonts\msyh.ttf`,
	`C:\Windows\Fonts\Deng.ttf`,
	`C:\Windows\Fonts\simhei.ttf`,
}
