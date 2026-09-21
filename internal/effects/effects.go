// Package effects is the charge trail and the impact puff.
//
// spawn_effect() in Godot creates an unshaded transparent SphereMesh and runs a
// parallel Tween that scales it and fades its alpha to zero over 0.3 s (linear
// interpolation, Godot's default TRANS_LINEAR), then frees it. The Tween runs
// on the frame clock, not the physics clock, so Animate is driven from the
// variable-rate update.
package effects

import (
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/math32"

	"chargedemo/internal/config"
	"chargedemo/internal/shaders"
)

type puff struct {
	node       *core.Node
	material   *shaders.Unlit
	elapsed    float32
	startAlpha float32
	endScale   float32
	baseColor  math32.Color4
}

// Pool owns the live puffs.
type Pool struct {
	scene  *core.Node
	trail  *geometry.Geometry
	impact *geometry.Geometry
	puffs  []*puff
}

// New returns a pool that parents its puffs to scene, as the GDScript's
// add_child(effect) on the root Node3D does.
func New(scene *core.Node, trail, impact *geometry.Geometry) *Pool {
	return &Pool{scene: scene, trail: trail, impact: impact}
}

// Spawn is spawn_effect().
func (p *Pool) Spawn(at math32.Vector3, impact bool) {
	color := config.TrailColor
	endScale := float32(config.TrailEndScale)
	geom := p.trail
	if impact {
		color = config.ImpactColor
		endScale = config.ImpactEndScale
		geom = p.impact
	}

	// Each puff fades on its own schedule, so each needs its own material.
	mat := shaders.NewUnlit(color, true)
	mesh := graphic.NewMesh(geom, mat)
	mesh.SetPositionVec(&at)
	p.scene.Add(mesh)

	p.puffs = append(p.puffs, &puff{
		node:       &mesh.Node,
		material:   mat,
		startAlpha: color.A,
		endScale:   endScale,
		baseColor:  color,
	})
}

// Animate advances every live puff and frees the finished ones.
func (p *Pool) Animate(timeStep float32) {
	live := p.puffs[:0]
	for _, e := range p.puffs {
		e.elapsed += timeStep
		t := e.elapsed / config.EffectDuration
		if t >= 1 {
			p.scene.Remove(e.node.GetNode())
			continue
		}
		scale := 1 + (e.endScale-1)*t
		e.node.SetScale(scale, scale, scale)

		faded := e.baseColor
		faded.A = e.startAlpha * (1 - t)
		e.material.SetColor(faded)

		live = append(live, e)
	}
	p.puffs = live
}
