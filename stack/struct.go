package stack

import "time"

type (
	Config struct {
		MaxStack   int    `yaml:"maxstack"`   // 叠猫猫队列上限
		MaxTime    int    `yaml:"maxtime"`    // 叠猫猫时间上限（小时数）
		GapTime    int    `yaml:"gaptime"`    // 叠猫猫退出后重新加入所需的时间
		OutOfStack string `yaml:"outofstack"` // 叠猫猫队列已满的回复
	}

	Data []Kitten

	Kitten struct {
		Id   int64     `yaml:"id"`   // QQ
		Name string    `yaml:"name"` // 群名片或昵称
		Time time.Time `yaml:"time"` // 进入队列的时间
	}
)
