package sfacg

import "time"

type (
	// Novel 表示一本小说的数据集
	Novel struct {
		ID         string   // 小说书号
		URL        string   // 小说网址
		Name       string   // 小说书名
		IsVip      bool     // 是否上架
		Writer     string   // 作者昵称
		HitNum     string   // 小说点击
		WordNum    string   // 小说字数
		Preview    string   // 章节预览
		HeadURL    string   // 头像网址
		CoverURL   string   // 封面网址
		Collection string   // 小说收藏
		NewChapter Chapter  // 章节信息
		Type       string   // 小说类型
		Introduce  string   // 小说简述
		Status     string   // 小说状态
		TagList    []string // 标签列表
		IsGet      bool     // 是否可以获取
	}

	// Chapter 表示一个章节的数据集
	Chapter struct {
		BookURL string    // 本书网址
		URL     string    // 章节网址
		Time    time.Time // 更新时间
		Title   string    // 章节名称
		WordNum int       // 章节字数
		LastURL string    // 上章网址
		NextURL string    // 下章网址
		IsGet   bool      // 是否可以获取
	}

	// Compare 表示章节之间比较的数据集
	Compare struct {
		Times   int           // 更新次数
		TimeGap time.Duration // 更新时间差
	}

	// Config 表示多项小说配置的数据集组成的数组
	Config []struct {
		BookID     string  `yaml:"bookid"`     // 报更书号
		BookName   string  `yaml:"bookname"`   // 报更书名
		GroupID    []int64 `yaml:"groupid"`    // 书友群号
		RecordURL  string  `yaml:"recordurl"`  // 上次更新链接
		UpdateTime string  `yaml:"updatetime"` // 上次更新时间
	}

	keyWord string // 搜索关键词
)
