// Package hud is the CanvasLayer HUD, rebuilt with g3n's gui package.
//
// Godot lays the HUD out in 1280x720 design pixels with stretch mode
// canvas_items / aspect expand: the whole canvas is scaled by
// min(w/1280, h/720) and whichever axis has room left over simply gains logical
// space, which the anchored panel grows into. g3n's gui has no global scale, so
// Layout multiplies every offset and font size by that factor itself and
// rebuilds on resize -- which it has to do anyway, because Godot's
// StyleBoxFlat has no g3n counterpart and the rounded panel and button are
// rasterised at their exact pixel size (see stylebox.go).
package hud

import (
	"image"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/text"
	"github.com/g3n/engine/texture"

	"chargedemo/internal/colors"
	"chargedemo/internal/config"
	"chargedemo/internal/sim"
)

// Layout owns the HUD widgets.
type Layout struct {
	scene *core.Node
	font  *text.Font

	root         *gui.Panel
	title        *gui.Label
	instructions *gui.Label
	panel        *gui.Image
	status       *gui.Label
	details      *gui.Label
	button       *gui.Image
	buttonLabel  *gui.Label

	width, height int
	scale         float32

	statusText string
	buttonText string
	disabled   bool
	hovered    bool
	pressed    bool

	// onCharge is called when the button is released over itself, which is what
	// Godot's ACTION_MODE_BUTTON_RELEASE does.
	onCharge func()
}

// New creates the HUD. It is laid out by the first Resize.
func New(scene *core.Node, font *text.Font, onCharge func()) *Layout {
	l := &Layout{
		scene:      scene,
		font:       font,
		onCharge:   onCharge,
		statusText: config.MsgReady,
		buttonText: config.ButtonIdle,
	}
	gui.Manager().Set(scene)
	return l
}

// panelBackground is the flattened panel fill. Godot's Panel is 94% opaque over
// the clear colour and is composited in sRGB space; doing that here keeps GL
// blending out of the picture. See colors.OverSrgb.
func panelBackground() math32.Color {
	return colors.OverSrgb(config.PanelBg, config.BgColor)
}

// Resize rebuilds the HUD for a new framebuffer size.
//
// Godot: content_scale_factor = min(w / 1280, h / 720), and the leftover space
// on the other axis becomes extra logical room.
func (l *Layout) Resize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	scale := math32.Min(float32(width)/config.BaseWidth, float32(height)/config.BaseHeight)
	if width == l.width && height == l.height {
		return
	}
	l.width, l.height, l.scale = width, height, scale

	l.clear()
	l.build()
}

func (l *Layout) clear() {
	if l.root != nil {
		l.scene.Remove(l.root)
		l.root = nil
	}
}

// px scales a design-pixel measurement to the current window.
func (l *Layout) px(v float32) float32 { return v * l.scale }

func (l *Layout) build() {
	l.root = gui.NewPanel(float32(l.width), float32(l.height))
	l.root.SetColor4(&math32.Color4{})
	l.root.SetRenderable(false)
	l.root.SetPosition(0, 0)
	l.scene.Add(l.root)

	// --- Title / Instructions ----------------------------------------------
	l.title = l.newLabel(config.Title, config.TitleFontSize, config.ControlFontColor)
	l.title.SetPosition(l.px(config.TitlePos.X), l.px(config.TitlePos.Y))
	l.root.Add(l.title)

	l.instructions = l.newLabel(config.Instructions, config.InstructionsFontSize, config.InstructionsColor)
	l.instructions.SetPosition(l.px(config.InstructionsPos.X), l.px(config.InstructionsPos.Y))
	l.root.Add(l.instructions)

	// --- Bottom panel -------------------------------------------------------
	// anchors_preset 12: pinned to the bottom edge, stretched horizontally.
	panelW := int(float32(l.width) - 2*l.px(config.PanelMarginX))
	panelH := int(l.px(config.PanelHeight))
	panelY := float32(l.height) - l.px(config.PanelBottomFromBottom) - float32(panelH)

	l.panel = gui.NewImageFromRGBA(roundedRect(panelW, panelH,
		int(l.px(config.PanelCornerRadius)), maxInt(int(l.px(config.PanelBorderWidth)), 1),
		panelBackground(), config.PanelBorder, config.BgColor))
	l.panel.SetPosition(l.px(config.PanelMarginX), panelY)
	l.root.Add(l.panel)

	l.status = l.newLabel(l.statusText, config.StatusFontSize, config.ControlFontColor)
	l.status.SetPosition(l.px(config.StatusPos.X), l.px(config.StatusPos.Y))
	l.panel.Add(l.status)

	l.details = l.newLabel(config.Details, config.DetailsFontSize, config.DetailsColor)
	l.details.SetPosition(l.px(config.DetailsPos.X), l.px(config.DetailsPos.Y))
	l.panel.Add(l.details)

	// --- Charge button ------------------------------------------------------
	buttonW := int(l.px(config.ButtonWidth))
	buttonH := int(l.px(config.ButtonHeight))
	l.button = gui.NewImageFromRGBA(l.buttonImage(buttonW, buttonH))
	l.button.SetPosition(
		float32(panelW)-l.px(config.ButtonRightInset)-float32(buttonW),
		l.px(config.ButtonTop))
	l.panel.Add(l.button)

	l.buttonLabel = l.newLabel(l.buttonText, config.ButtonFontSize, config.ControlFontColor)
	l.centreButtonLabel(float32(buttonW), float32(buttonH))
	l.button.Add(l.buttonLabel)

	l.subscribeButton()
	l.refreshButton()
}

