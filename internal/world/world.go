// Package world builds the 3D scene: WorldEnvironment, Sun, Camera3D, Ground,
// grid and actors.
//
// Mirrors the node tree of scenes/charge_demo.tscn plus the grid that
// add_grid() builds at runtime. g3n shares Godot's handedness -- Y up, -Z
// forward -- so every transform carries over unchanged.
//
// There is no light node: g3n's directional light drives its own Blinn-Phong
// shaders, and this port's lighting lives entirely in internal/shaders.
package world

import (
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/math32"

	"chargedemo/internal/config"
	"chargedemo/internal/meshes"
	"chargedemo/internal/shaders"
)

// Scenery is everything the rest of the app needs to reach back into.
type Scenery struct {
	Scene  *core.Node
	Camera *CameraRig
	Player *core.Node
	Target *core.Node

	// The lit materials, so the analytic shadow test can be pointed at the
	// current capsule positions each frame.
	Lit []*shaders.Lit

	// Reused by the effect puffs, which are spawned and freed constantly.
	TrailGeom  *geometry.Geometry
	ImpactGeom *geometry.Geometry
}

// Create builds the scene.
func Create() *Scenery {
	s := &Scenery{Scene: core.NewNode()}

	s.Camera = newCameraRig()
	s.Scene.Add(s.Camera.Camera)

	// --- Ground -------------------------------------------------------------
	groundMat := shaders.NewLit(config.GroundAlbedo, config.GroundMetallic, config.GroundRoughness, shaders.SkipNone)
	s.Lit = append(s.Lit, groundMat)
	ground := graphic.NewMesh(meshes.NewBox(config.GroundSize), groundMat)
	ground.SetPositionVec(&config.GroundMeshAt)
	s.Scene.Add(ground)

	s.addGrid()

	// --- Actors -------------------------------------------------------------
	capsule := meshes.NewCapsule(config.CapsuleRadius, config.CapsuleHeight, 64, 8)

	playerMat := shaders.NewLit(config.BlueAlbedo, config.BlueMetallic, config.BlueRoughness, shaders.SkipPlayer)
	s.Lit = append(s.Lit, playerMat)
	player := graphic.NewMesh(capsule, playerMat)
	player.SetPositionVec(&config.PlayerStart)
	s.Scene.Add(player)
	s.Player = &player.Node

	targetMat := shaders.NewLit(config.RedAlbedo, config.RedMetallic, config.RedRoughness, shaders.SkipTarget)
	s.Lit = append(s.Lit, targetMat)
	target := graphic.NewMesh(capsule, targetMat)
	target.SetPositionVec(&config.TargetStart)
	s.Scene.Add(target)
	s.Target = &target.Node

	s.TrailGeom = meshes.NewSphere(config.TrailRadius, config.TrailRadius*2, 32, 16)
	s.ImpactGeom = meshes.NewSphere(config.ImpactRadius, config.ImpactRadius*2, 32, 16)

	return s
}

// SyncFrame pushes the simulation's positions onto the scene and refreshes the
// per-frame shader inputs.
func (s *Scenery) SyncFrame(player, target math32.Vector3) {
	s.Player.SetPositionVec(&player)
	s.Target.SetPositionVec(&target)
	for _, m := range s.Lit {
		m.SetShadowCasters(player, target)
		m.SetCameraPosition(config.CameraPos)
	}
}

// addGrid is add_grid() / add_grid_line().
func (s *Scenery) addGrid() {
	// Unlit, and only 1.5 cm tall. Godot leaves these as shadow casters but at
	// that height they cast nothing; the analytic shadow test only knows about
	// the two capsules anyway.
	mat := shaders.NewUnlit(math32.Color4{R: config.GridColor.R, G: config.GridColor.G, B: config.GridColor.B, A: 1}, false)
	xLine := meshes.NewBox(config.GridXSize)
	zLine := meshes.NewBox(config.GridZSize)

	for x := -12; x <= 12; x += 2 {
		line := graphic.NewMesh(xLine, mat)
		line.SetPosition(float32(x), config.GridY, 0)
		s.Scene.Add(line)
	}
	for z := -9; z <= 9; z += 2 {
		line := graphic.NewMesh(zLine, mat)
		line.SetPosition(0, config.GridY, float32(z))
		s.Scene.Add(line)
	}
}
