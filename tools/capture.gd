# Dev-only autoload, injected into a *copy* of the Godot project by
# tools/capture-godot.cmd. It mirrors the --screenshot / --ticks / --press
# switches of the rbfx build so the two can be driven identically:
# both count physics ticks, not rendered frames.
extends Node

var ticks := 0
var capture_tick := 30
var out_path := ""
var press := ""
var press_tick := 0
var done := false


func _ready() -> void:
	# Keep running while the tree is paused, so the capture can freeze the
	# simulation and still finish.
	process_mode = Node.PROCESS_MODE_ALWAYS
	for arg in OS.get_cmdline_user_args():
		var parts := arg.split("=", true, 1)
		if parts.size() < 2:
			continue
		match parts[0]:
			"--screenshot": out_path = parts[1]
			"--ticks": capture_tick = int(parts[1])
			"--press": press = parts[1]
			"--press-at": press_tick = int(parts[1])


func _physics_process(_delta: float) -> void:
	if done:
		return
	ticks += 1
	var demo := get_tree().current_scene
	if press != "" and ticks == press_tick:
		if press == "space":
			demo.start_charge()
		elif press == "r":
			demo.reset_demo()
	if ticks >= capture_tick:
		done = true
		print("STATE tick=%d player=(%.4f, %.4f, %.4f) target=(%.4f, %.4f, %.4f) hits=%d cooldown=%.4f charging=%d" % [
			ticks,
			demo.player.position.x, demo.player.position.y, demo.player.position.z,
			demo.target.position.x, demo.target.position.y, demo.target.position.z,
			demo.hits, demo.cooldown, 1 if demo.charging else 0])
		if out_path != "":
			# Pause first: without it another physics tick slips in before the
			# next draw, and the image ends up one tick ahead of the STATE line.
			get_tree().paused = true
			await RenderingServer.frame_post_draw
			get_viewport().get_texture().get_image().save_png(out_path)
		get_tree().quit()
