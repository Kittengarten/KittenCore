package chapter

import "time"

type (
	// Chapter 一个章节
	Chapter struct {
		Update  time.Time // 更新时间
		URL     string    // URL 章节链接
		Title   string    // Title 章节名称
		PreURL  string    // PreURL 上章链接
		NextURL string    // NextURL 下章链接
		WordNum int       // WordNum 章节字数
		IsVIP   bool      // IsVIP 是否付费章节
	}

	// Compare 章节之间比较的数据集
	Compare struct {
		Times         int // 当日更新次数
		TodayWordNum  int // 当日更新字数
		time.Duration     // 距离上次更新的时间差
	}
)
