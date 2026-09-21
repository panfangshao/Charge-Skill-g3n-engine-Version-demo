// Command chargedemo is an rbfx-free, Godot-free reimplementation of
// ../aaa/scenes/charge_demo.tscn on top of g3n.
//
// It deliberately does not use g3n's app package: that pulls in the audio
// subsystem, which links OpenAL32.dll and libvorbis.dll at load time. Neither
// ships with Windows and this demo has no sound, so the window, renderer and
// loop are assembled here instead -- the same handful of calls app.App() makes,
// minus the audio device.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"runtime"
	"time"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"

	"chargedemo/internal/config"
	"chargedemo/internal/effects"
	"chargedemo/internal/fontutil"
	"chargedemo/internal/hud"
	"chargedemo/internal/labels"
	"chargedemo/internal/shaders"
	"chargedemo/internal/sim"
	"chargedemo/internal/world"
)

// Dev capture switches, for diffing this build against the Godot one -- see
// tools/compare.py and tools/capture-godot.cmd. Counted in physics ticks rather
// than rendered frames, because that is the clock both engines share: frame
// rates differ, 60 Hz ticks do not.
var (
	flagScreenshot = flag.String("screenshot", "", "save a PNG after -ticks physics ticks, then quit")
	flagTicks      = flag.Int("ticks", 30, "physics tick to capture on")
	flagPress      = flag.String("press", "", "key to fire once: space or r")
	flagPressAt    = flag.Int("press-at", 0, "physics tick on which to fire -press")
)

func main() {
	flag.Parse()
	// OpenGL calls have to stay on the thread that created the context.
	runtime.LockOSThread()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := window.Init(config.BaseWidth, config.BaseHeight, config.WindowTitle); err != nil {
		return fmt.Errorf("window: %w", err)
	}
	win := window.Get().(*window.GlfwWindow)
	defer win.Destroy()

	gs := win.Gls()
	rend := renderer.NewRenderer(gs)
	if err := rend.AddDefaultShaders(); err != nil {
		return fmt.Errorf("default shaders: %w", err)
	}
	shaders.Register(rend)

	font, fontPath, err := fontutil.Load(config.FontCandidates, config.StatusFontSize)
	if err != nil {
		return fmt.Errorf("font: %w", err)
	}
	fmt.Println("using font", fontPath)

	scene := core.NewNode()
	scenery := world.Create()
	scene.Add(scenery.Scene)

	labels.Attach(scenery.Player, font, config.PlayerLabel, config.PlayerLabelColor)
	labels.Attach(scenery.Target, font, config.TargetLabel, config.TargetLabelColor)

	state := sim.New()
	pool := effects.New(scenery.Scene, scenery.TrailGeom, scenery.ImpactGeom)
	panel := hud.New(scene, font, state.StartCharge)

	// KeyState is g3n's held-key tracker; the GLFW handle underneath takes
	// its own key type, so go through this instead.
	keys := window.NewKeyState(win)
	defer keys.Dispose()

	// Godot's _unhandled_key_input: event driven, and it ignores auto-repeat.
	win.Subscribe(window.OnKeyDown, func(_ string, ev interface{}) {
		switch ev.(*window.KeyEvent).Key {
		case window.KeySpace:
			state.StartCharge()
		case window.KeyR:
			state.ResetDemo()
		case window.KeyEscape:
			win.SetShouldClose(true)
		}
	})

	// The Godot project leaves anti-aliasing off; g3n asks GLFW for 8x samples
	// unconditionally, so turn it back off here.
	gs.Disable(gls.MULTISAMPLE)
	gs.ClearColor(config.BgColor.R, config.BgColor.G, config.BgColor.B, 1)

	loop := &appLoop{
		win: win, gs: gs, rend: rend, keys: keys,
		scene: scene, scenery: scenery, state: state, pool: pool, hud: panel,
	}
	return loop.run()
}

type appLoop struct {
	win     *window.GlfwWindow
	gs      *gls.GLS
	rend    *renderer.Renderer
	keys    *window.KeyState
	scene   *core.Node
	scenery *world.Scenery
	state   *sim.State
	pool    *effects.Pool
	hud     *hud.Layout

	accumulator float32
	ticks       int
	captured    bool
}

