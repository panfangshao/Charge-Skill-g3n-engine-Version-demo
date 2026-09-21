// Package sim is charge_demo.gd with no engine underneath it.
//
// Keeping the simulation free of g3n means Step can be driven from a scripted
// input tape in the tests, and that the numbers can be compared directly
// against the ones the Godot build prints.
package sim

import "math"

// Vec2 is the XZ plane. The simulation never needs anything else: both actors
// are pinned to y = ActorY every frame by clampToArena.
type Vec2 struct{ X, Z float32 }

func (v Vec2) Add(o Vec2) Vec2      { return Vec2{v.X + o.X, v.Z + o.Z} }
func (v Vec2) Sub(o Vec2) Vec2      { return Vec2{v.X - o.X, v.Z - o.Z} }
func (v Vec2) Scale(s float32) Vec2 { return Vec2{v.X * s, v.Z * s} }
func (v Vec2) Dot(o Vec2) float32   { return v.X*o.X + v.Z*o.Z }
func (v Vec2) LenSq() float32       { return v.X*v.X + v.Z*v.Z }
func (v Vec2) Len() float32         { return float32(math.Sqrt(float64(v.LenSq()))) }

// Body is the part of Godot's CharacterBody3D this demo actually uses.
//
// Both actors are capsules of identical radius and height standing on a flat
// static box, which reduces the whole simulation to circle-vs-circle sweeping
// in the XZ plane. That is reproduced exactly here -- including the slide
// response and the per-frame collision flag the gameplay script inspects --
// rather than handed to a rigid-body engine, because g3n has no physics at all
// and Godot ran this scene on Jolt. Neither would reproduce the other's
// contact solver; the reduced problem has a closed-form answer both agree on.
type Body struct {
	X, Y, Z  float32
	Velocity Vec2
	Radius   float32
}

func (b *Body) xz() Vec2 { return Vec2{b.X, b.Z} }

func (b *Body) setXZ(v Vec2) {
	b.X = v.X
	b.Z = v.Z
}

const (
	maxSlides  = 6
	safeMargin = 0.001
)

// sweep returns the earliest t in [0, 1] at which the moving circle touches the
// static one, and whether there is such a t.
func sweep(from, motion, obstacle Vec2, sumRadius float32) (float32, bool) {
	d := from.Sub(obstacle)
	a := motion.LenSq()
	if a <= 1e-12 {
		return 0, false
	}
	b := 2 * d.Dot(motion)
	c := d.LenSq() - sumRadius*sumRadius
	disc := b*b - 4*a*c
	if disc < 0 {
		return 0, false
	}
	t := (-b - float32(math.Sqrt(float64(disc)))) / (2 * a)
	if t < 0 || t > 1 {
		return 0, false
	}
	return t, true
}

func depenetrate(body *Body, obstacle *Body) {
	sumR := body.Radius + obstacle.Radius
	delta := body.xz().Sub(obstacle.xz())
	dist := delta.Len()
	if dist >= sumR {
		return
	}
	if dist <= 1e-6 {
		delta = Vec2{X: 1}
		dist = 1
	}
	push := (sumR - dist) + safeMargin
	body.setXZ(body.xz().Add(delta.Scale(push / dist)))
}

// moveAndSlide advances body by its velocity, sliding along obstacle if it runs
// into it. It reports whether a collision happened this frame, which is Godot's
// get_slide_collision_count / get_slide_collision.
func moveAndSlide(body *Body, obstacle *Body, delta float32) bool {
	depenetrate(body, obstacle)

	motion := body.Velocity.Scale(delta)
	collided := false

	for i := 0; i < maxSlides; i++ {
		if motion.LenSq() <= 1e-14 {
			break
		}

		t, hit := sweep(body.xz(), motion, obstacle.xz(), body.Radius+obstacle.Radius)
		if !hit {
			body.setXZ(body.xz().Add(motion))
			break
		}

		body.setXZ(body.xz().Add(motion.Scale(t)))
		collided = true

		normal := body.xz().Sub(obstacle.xz())
		length := normal.Len()
		if length <= 1e-6 {
			break
		}
		normal = normal.Scale(1 / length)

		body.setXZ(body.xz().Add(normal.Scale(safeMargin)))

		remaining := motion.Scale(1 - t)
		motion = remaining.Sub(normal.Scale(remaining.Dot(normal)))
	}

	return collided
}

// moveToward is Godot's Vector3.move_toward, for the 2D knockback vector.
func moveToward(value, target Vec2, delta float32) Vec2 {
	toTarget := target.Sub(value)
	dist := toTarget.Len()
	if dist <= delta || dist < 1e-9 {
		return target
	}
	return value.Add(toTarget.Scale(delta / dist))
}

// limitLength is Godot's Vector2.limit_length.
func limitLength(v Vec2, limit float32) Vec2 {
	length := v.Len()
	if length > limit && length > 0 {
		return v.Scale(limit / length)
	}
	return v
}

func clampf(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
