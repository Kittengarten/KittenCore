// Package seg 用于表示消息段类型
package seg

const (
	// Text 纯文本
	Text = `text`
	// Face QQ 表情
	Face = `face`
	// Image 图片
	Image = `image`
	// Record 语音
	Record = `record`
	// Video 短视频
	Video = `video`
	// At @某人
	At = `at`
	// RPS 猜拳魔法表情
	RPS = `rps`
	// Dice 掷骰子魔法表情
	Dice = `dice`
	// Shake 窗口抖动（戳一戳）
	Shake = `shake`
	// Poke 戳一戳
	Poke = `poke`
	// Anonymous 匿名发消息
	Anonymous = `anonymous`
	// Share 链接分享
	Share = `share`
	// Contact 推荐好友或群
	Contact = `contact`
	// Location 位置
	Location = `location`
	// Music 音乐分享
	Music = `music`
	// Reply 回复
	Reply = `reply`
	// Forward 合并转发
	Forward = `forward`
	// Node 合并转发节点
	Node = `node`
	// XML 消息
	XML = `xml`
	// JSON 消息
	JSON = `json`
)
