package novel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/core/times/repeat"
	"github.com/Kittengarten/KittenCore/kitten/core/times/retry"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/usr"
	"github.com/Kittengarten/KittenCore/plugin/track/chapter"
	"github.com/Kittengarten/KittenCore/plugin/track/platform"
	"github.com/Kittengarten/KittenCore/plugin/track/status"

	"github.com/wdvxdr1123/ZeroBot/message"
)

type (
	// Restorer 恢复器
	Restorer interface {
		URLRestorer
		PlatformRestorer
	}
	// URLRestorer URL 恢复器
	URLRestorer interface {
		RestoreURL(nv *Novel) string
	}
	// PlatformRestorer 平台恢复器
	PlatformRestorer interface {
		RestorePlatform(nv *Novel)
	}
)

// GlobalRestorer 全局恢复器
var GlobalRestorer Restorer

var (
	// ErrorNotImplemented 未实现喵！
	ErrorNotImplemented = errors.New(`未实现喵！`)
	// ErrNoComment 没有评论内容喵！
	ErrNoComment = retry.NewError(errors.New(`没有评论内容喵！`), true)
)

// TryCommentUpdate 尝试评论更新
func TryCommentUpdate(
	handler *msg.Handler,
	ids []message.ID,
	users []usr.QQ,
	nv *Novel,
) {
	if Export.Commenter == nil {
		log.Error(`更新点评`, ErrorNotImplemented)
		return
	}
	utils.Go(`异步评论更新`, func() {
		var (
			// 独立上下文，不继承上游
			ctx, cancel = context.WithTimeout(handler, 5*time.Minute)
			s           string
		)
		defer cancel()
		retry.Do(
			ctx,
			retry.Default(),
			func() (err error) {
				s, err = Export.CommentUpdate(handler.SetContext(ctx), nv)
				return err
			},
		)
		if s == `` {
			return
		}
		if err := repeat.IterS(
			handler,
			repeat.New(0, time.Second, 2*time.Second),
			users,
			func(i int, u usr.QQ) error {
				_ = handler.Quote(ids[i]).Text(s).Send(u)
				return nil
			},
		); err != nil {
			log.Error(err)
		}
	})
}

// Update 更新信息
func (nv *Novel) Update(ctx context.Context) string {
	defer GlobalRestorer.RestorePlatform(nv)
	return fmt.Sprintf(`《%s》更新了喵～
%s%s
更新字数：%d 字（%s）%s`,
		nv.Name,
		nv.Title,
		GlobalRestorer.RestoreURL(nv),
		nv.WordNum, map[bool]string{
			true:  `付费`,
			false: `免费`,
		}[nv.IsVIP],
		func() string {
			tr, err := nv.todayReport(ctx)
			if err != nil {
				log.Error(err)
				return ``
			}
			return "\n间隔时间：" + tr
		}(),
	)
}

// 今日报更
func (nv *Novel) todayReport(ctx context.Context) (string, error) {
	if err := nv.makeCompare(ctx); err != nil {
		return ``, err
	}
	if nv.Times == 0 {
		return ``, status.ErrStatus(nv.URL, status.TimeException)
	}
	s, err := nv.DurationConvert()
	if err != nil {
		return ``, err
	}
	return s.String() + "\n" + nv.todayUpdate(), nil
}

// 与上次更新比较
func (nv *Novel) makeCompare(ctx context.Context) (err error) {
	var this, pre *chapter.Chapter
	this = nv.Chapter
	if this.PreURL == `` || this.PreURL == nv.URL {
		return status.ErrStatus(nv.URL, status.OnlyAChapter)
	}
	if pre, err = NewChapter(ctx, nv, this.PreURL); err != nil {
		return err
	}
	nv.TodayWordNum = this.WordNum
	nv.Duration = max(time.Second, this.Update.Sub(pre.Update))
	for nv.Times = 1; equal.IsSameDate4AM(pre.Update, this.Update) &&
		pre.PreURL != nv.URL; nv.Times++ {
		select {
		case <-times.RandDelayRange(time.Second, 2*time.Second):
			this = pre
			nv.TodayWordNum += this.WordNum
			if pre, err = NewChapter(ctx, nv, this.PreURL); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// NewChapter 获取章节
func NewChapter(ctx context.Context, nv *Novel, cpURL string) (*chapter.Chapter, error) {
	p, err := platform.Get(nv.Platform)
	if err != nil {
		return nil, err
	}
	return chapter.New(ctx, p, cpURL)
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
