package sim

import (
	"fmt"

	"chargedemo/internal/config"
)

// EffectSpawn is a queued spawn_effect() call, drained by the effect pool each
// frame.
type EffectSpawn struct {
	X, Y, Z float32
	Impact  bool
}

// State is charge_demo.gd. Step is a line-by-line port of _physics_process and
// runs on Godot's fixed 60 Hz tick; StartCharge and ResetDemo correspond to
// _unhandled_key_input, which Godot drives from input events rather than from
// the physics tick.
type State struct {
	Player         Body
	Target         Body
	CooldownLeft   float32
	Charging       bool
	ChargeDir      Vec2
	DistanceLeft   float32
	Knockback      Vec2
	TrailTimer     float32
	Hits           int
	Status         string
	PendingEffects []EffectSpawn
}

// New returns the state the scene starts in.
func New() *State {
	s := &State{Status: config.MsgReady}
	s.Player = Body{X: config.PlayerStart.X, Y: config.PlayerStart.Y, Z: config.PlayerStart.Z, Radius: config.CapsuleRadius}
	s.Target = Body{X: config.TargetStart.X, Y: config.TargetStart.Y, Z: config.TargetStart.Z, Radius: config.CapsuleRadius}
	return s
}

// clampToArena is clamp_to_arena().
func clampToArena(b *Body) {
	b.X = clampf(b.X, -config.ArenaX, config.ArenaX)
	b.Z = clampf(b.Z, -config.ArenaZ, config.ArenaZ)
	b.Y = config.ActorY
}

// Step is Godot's _physics_process. movement is the raw -1/0/+1 input pair the
// GDScript builds from the movement keys.
func (s *State) Step(delta float32, movement Vec2) {
	s.CooldownLeft = maxf(0, s.CooldownLeft-delta)

	s.Target.Velocity = s.Knockback
	player := s.Player
	moveAndSlide(&s.Target, &player, delta)
	s.Knockback = moveToward(s.Knockback, Vec2{}, config.KnockbackDecel*delta)
	clampToArena(&s.Target)

	if s.Charging {
		beforeX, beforeZ := s.Player.X, s.Player.Z
		s.Player.Velocity = s.ChargeDir.Scale(config.ChargeSpeed)
		target := s.Target
		collided := moveAndSlide(&s.Player, &target, delta)
		s.DistanceLeft -= config.ChargeSpeed * delta
		s.TrailTimer -= delta
		if s.TrailTimer <= 0 {
			s.spawnEffect(s.Player.X, s.Player.Y, s.Player.Z, false)
			s.TrailTimer = config.TrailInterval
		}
		if collided {
			s.hitTarget()
		}
		moved := Vec2{s.Player.X - beforeX, s.Player.Z - beforeZ}.Len()
		if s.Charging && (s.DistanceLeft <= 0 || moved < 0.001) {
			s.Charging = false
			s.Status = config.MsgChargeEnd
		}
	} else {
		input := limitLength(movement, 1)
		s.Player.Velocity = input.Scale(config.MoveSpeed)
		target := s.Target
		moveAndSlide(&s.Player, &target, delta)
	}

	clampToArena(&s.Player)
}

// StartCharge is start_charge().
func (s *State) StartCharge() {
	if s.Charging || s.CooldownLeft > 0 {
		return
	}

	offset := Vec2{s.Target.X - s.Player.X, s.Target.Z - s.Player.Z}
	length := offset.Len()
	if length > config.ChargeRange {
		s.Status = config.MsgOutOfRange
		return
	}

	if length > 0 {
		s.ChargeDir = offset.Scale(1 / length)
	} else {
		s.ChargeDir = Vec2{}
	}
	s.DistanceLeft = minf(length+1, config.ChargeRange)
	s.Charging = true
	s.CooldownLeft = config.Cooldown
	s.TrailTimer = 0
	s.Status = config.MsgCharging
}

// hitTarget is hit_target().
func (s *State) hitTarget() {
	s.Charging = false
	s.Knockback = s.ChargeDir.Scale(config.KnockbackSpeed)
	s.Hits++
	s.Status = fmt.Sprintf(config.MsgHitFormat, s.Hits)
	s.spawnEffect(s.Target.X, s.Target.Y, s.Target.Z, true)
}

// ResetDemo is reset_demo().
func (s *State) ResetDemo() {
	s.Charging = false
	s.CooldownLeft = 0
	s.Knockback = Vec2{}
	s.Hits = 0
	s.Player.X, s.Player.Y, s.Player.Z = config.PlayerStart.X, config.PlayerStart.Y, config.PlayerStart.Z
	s.Target.X, s.Target.Y, s.Target.Z = config.TargetStart.X, config.TargetStart.Y, config.TargetStart.Z
	s.Player.Velocity = Vec2{}
	s.Target.Velocity = Vec2{}
	s.Status = config.MsgReset
}

func (s *State) spawnEffect(x, y, z float32, impact bool) {
	s.PendingEffects = append(s.PendingEffects, EffectSpawn{X: x, Y: y, Z: z, Impact: impact})
}

// ButtonText is the label Godot rebuilds every physics frame.
func (s *State) ButtonText() string {
	if s.CooldownLeft > 0 {
		return fmt.Sprintf(config.MsgCooldownFormat, s.CooldownLeft)
	}
	return config.ButtonIdle
}

// ButtonDisabled is `cooldown > 0.0 or charging`.
func (s *State) ButtonDisabled() bool { return s.CooldownLeft > 0 || s.Charging }

func maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
