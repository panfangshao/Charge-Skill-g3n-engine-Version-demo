// Package colors reproduces Godot's colour-space behaviour on the CPU.
//
// Godot stores colours as sRGB and lights in linear space. Its gl_compatibility
// renderer then tonemaps and re-encodes *per fragment* rather than in a resolve
// pass, which this port reproduces: in GLSL for lit surfaces (see
// internal/shaders), and here for unlit ones.
package colors

import (
	"math"

	"github.com/g3n/engine/math32"
)

// SrgbToLinear is the exact sRGB transfer function, not the cheap polynomial
// fit engines often ship. The difference shows in the near-black background
// this scene is full of.
func SrgbToLinear(c float32) float32 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return float32(math.Pow(float64((c+0.055)/1.055), 2.4))
}

// LinearToSrgb is the inverse of SrgbToLinear.
func LinearToSrgb(c float32) float32 {
	if c <= 0.0031308 {
		return c * 12.92
	}
	return 1.055*float32(math.Pow(float64(c), 1.0/2.4)) - 0.055
}

// TonemapFilmic is Godot's tonemap_filmic with its exposure bias of 2.0 folded
// into A and B, and the default white point of 1.0.
func TonemapFilmic(c float32) float32 {
	const (
		a     = 0.88 // 0.22 * bias * bias
		b     = 0.60 // 0.30 * bias
		cc    = 0.10
		d     = 0.20
		e     = 0.01
		f     = 0.30
		white = 1.0
	)
	mapped := ((c*(a*c+cc*b) + d*e) / (c*(a*c+b) + d*f)) - e/f
	const whiteMapped float32 = ((white*(a*white+cc*b) + d*e) / (white*(a*white+b) + d*f)) - e/f
	return mapped / whiteMapped
}

// Lin linearises an sRGB colour the way Godot stores it, for the lit shader.
func Lin(c math32.Color) math32.Vector3 {
	return math32.Vector3{X: SrgbToLinear(c.R), Y: SrgbToLinear(c.G), Z: SrgbToLinear(c.B)}
}

// Unshaded is what Godot's SHADING_MODE_UNSHADED actually puts on screen.
//
// The compatibility renderer tonemaps inside the object shader, so an unlit
// surface ends up as srgb(tonemap(linear(albedo))). Baking that here lets the
// unlit geometry -- grid, effect puffs, labels -- ride an ordinary unlit
// material and still land on the same pixels.
func Unshaded(c math32.Color4) math32.Color4 {
	f := func(v float32) float32 { return LinearToSrgb(TonemapFilmic(SrgbToLinear(v))) }
	return math32.Color4{R: f(c.R), G: f(c.G), B: f(c.B), A: c.A}
}

// Unshaded3 is Unshaded for an opaque colour.
func Unshaded3(c math32.Color) math32.Color {
	out := Unshaded(math32.Color4{R: c.R, G: c.G, B: c.B, A: 1})
	return math32.Color{R: out.R, G: out.G, B: out.B}
}

// OverSrgb composites top over bottom the way Godot does: in sRGB space.
//
// Godot draws its canvas items after the per-fragment tonemap, so a translucent
// Control blends against sRGB values, while a GL blend works on whatever is
// already in the frame buffer. Both HUD stacks here sit on a known, uniform
// background, so flattening them up front removes the difference entirely.
func OverSrgb(top math32.Color4, bottom math32.Color) math32.Color {
	a := top.A
	return math32.Color{
		R: top.R*a + bottom.R*(1-a),
		G: top.G*a + bottom.G*(1-a),
		B: top.B*a + bottom.B*(1-a),
	}
}
