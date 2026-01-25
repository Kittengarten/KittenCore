package kitten

import (
	"time"
)

type Item byte // 对上下文的检查类型

const (
	Unknown Item = iota // Unknown 未知检查类型
	Caller              // Caller zero.APICaller
	Event               // Event *zero.Event
)

const Timeout = 5 * time.Minute // Timeout 超时时间 5 分钟