func (l *Layout) newLabel(msg string, size float32, color math32.Color) *gui.Label {
	label := gui.NewLabelWithFont(msg, l.font)
	// gui.Label pads itself by 2px vertically; Godot's Label does not.
	label.SetPaddings(0, 0, 0, 0)
	label.SetFontSize(float64(l.px(size)))
	label.SetColor(&color)
	return label
}

func (l *Layout) centreButtonLabel(buttonW, buttonH float32) {
	w, h := l.buttonLabel.ContentWidth(), l.buttonLabel.ContentHeight()
	l.buttonLabel.SetPosition((buttonW-w)/2, (buttonH-h)/2)
}

// buttonImage rasterises the button in whichever of Godot's four default-theme
// states it is currently in.
func (l *Layout) buttonImage(w, h int) *image.RGBA {
	state := config.ButtonNormalBg
	switch {
	case l.disabled:
		state = config.ButtonDisabledBg
	case l.pressed:
		state = config.ButtonPressedBg
	case l.hovered:
		state = config.ButtonHoverBg
	}
	behind := panelBackground()
	return roundedRect(w, h, int(l.px(config.ButtonCornerRadius)), 0,
		colors.OverSrgb(state, behind), behind, behind)
}

func (l *Layout) subscribeButton() {
	l.button.Subscribe(gui.OnCursorEnter, func(string, interface{}) {
		l.hovered = true
		l.refreshButton()
	})
	l.button.Subscribe(gui.OnCursorLeave, func(string, interface{}) {
		l.hovered = false
		l.pressed = false
		l.refreshButton()
	})
	l.button.Subscribe(gui.OnMouseDown, func(string, interface{}) {
		if l.disabled {
			return
		}
		l.pressed = true
		l.refreshButton()
	})
	l.button.Subscribe(gui.OnMouseUp, func(string, interface{}) {
		wasPressed := l.pressed
		l.pressed = false
		l.refreshButton()
		// Godot's Button fires on release over itself, not on push.
		if wasPressed && !l.disabled && l.onCharge != nil {
			l.onCharge()
		}
	})
}

func (l *Layout) refreshButton() {
	if l.button == nil {
		return
	}
	w := int(l.button.ContentWidth())
	h := int(l.button.ContentHeight())
	l.button.SetTexture(texture.NewTexture2DFromRGBA(l.buttonImage(w, h)))

	color := config.ControlFontColor
	if l.disabled {
		c := config.ButtonFontDisabledColor
		// Godot fades the label rather than recolouring it; the button sits on
		// a known fill, so fade it against that.
		color = colors.OverSrgb(c, colors.OverSrgb(config.ButtonDisabledBg, panelBackground()))
	}
	l.buttonLabel.SetColor(&color)
}

// Update pushes the simulation's state onto the widgets, as the tail of
// _physics_process does.
func (l *Layout) Update(state *sim.State) {
	if l.root == nil {
		return
	}
	if state.Status != l.statusText {
		l.statusText = state.Status
		l.status.SetText(l.statusText)
	}
	if text := state.ButtonText(); text != l.buttonText {
		l.buttonText = text
		l.buttonLabel.SetText(text)
		l.centreButtonLabel(l.button.ContentWidth(), l.button.ContentHeight())
	}
	if disabled := state.ButtonDisabled(); disabled != l.disabled {
		l.disabled = disabled
		l.refreshButton()
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
