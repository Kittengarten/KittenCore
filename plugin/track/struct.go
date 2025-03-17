package track

import (
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
)

type (
	// 一本小说
	novel struct {
		Platform           // 小说平台
		id         string  // 小说书号
		name       string  // 小说书名
		url        string  // 小说网址
		writerInfo         // 作者信息
		data               // 小说数据
		info               // 小说信息
		newChapter chapter // 章节信息
		compare            // 章节之间比较（报更时才初始化）
	}

	// 作者信息
	writerInfo struct {
		writer  string // 作者昵称
		headURL string // 头像网址
	}

	// 小说数据
	data struct {
		right      []string // 版权状态
		collection string   // 小说收藏
		hitNum     string   // 小说点击
		wordNum    string   // 小说字数
	}

	// 小说信息
	info struct {
		coverURL  string   // 封面网址
		preview   string   // 章节预览（仅 SFACG）
		theme     string   // 小说类型（主题）
		introduce string   // 小说简述
		status    string   // 小说状态
		item      []string // 小说参加的项目
		tagList   []string // 标签列表
	}

	// 一个章节
	chapter struct {
		bookURL   string // 本书网址
		url       string // 章节网址
		isVIP     bool   // 是否付费章节
		time.Time        // 更新时间
		title     string // 章节名称
		wordNum   int    // 章节字数
		preURL    string // 上章网址
		nextURL   string // 下章网址
	}

	// 章节之间比较的数据集
	compare struct {
		times         int // 当日更新次数
		todayWordNum  int // 当日更新字数
		time.Duration     // 距离上次更新的时间差
	}

	// 多项小说报更项目的数据集组成的切片
	books []book

	// 小说报更项目的数据集
	book struct {
		Platform               // 报更平台
		BookID     string      // 报更书号（为了未来兼容性，不使用数值）
		BookName   string      // 报更书名
		Writer     string      // 小说作者
		Users      []kitten.QQ // 用户，正数代表 QQ 号，负数代表群号
		RecordURL  string      // 上次更新链接
		UpdateTime time.Time   // 上次更新时间
	}

	Platform string // 平台

	keyword string // 搜索关键词
)
