package eekda

import "time"

type (
	// Today 今天吃什么
	Today struct {
		Time      time.Time `yaml:"time"`      // 生成时间
		Breakfast int64     `yaml:"breakfast"` // 早餐
		Lunch     int64     `yaml:"lunch"`     // 午餐
		LowTea    int64     `yaml:"lowtea"`    // 下午茶
		Dinner    int64     `yaml:"dinner"`    // 晚餐
		Supper    int64     `yaml:"supper"`    // 夜宵
	}

	// Stat 饮食统计数据
	Stat []Kitten

	// Kitten 猫猫数据
	Kitten struct {
		ID        int64  `yaml:"id"`        // QQ
		Name      string `yaml:"name"`      // 群名片或昵称
		Breakfast int64  `yaml:"breakfast"` // 早餐次数
		Lunch     int64  `yaml:"lunch"`     // 午餐次数
		LowTea    int64  `yaml:"lowtea"`    // 下午茶次数
		Dinner    int64  `yaml:"dinner"`    // 晚餐次数
		Supper    int64  `yaml:"supper"`    // 夜宵次数
	}
)
