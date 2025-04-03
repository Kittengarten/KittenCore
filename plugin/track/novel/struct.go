package novel

import (
	"github.com/Kittengarten/KittenCore/plugin/track/chapter"
)

type (
	// Novel 一本小说
	Novel struct {
		Platform         string // Platform 小说平台
		ID               string // ID 小说书号
		Name             string // Name 小说书名
		URL              string // URL 小说链接
		WriterInfo              // WriterInfo 作者信息
		Data                    // Data 小说数据
		Info                    // Info 小说信息
		*chapter.Chapter        // 新章节信息
		chapter.Compare         // 章节之间比较（报更时才初始化）
	}

	// WriterInfo 作者信息
	WriterInfo struct {
		Writer  string // Writer 作者昵称
		HeadURL string // HeadURL 头像链接
	}

	// Data 小说数据
	Data struct {
		Protagonists []string // Protagonists 主角
		Right        []string // Right 版权状态
		Collection   string   // Collection 小说收藏
		HitNum       string   // HitNum 小说点击
		TotalWordNum string   // TotalWordNum 小说字数
	}

	// Info 小说信息
	Info struct {
		CoverURL  string   // CoverURL 封面链接
		Preview   string   // Preview 章节预览（仅 SFACG）
		Theme     string   // Theme 小说类型（主题）
		Introduce string   // Introduce 小说简述
		Status    string   // Status 小说状态
		Item      []string // Item 小说参加的项目
		TagList   []string // TagList 标签列表
	}
)
