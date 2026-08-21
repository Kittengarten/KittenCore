package chapter

import "time"

// Chapter 一个章节
type Chapter struct {
	Update  time.Time // 更新时间
	URL     string    // URL 章节链接
	Title   string    // Title 章节名称
	PreURL  string    // PreURL 上章链接
	NextURL string    // NextURL 下章链接
	WordNum int       // WordNum 章节字数
	IsVIP   bool      // IsVIP 是否付费章节
	NeedPay bool      // NeedPay 是否需要付费
}
