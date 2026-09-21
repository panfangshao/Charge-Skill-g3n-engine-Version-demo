// Package shaders holds the GLSL this port adds to g3n, and the materials that
// drive it.
//
// g3n's own lighting is Blinn-Phong in camera space, which cannot be talked
// into matching Godot's GGX-plus-filmic-tonemap output. Rather than fight it,
// these register two extra programs with the renderer's shader manager and
// bypass g3n's material types entirely.
package shaders

import (
	_ "embed"

	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/texture"

	"chargedemo/internal/colors"
	"chargedemo/internal/config"
)

//go:embed godotlit_vertex.glsl
var godotLitVertex string

//go:embed godotlit_fragment.glsl
var godotLitFragment string

//go:embed unlit_vertex.glsl
var unlitVertex string

//go:embed unlit_fragment.glsl
var unlitFragment string

//go:embed label_vertex.glsl
var labelVertex string

//go:embed label_fragment.glsl
var labelFragment string

const (
	godotLitProgram = "godotlit"
	unlitProgram    = "godotunlit"
	labelProgram    = "godotlabel"
)

// Register adds this port's programs to the renderer. Call once, after
// AddDefaultShaders.
func Register(rend *renderer.Renderer) {
	rend.AddShader("godotlit_vertex", godotLitVertex)
	rend.AddShader("godotlit_fragment", godotLitFragment)
	rend.AddProgram(godotLitProgram, "godotlit_vertex", "godotlit_fragment")

	rend.AddShader("godotunlit_vertex", unlitVertex)
	rend.AddShader("godotunlit_fragment", unlitFragment)
	rend.AddProgram(unlitProgram, "godotunlit_vertex", "godotunlit_fragment")

	rend.AddShader("godotlabel_vertex", labelVertex)
	rend.AddShader("godotlabel_fragment", labelFragment)
	rend.AddProgram(labelProgram, "godotlabel_vertex", "godotlabel_fragment")
}

// SkipNone and friends tell the lit shader which capsule, if any, is the
// surface currently being shaded, so it does not shadow itself.
const (
	SkipNone   = 0
	SkipPlayer = 1
	SkipTarget = 2
)

// Lit is one StandardMaterial3D from the Godot scene.
type Lit struct {
	material.Material
	uniParams gls.Uniform
	uniShadow gls.Uniform
	uniCamera gls.Uniform

	params [5]math32.Vector3
	shadow [3]math32.Vector3
	camera math32.Vector3
}

// NewLit builds a lit material. The ambient and sun radiances are the same for
// every material in the scene; they live in the material block because that is
// the only uniform storage a g3n material owns.
func NewLit(albedo math32.Color, metallic, roughness float32, skip float32) *Lit {
	m := new(Lit)
	m.Material.Init()
	m.SetShader(godotLitProgram)
	m.uniParams.Init("MatParams")
	m.uniShadow.Init("ShadowParams")
	m.uniCamera.Init("CameraPosition")

	ambient := colors.Lin(config.AmbientColor)
	ambient.MultiplyScalar(config.AmbientEnergy * config.AmbientGain)
	sun := colors.Lin(config.SunColor)
	sun.MultiplyScalar(config.SunEnergy * config.LightGain)
	// SunTravel is the direction the light travels; shading wants the direction
	// back towards it.
	sunDir := config.SunTravel
	sunDir.Negate()
	sunDir.Normalize()

	m.params[0] = math32.Vector3{X: albedo.R, Y: albedo.G, Z: albedo.B}
	m.params[1] = ambient
	m.params[2] = sun
	m.params[3] = sunDir
	m.params[4] = math32.Vector3{X: metallic, Y: roughness, Z: 1}

	m.shadow[2] = math32.Vector3{
		X: config.CapsuleRadius,
		Y: config.CapsuleHeight/2 - config.CapsuleRadius,
		Z: skip,
	}
	return m
}

// SetShadowCasters points the analytic shadow test at the current capsule
// positions. Called once per frame, from the app loop.
func (m *Lit) SetShadowCasters(player, target math32.Vector3) {
	m.shadow[0] = player
	m.shadow[1] = target
}

// SetCameraPosition supplies the eye point the view vector is measured from.
func (m *Lit) SetCameraPosition(p math32.Vector3) { m.camera = p }

// RenderSetup is called by the engine before drawing an object with this
// material.
func (m *Lit) RenderSetup(gs *gls.GLS) {
	m.Material.RenderSetup(gs)
	gs.Uniform3fv(m.uniParams.Location(gs), 5, &m.params[0].X)
	gs.Uniform3fv(m.uniShadow.Location(gs), 3, &m.shadow[0].X)
	gs.Uniform3f(m.uniCamera.Location(gs), m.camera.X, m.camera.Y, m.camera.Z)
}

// Unlit is Godot's SHADING_MODE_UNSHADED. The colour handed in is the raw Godot
// albedo; the tonemap Godot would have applied is baked in here.
type Unlit struct {
	material.Material
	uniParams gls.Uniform
	params    [2]math32.Vector3
}

// NewUnlit builds an unlit material. Transparent materials get alpha blending
// and depth writes off, matching Godot's TRANSPARENCY_ALPHA.
func NewUnlit(color math32.Color4, transparent bool) *Unlit {
	m := new(Unlit)
	m.Material.Init()
	m.SetShader(unlitProgram)
	m.uniParams.Init("MatParams")
	m.SetColor(color)
	if transparent {
		m.SetTransparent(true)
		m.SetDepthMask(false)
		m.SetBlending(material.BlendNormal)
	}
	return m
}

// SetColor sets the Godot albedo, baking the tonemapper into it.
func (m *Unlit) SetColor(color math32.Color4) {
	baked := colors.Unshaded(color)
	m.params[0] = math32.Vector3{X: baked.R, Y: baked.G, Z: baked.B}
	m.params[1] = math32.Vector3{X: baked.A}
}

// RenderSetup is called by the engine before drawing an object with this
// material.
func (m *Unlit) RenderSetup(gs *gls.GLS) {
	m.Material.RenderSetup(gs)
	gs.Uniform3fv(m.uniParams.Location(gs), 2, &m.params[0].X)
}

// Label is the textured, unlit quad behind a Label3D. Sprite only uploads MVP,
// so this program deliberately asks for nothing else.
type Label struct {
	material.Material
}

// NewLabel builds a label material around an already-rasterised text texture.
func NewLabel(tex *texture.Texture2D) *Label {
	m := new(Label)
	m.Material.Init()
	m.SetShader(labelProgram)
	m.SetTransparent(true)
	m.SetDepthMask(false)
	m.SetBlending(material.BlendNormal)
	m.SetSide(material.SideDouble)
	m.AddTexture(tex)
	return m
}
