package sfacg

import "time"

type (
	SFAPI uint8
	Novel struct {
		Id         string   // 小说书号
		Url        string   // 小说网址
		Name       string   // 小说书名
		IsVip      bool     // 是否上架
		Writer     string   // 作者昵称
		HitNum     string   // 小说点击
		WordNum    string   // 小说字数
		Preview    string   // 章节预览
		HeadUrl    string   // 头像网址
		CoverUrl   string   // 封面网址
		Collection string   // 小说收藏
		NewChapter Chapter  // 章节信息
		Type       string   // 小说类型
		Introduce  string   // 小说简述
		Status     string   // 小说状态
		TagList    []string // 标签列表
		IsGet      bool     // 是否可以获取
	}

	Chapter struct {
		BookUrl string    // 本书网址
		Url     string    // 章节网址
		Time    time.Time // 更新时间
		Title   string    // 章节名称
		WordNum int       // 章节字数
		LastUrl string    // 上章网址
		NextUrl string    // 下章网址
		IsGet   bool      // 是否可以获取
	}

	Compare struct {
		Times   int           // 更新次数
		TimeGap time.Duration // 更新时间差
	}

	Config []struct {
		BookId     string  // 报更书号
		GroupID    []int64 // 书友群号
		RecordUrl  string  // 更新记录
		Updatetime string  // 更新时间

	}
)
