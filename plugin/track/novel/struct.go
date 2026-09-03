package novel

import (
	"context"
	"fmt"
	"time"

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
		UpdateData              // UpdateData 更新数据
		*chapter.Chapter        // Chapter 新章节信息
	}

	// WriterInfo 作者信息
	WriterInfo struct {
		Writer  string // Writer 作者昵称
		HeadURL string // HeadURL 头像链接
	}

	// Data 小说数据
	Data struct {
		Right        []string // Right 版权状态
		Protagonists []string // Protagonists 主角
		Collection   uint64   // Collection 小说收藏
		HitNum       uint64   // HitNum 小说点击
	}

	// Info 小说信息
	Info struct {
		CoverURL  string   // CoverURL 封面链接
		Theme     string   // Theme 小说类型（主题）
		Introduce string   // Introduce 小说简述
		Item      []string // Item 小说参加的项目
		TagList   []string // TagList 标签列表
	}

	// UpdateData 更新数据
	UpdateData struct {
		Status                   // Status 小说状态
		TotalWordNum      uint64 // TotalWordNum 小说字数
		TodayWordNum      uint64 // TodayWordNum 当日更新字数
		WeekDailyWordNum  uint64 // WeekDailyWordNum 七日平均字数
		MonthDailyWordNum uint64 // MonthDailyWordNum 三十日平均字数
		DailyWordNum      uint64 // DailyWordNum 全书每日平均字数
		time.Duration            // Duration 距离上次更新的时间差
		Times             uint   // Times 当日更新次数
	}

	// Status 小说状态
	Status string
)

type (
	// Commenter 小说点评者
	Commenter interface {
		// CommentNovel 评论小说信息
		CommentNovel(ctx context.Context, nv *Novel) string
		// CommentUpdate 评论小说更新
		CommentUpdate(ctx context.Context, nv *Novel) (string, error)
	}

	// BookCommentBox 小说点评框
	BookCommentBox func(nv fmt.Stringer) string
)

const (
	// Completed 已完结
	Completed Status = `已完结`
	// Ongoing 连载中
	Ongoing Status = `连载中`
	// Inactive 断更
	Inactive Status = `断更`
)