func (a *appLoop) run() error {
	last := time.Now()
	for !a.win.ShouldClose() {
		now := time.Now()
		delta := float32(now.Sub(last).Seconds())
		last = now

		width, height := a.win.GetFramebufferSize()
		a.gs.Viewport(0, 0, int32(width), int32(height))
		a.scenery.Camera.Resize(width, height)
		a.hud.Resize(width, height)

		a.step(delta)

		a.gs.Clear(gls.COLOR_BUFFER_BIT | gls.DEPTH_BUFFER_BIT)
		if err := a.rend.Render(a.scene, a.scenery.Camera); err != nil {
			return fmt.Errorf("render: %w", err)
		}

		if !a.captured && *flagScreenshot != "" && a.ticks >= *flagTicks {
			a.captured = true
			a.dumpState()
			if err := savePNG(a.gs, width, height, *flagScreenshot); err != nil {
				return err
			}
			a.win.SetShouldClose(true)
		}

		a.win.SwapBuffers()
		a.win.PollEvents()
	}
	return nil
}

// step runs Godot's fixed 60 Hz _physics_process, capped so a stalled frame
// cannot spiral.
func (a *appLoop) step(delta float32) {
	a.accumulator += delta
	movement := a.readMovement()

	steps := 0
	for a.accumulator >= config.PhysicsStep && steps < config.MaxPhysicsSteps {
		a.ticks++
		a.state.Step(config.PhysicsStep, movement)
		// After the step, because Godot runs an autoload's _physics_process
		// after the current scene's, and tools/capture-godot.cmd presses from an
		// autoload. Keeping the two harnesses on the same side of the tick is
		// what makes the state dumps comparable.
		if *flagPress != "" && a.ticks == *flagPressAt {
			switch *flagPress {
			case "space":
				a.state.StartCharge()
			case "r":
				a.state.ResetDemo()
			}
		}
		a.accumulator -= config.PhysicsStep
		steps++
	}
	if steps == config.MaxPhysicsSteps {
		a.accumulator = 0
	}

	a.scenery.SyncFrame(
		math32.Vector3{X: a.state.Player.X, Y: a.state.Player.Y, Z: a.state.Player.Z},
		math32.Vector3{X: a.state.Target.X, Y: a.state.Target.Y, Z: a.state.Target.Z})

	for _, e := range a.state.PendingEffects {
		a.pool.Spawn(math32.Vector3{X: e.X, Y: e.Y, Z: e.Z}, e.Impact)
	}
	a.state.PendingEffects = a.state.PendingEffects[:0]
	a.pool.Animate(delta)

	a.hud.Update(a.state)
}

// readMovement is the -1/0/+1 input pair charge_demo.gd builds from the
// movement keys. Godot reads physical keys; GLFW key codes are physical too.
func (a *appLoop) readMovement() sim.Vec2 {
	held := func(x, y window.Key) float32 {
		if a.keys.Pressed(x) || a.keys.Pressed(y) {
			return 1
		}
		return 0
	}
	return sim.Vec2{
		X: held(window.KeyD, window.KeyRight) - held(window.KeyA, window.KeyLeft),
		Z: held(window.KeyS, window.KeyDown) - held(window.KeyW, window.KeyUp),
	}
}

// dumpState prints the simulation state, so the two builds can be diffed
// numerically and not only by eye.
func (a *appLoop) dumpState() {
	charging := 0
	if a.state.Charging {
		charging = 1
	}
	fmt.Printf("STATE tick=%d player=(%.4f, %.4f, %.4f) target=(%.4f, %.4f, %.4f) hits=%d cooldown=%.4f charging=%d\n",
		a.ticks,
		a.state.Player.X, a.state.Player.Y, a.state.Player.Z,
		a.state.Target.X, a.state.Target.Y, a.state.Target.Z,
		a.state.Hits, a.state.CooldownLeft, charging)
}

// savePNG reads the frame that was just rendered. It runs before SwapBuffers,
// so the back buffer still holds it.
func savePNG(gs *gls.GLS, width, height int, path string) error {
	buf := gs.ReadPixels(0, 0, width, height, gls.RGBA, gls.UNSIGNED_BYTE)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// OpenGL hands back rows bottom-up.
	stride := width * 4
	for y := 0; y < height; y++ {
		copy(img.Pix[y*stride:(y+1)*stride], buf[(height-1-y)*stride:(height-y)*stride])
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
