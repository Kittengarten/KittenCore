package novel

import (
	"fmt"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/plugin/track/chapter"
	"github.com/Kittengarten/KittenCore/plugin/track/status"

	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	// NewChapter 获取章节
	NewChapter func(nv *Novel, cpURL string) (*chapter.Chapter, error)
	// CommentUpdate 评论更新
	CommentUpdate = func(nv *Novel, cpID string) string {
		// 默认为空实现
		return ``
	}
)

// 与上次更新比较
func (nv *Novel) makeCompare() (err error) {
	var this, pre *chapter.Chapter
	this = nv.Chapter
	if this.PreURL == `` || this.PreURL == nv.URL {
		return status.ErrStatus(nv.URL, status.OnlyAChapter)
	}
	if pre, err = NewChapter(nv, this.PreURL); err != nil {
		return err
	}
	nv.TodayWordNum = this.WordNum
	nv.Duration = max(time.Second, this.Sub(pre.Time))
	for nv.Times = 1; equal.IsSameDate(pre.Time, this.Time) &&
		pre.PreURL != nv.URL; nv.Times++ {
		times.RandomDelayRange(time.Second, 2*time.Second)
		this = pre
		nv.TodayWordNum += this.WordNum
		if pre, err = NewChapter(nv, this.PreURL); err != nil {
			return err
		}
	}
	return nil
}

// TryCommentUpdate 尝试评论更新
func TryCommentUpdate(
	msgr *kitten.Messager,
	msgID []message.ID,
	users []kitten.QQ,
	nv *Novel,
	cpID string,
) {
	const tryCount = 5 // 重试最多 5 次
	for range tryCount {
		s := CommentUpdate(nv, cpID)
		if s == `` {
			continue
		}
		for i, user := range users {
			msgr.Reply(msgID[i]).Text(s).Send(user)
		}
		return
	}
	kitten.Error(`评论更新失败喵！`)
}

var (
	// RestoreURL 恢复 URL
	RestoreURL func(nv *Novel) string
	// RestorePlatform 恢复平台
	RestorePlatform func(nv *Novel)
)

// 更新信息
func (nv *Novel) Update() string {
	defer RestorePlatform(nv)
	return fmt.Sprintf(`《%s》更新了喵～
%s%s
更新字数：%d 字（%s）%s`,
		nv.Name,
		nv.Title,
		RestoreURL(nv),
		nv.Chapter.WordNum, func(v bool) string {
			if v {
				return `付费`
			}
			return `免费`
		}(nv.IsVIP),
		func() string {
			tr, err := nv.todayReport()
			if err != nil {
				kitten.Error(err)
				return ``
			}
			return "\n间隔时间：" + tr
		}(),
	)
}

// 今日报更
func (nv *Novel) todayReport() (string, error) {
	if err := nv.makeCompare(); err != nil {
		return ``, err
	}
	s, err := nv.DurationConvert()
	if err != nil {
		return ``, err
	}
	return s.String() + "\n" + nv.todayUpdate(), nil
}

// 距上次更新时间的时间差转换为时间间隔的结构体
func (nv *Novel) DurationConvert() (times.TimeDuration, error) {
	// 如果时间早于 2006.1.2 15:04:05
	if s, _ := time.Parse(times.Layout, times.Layout); nv.Duration > time.Since(s) {
		return times.TimeDuration{}, status.ErrStatus(nv.URL, status.TimeException)
	}
	return times.ConvertTimeDuration(nv.Duration), nil
}

// 今日更新信息
func (nv *Novel) todayUpdate() string {
	switch nv.Times {
	case 0:
		return ``
	case 1:
		return `当日第 1 更`
	default:
		return fmt.Sprint(`当日第 `, nv.Times, ` 更，日更 `, nv.TodayWordNum, ` 字`)
	}
}
