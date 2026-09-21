package world

import (
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/math32"

	"chargedemo/internal/config"
)

// CameraRig is the scene's Camera3D.
type CameraRig struct {
	*camera.Camera
}

func newCameraRig() *CameraRig {
	// Godot's KEEP_HEIGHT: `size` is the vertical extent, which is exactly what
	// g3n's Vertical axis means.
	cam := camera.NewOrthographic(config.BaseWidth/config.BaseHeight, 0.05, 4000, config.CameraOrthoHeight, camera.Vertical)
	cam.SetPositionVec(&config.CameraPos)
	cam.SetRotationX(math32.DegToRad(config.CameraPitchDeg))
	return &CameraRig{Camera: cam}
}

// Resize keeps the horizontal extent following the window, which is what
// Godot's KEEP_HEIGHT does.
func (c *CameraRig) Resize(width, height int) {
	if height <= 0 {
		return
	}
	c.SetAspect(float32(width) / float32(height))
}
