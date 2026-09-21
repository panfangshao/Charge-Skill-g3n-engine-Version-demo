// Package labels is the two Label3D nodes floating above the capsules.
//
// g3n has no world-space text, but it does have Sprite: a camera-facing quad,
// which is what Label3D's billboard mode amounts to. The text is rasterised
// once into a texture, outline included, and the sprite is sized in world units
// at Godot's pixel_size so it covers the same screen area.
package labels

import (
	"image"
	"image/draw"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/text"
	"github.com/g3n/engine/texture"

	"chargedemo/internal/colors"
	"chargedemo/internal/config"
	"chargedemo/internal/shaders"
)

// Attach parents a Label3D-equivalent to the actor node.
func Attach(parent *core.Node, font *text.Font, msg string, color math32.Color) {
	img := render(font, msg, color)
	bounds := img.Bounds()

	tex := texture.NewTexture2DFromRGBA(img)
	tex.SetMagFilter(gls.LINEAR)
	tex.SetMinFilter(gls.LINEAR)

	mat := shaders.NewLabel(tex)
	sprite := graphic.NewSprite(
		float32(bounds.Dx())*config.LabelPixelSize,
		float32(bounds.Dy())*config.LabelPixelSize,
		mat)
	sprite.SetPosition(0, config.LabelOffsetY, 0)
	parent.Add(sprite)
}

// outlineOffsets are the directions the outline copies are pushed in. Godot
// dilates the glyph outline properly; stamping the text around a circle is the
// closest a plain rasteriser gets, and at this size the two are hard to tell
// apart.
func outlineOffsets(radius int) []image.Point {
	var pts []image.Point
	r2 := radius * radius
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if dx*dx+dy*dy <= r2 {
				pts = append(pts, image.Point{X: dx, Y: dy})
			}
		}
	}
	return pts
}

// render rasterises the label, outline first and fill on top.
func render(font *text.Font, msg string, color math32.Color) *image.RGBA {
	// Godot's outline_size is the diameter of the dilation; half of it is the
	// radius the stamps are pushed out by.
	radius := config.LabelOutline / 2

	font.SetPointSize(config.LabelFontSize)
	font.SetDPI(72) // a point is a pixel, matching Godot's pixel font_size
	font.SetLineSpacing(1)

	// Label3D draws unshaded, and Godot tonemaps that before it reaches the
	// frame buffer, so both colours go through the same bake as the grid.
	fill := colors.Unshaded3(color)
	outline := colors.Unshaded3(math32.Color{})

	font.SetColor(&math32.Color4{R: outline.R, G: outline.G, B: outline.B, A: 1})
	stamp := font.DrawText(msg)
	w, h := stamp.Bounds().Dx(), stamp.Bounds().Dy()

	out := image.NewRGBA(image.Rect(0, 0, w+2*radius, h+2*radius))
	for _, p := range outlineOffsets(radius) {
		draw.Draw(out, stamp.Bounds().Add(image.Point{X: radius + p.X, Y: radius + p.Y}),
			stamp, image.Point{}, draw.Over)
	}

	font.SetColor(&math32.Color4{R: fill.R, G: fill.G, B: fill.B, A: 1})
	top := font.DrawText(msg)
	draw.Draw(out, top.Bounds().Add(image.Point{X: radius, Y: radius}), top, image.Point{}, draw.Over)

	return out
}
