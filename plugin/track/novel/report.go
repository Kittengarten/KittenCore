package novel

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/plugin/track/chapter"
	"github.com/Kittengarten/KittenCore/plugin/track/status"

	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	// NewChapter 获取章节
	NewChapter func(nv *Novel, cpURL string) (*chapter.Chapter, error)
	// CommentUpdate (nv *Novel) (string, error) 评论更新
	CommentUpdate = func(_ *Novel) (string, error) {
		// 默认为空实现
		return ``, nil
	}
)

// ErrNoComment 没有评论内容喵！
var ErrNoComment = errors.New(`没有评论内容喵！`)

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
) {
	const tryCount = 5 // 重试最多 5 次
	t := time.NewTicker(shttp.TimeOutSeconds * time.Second)
	defer t.Stop()
	for range tryCount {
		s, err := CommentUpdate(nv)
		if err == nil {
			for i, user := range users {
				msgr.Reply(msgID[i]).Text(s).Send(user)
			}
			return
		}
		switch as := struct {
			ErrURL   *url.Error
			ErrShttp *shttp.Error
		}{}; {
		case errors.Is(err, ErrNoComment),
			errors.As(err, &as.ErrURL),
			errors.As(err, &as.ErrShttp):
			// 在以下错误时重试：
			// 没有评论内容喵！
			// *url.Error
			// *shttp.Error
			kitten.Error(err)
			<-t.C
			continue
		}
		break
	}
	kitten.Error(`评论更新失败喵！`)
}

var (
	// RestoreURL 恢复 URL
	RestoreURL func(nv *Novel) string
	// RestorePlatform 恢复平台
	RestorePlatform func(nv *Novel)
)

// Update 更新信息
func (nv *Novel) Update() string {
	defer RestorePlatform(nv)
	return fmt.Sprintf(`《%s》更新了喵～
%s%s
更新字数：%d 字（%s）%s`,
		nv.Name,
		nv.Title,
		RestoreURL(nv),
		nv.WordNum, func(v bool) string {
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

// DurationConvert 距上次更新时间的时间差转换为时间间隔的结构体
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
