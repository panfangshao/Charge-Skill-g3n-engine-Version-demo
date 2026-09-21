package hud

import (
	"image"
	"image/color"
	"math"

	"github.com/g3n/engine/math32"
)

// roundedRect rasterises Godot's StyleBoxFlat at an exact pixel size.
//
// The corners are filled with `behind` rather than left transparent. g3n's
// panel shader composites a texture over the panel's content colour and then
// un-premultiplies by dividing by the result alpha, which is a division by zero
// wherever the texture is fully transparent. Baking the background in avoids
// that, and costs nothing here: the HUD panel and the button both sit on the
// flat clear colour, and Godot's own panel is composited against it too (see
// colors.OverSrgb).
func roundedRect(width, height, radius, borderWidth int, fill, border, behind math32.Color) *image.RGBA {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	// A radius larger than half the shorter side would fold the corners over.
	if maxRadius := min(width, height) / 2; radius > maxRadius {
		radius = maxRadius
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	halfW := float64(width) / 2
	halfH := float64(height) / 2

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			px := float64(x) + 0.5 - halfW
			py := float64(y) + 0.5 - halfH

			outer := roundedRectDistance(px, py, halfW, halfH, float64(radius))
			inner := roundedRectDistance(px, py,
				halfW-float64(borderWidth), halfH-float64(borderWidth),
				math.Max(float64(radius-borderWidth), 0))

			outerCoverage := coverage(outer)
			innerCoverage := coverage(inner)

			// fill inside the inner edge, border between the two edges, and
			// whatever is behind the panel outside the outer edge.
			c := blend(behind, border, outerCoverage)
			if borderWidth <= 0 {
				c = blend(behind, fill, outerCoverage)
			} else {
				c = blend(c, fill, innerCoverage)
			}
			img.Set(x, y, color.RGBA{
				R: toByte(c.R),
				G: toByte(c.G),
				B: toByte(c.B),
				A: 255,
			})
		}
	}
	return img
}

// roundedRectDistance is the signed distance to a rounded rectangle centred on
// the origin, negative inside.
func roundedRectDistance(x, y, halfX, halfY, radius float64) float64 {
	qx := math.Abs(x) - (halfX - radius)
	qy := math.Abs(y) - (halfY - radius)
	outside := math.Hypot(math.Max(qx, 0), math.Max(qy, 0))
	inside := math.Min(math.Max(qx, qy), 0)
	return outside + inside - radius
}

// coverage of a pixel whose centre sits `distance` from the edge, antialiased
// over one pixel.
func coverage(distance float64) float64 {
	v := 0.5 - distance
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func blend(under, over math32.Color, alpha float64) math32.Color {
	a := float32(alpha)
	return math32.Color{
		R: under.R*(1-a) + over.R*a,
		G: under.G*(1-a) + over.G*a,
		B: under.B*(1-a) + over.B*a,
	}
}

func toByte(v float32) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
