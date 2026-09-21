package sim

import (
	"math"
	"testing"

	"chargedemo/internal/config"
)

// The reference figures come from running ../aaa under tools/capture-godot.cmd,
// which presses the same key on the same physics tick.

// run advances the state by `ticks` physics steps with a constant input, as the
// app's fixed-rate loop does. pressAt fires the charge after that tick's step,
// matching both capture harnesses.
func run(s *State, ticks int, movement Vec2, pressAt int) {
	for tick := 1; tick <= ticks; tick++ {
		s.Step(config.PhysicsStep, movement)
		if tick == pressAt {
			s.StartCharge()
		}
	}
}

func near(t *testing.T, what string, got, want, tol float32) {
	t.Helper()
	if math.Abs(float64(got-want)) > float64(tol) {
		t.Errorf("%s: got %.4f, want %.4f +/- %.4f", what, got, want, tol)
	}
}

func TestSpawnState(t *testing.T) {
	s := New()
	near(t, "player x", s.Player.X, -5, 1e-5)
	near(t, "target x", s.Target.X, 4, 1e-5)
	if s.Status != config.MsgReady {
		t.Errorf("status: got %q", s.Status)
	}
	if s.ButtonDisabled() {
		t.Error("button should start enabled")
	}
	if s.ButtonText() != config.ButtonIdle {
		t.Errorf("button label: got %q", s.ButtonText())
	}
}

// TestChargeFromSpawn is the reference run.
//
//	Godot: player=(3.0977, 0.9000, 0.0000) target=(7.7000, ...) hits=1 cooldown=1.6667
func TestChargeFromSpawn(t *testing.T) {
	s := New()
	run(s, 90, Vec2{}, 10)

	near(t, "player x matches Godot", s.Player.X, 3.0977, 0.01)
	near(t, "player stays on z = 0", s.Player.Z, 0, 1e-4)
	near(t, "target x matches Godot", s.Target.X, 7.7, 0.01)
	near(t, "cooldown matches Godot", s.CooldownLeft, 1.6667, 0.001)
	if s.Hits != 1 {
		t.Errorf("hits: got %d, want 1", s.Hits)
	}
	if s.Charging {
		t.Error("charge should have finished")
	}
	if !s.ButtonDisabled() {
		t.Error("button should be disabled while cooling down")
	}
}

// TestMidCharge checks the halfway state Godot reports on tick 20.
func TestMidCharge(t *testing.T) {
	s := New()
	run(s, 20, Vec2{}, 10)
	near(t, "player x matches Godot", s.Player.X, -1.0, 1e-3)
	near(t, "cooldown matches Godot", s.CooldownLeft, 2.8333, 1e-3)
	if !s.Charging {
		t.Error("should still be charging")
	}
}

func TestCapsulesDoNotOverlap(t *testing.T) {
	s := New()
	run(s, 90, Vec2{}, 10)
	if gap := s.Target.X - s.Player.X; gap < 2*config.CapsuleRadius-1e-3 {
		t.Errorf("capsules overlap: gap %.4f", gap)
	}
}

func TestOutOfRangeIsRefused(t *testing.T) {
	s := New()
	// Walk left into the arena wall, which puts the target well past 14 m.
	run(s, 180, Vec2{X: -1}, -1)
	near(t, "player clamped to the arena", s.Player.X, -config.ArenaX, 1e-4)

	s.StartCharge()
	if s.Charging {
		t.Error("charge should have been refused")
	}
	if s.Status != config.MsgOutOfRange {
		t.Errorf("status: got %q", s.Status)
	}
	near(t, "no cooldown spent", s.CooldownLeft, 0, 1e-6)
}

func TestCooldownBlocksASecondCharge(t *testing.T) {
	s := New()
	run(s, 30, Vec2{}, 1)
	hits := s.Hits

	s.StartCharge()
	if s.Hits != hits {
		t.Error("second charge should do nothing")
	}
	if s.Status == config.MsgCharging {
		t.Error("status should not say charging")
	}

	run(s, 181, Vec2{}, -1)
	near(t, "cooldown expires after 3 s", s.CooldownLeft, 0, 1e-6)
	if s.ButtonDisabled() {
		t.Error("button should be enabled again")
	}
}

func TestResetRestoresTheSpawn(t *testing.T) {
	s := New()
	run(s, 90, Vec2{}, 10)
	s.ResetDemo()

	near(t, "player back at spawn", s.Player.X, -5, 1e-5)
	near(t, "target back at spawn", s.Target.X, 4, 1e-5)
	near(t, "cooldown cleared", s.CooldownLeft, 0, 1e-6)
	if s.Hits != 0 {
		t.Errorf("hits: got %d, want 0", s.Hits)
	}
	if s.Status != config.MsgReset {
		t.Errorf("status: got %q", s.Status)
	}
}

// TestWalkingSpeed: MOVE_SPEED is 5 m/s, so a second of held input is 5 m.
func TestWalkingSpeed(t *testing.T) {
	s := New()
	run(s, 60, Vec2{Z: -1}, -1)
	near(t, "one second of W is 5 m", s.Player.Z, -5, 1e-3)
	near(t, "pinned to the actor height", s.Player.Y, config.ActorY, 1e-6)
}

// TestDiagonalIsNotFaster covers Godot's limit_length.
func TestDiagonalIsNotFaster(t *testing.T) {
	s := New()
	run(s, 60, Vec2{X: 1, Z: -1}, -1)
	moved := Vec2{s.Player.X - config.PlayerStart.X, s.Player.Z - config.PlayerStart.Z}
	near(t, "diagonal covers the same distance as straight", moved.Len(), 5, 1e-3)
}

func TestArenaBounds(t *testing.T) {
	s := New()
	run(s, 600, Vec2{X: 1, Z: 1}, -1)
	near(t, "clamped in +x", s.Player.X, config.ArenaX, 1e-4)
	near(t, "clamped in +z", s.Player.Z, config.ArenaZ, 1e-4)
}

func TestCooldownLabel(t *testing.T) {
	s := New()
	run(s, 1, Vec2{}, 1)
	// Godot formats this as "冷却 %.1f 秒".
	if got, want := s.ButtonText(), "冷却 3.0 秒"; got != want {
		t.Errorf("button label: got %q, want %q", got, want)
	}
}

func TestChargeQueuesTrailEffects(t *testing.T) {
	s := New()
	run(s, 20, Vec2{}, 10)
	if len(s.PendingEffects) == 0 {
		t.Fatal("charging should have queued trail puffs")
	}
	for _, e := range s.PendingEffects {
		if e.Impact {
			t.Error("no impact puff expected mid-charge")
		}
	}
}
