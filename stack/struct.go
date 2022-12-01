package stack

import "time"

type (
	// Config 叠猫猫配置
	Config struct {
		MaxStack    int    `yaml:"maxstack"`    // 叠猫猫队列上限
		MaxTime     int    `yaml:"maxtime"`     // 叠猫猫时间上限（小时数）
		GapTime     int    `yaml:"gaptime"`     // 叠猫猫主动退出或压猫猫、被压坏、摔下来后重新加入所需的时间（小时数）
		OutOfStack  string `yaml:"outofstack"`  // 叠猫猫队列已满的回复
		MaxCount    int    `yaml:"maxcount"`    // 被压次数上限
		FailPercent int    `yaml:"failpercent"` // 叠猫猫每层失败概率百分数
	}

	// Data 叠猫猫数据
	Data []Kitten

	// Kitten 猫猫数据
	Kitten struct {
		ID    int64     `yaml:"id"`    // QQ
		Name  string    `yaml:"name"`  // 群名片或昵称
		Time  time.Time `yaml:"time"`  // 进入队列的时间
		Count int       `yaml:"count"` // 被压次数
	}

	// DataPath 叠猫猫数据 + 存储路径
	DataPath struct {
		Data
		Path string
	}
)
