package config

// Every user-visible string, copied from the Godot project (../aaa).

const (
	// project.godot: application/config/name
	WindowTitle = "新建游戏项目"

	// Label3D nodes
	PlayerLabel = "玩家 / YOU"
	TargetLabel = "冲锋目标"

	// HUD labels
	Title        = "冲锋训练场  /  CHARGE"
	Instructions = "WASD / 方向键移动    ·    空格 / 点击按钮冲锋    ·    R 重置"
	Details      = "自动瞄准目标 · 最大距离 14 米 · 命中击退 · 冷却 3 秒"
	ButtonIdle   = "冲锋 [空格]"

	// charge_demo.gd status messages
	MsgReady      = "技能就绪：向红色目标发起冲锋"
	MsgCharging   = "冲锋中！"
	MsgChargeEnd  = "冲锋结束 · 等待冷却后再次尝试"
	MsgOutOfRange = "目标超过 14 米，请先靠近红色角色"
	MsgReset      = "已重置 · 技能就绪：向红色目标发起冲锋"

	// "命中！目标被击退  ·  累计命中 %d 次"
	MsgHitFormat = "命中！目标被击退  ·  累计命中 %d 次"
	// "冷却 %.1f 秒"
	MsgCooldownFormat = "冷却 %.1f 秒"
)
