package stat

import "github.com/Kittengarten/KittenCore/kitten/msg"

// Temperature 温度
var Temperature = 45.0

// API 名称
type API string

const Fanqie API = `fanqie` // 番茄 API

var (
	// Tracker 用于更新播报的 Bot 实例
	Tracker *msg.Handler
	// APIHOST API 主机地址
	APIHOST = make(map[API]string)
)

// Weight 自身叠猫猫体重（0.1 kg 数）
var Weight int
