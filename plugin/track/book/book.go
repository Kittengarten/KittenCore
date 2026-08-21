package book

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/core/times/repeat"
	"github.com/Kittengarten/KittenCore/kitten/core/times/stale"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/plugin/track/check"
	"github.com/Kittengarten/KittenCore/plugin/track/novel"

	"github.com/Kittengarten/KittenCore/plugin/track/platform"
)

const cycle = shttp.Timeout // 最小循环间隔 10 秒（最大可翻倍）

// Report 执行报更
func (c *Books) Report(
	handler *msg.Handler,
	cu chan Books,
	path fio.PathRWMutex,
) {
	select {
	case <-handler.Done():
		log.Error(handler.Err())
		return
	default:
	}
	if err := repeat.IterS(
		handler,
		repeat.New(0, cycle, 2*cycle),
		*c,
		func(i int, b Book) error {
			b.Report(handler, c, i)
			return nil
		},
	); err != nil {
		log.Error(err)
	}
	// 异步保存配置
	save := make(chan error, 1)
	utils.Go(`保存报更配置`, func() {
		path.Lock()
		defer path.Unlock()
		save <- c.SaveConfig(handler, cu, path.Path, true)
		close(save)
	})
	if err := <-save; err != nil {
		log.Error(ErrSave, err)
		return
	}
}

// Report 执行单本书的报更
func (b Book) Report(
	handler *msg.Handler,
	c *Books, i int,
) {
	p, err := platform.Get(b.Platform)
	if err != nil {
		log.Error(`平台 `, b.Platform, ` 错误：`, err)
		return
	}
	nv, err := p.Init(handler, b.BookID)
	if err != nil {
		if !errors.Is(err, shttp.ErrNoUpdate) {
			log.Error(`初始化 `, p, ` 小说书号 `, b.BookID, ` 错误：`, err)
		}
		// 无更新，跳过
		nv.Put()
		return
	}
	check.RecordTime() // 只记录实际检测的 *novel.Novel 数量
	// 记录当日更新字数
	(*c)[i].TodayWordNum = nv.TodayWordNum
	if !equal.IsSameDate4AM(time.Now(), b.UpdateTime) {
		(*c)[i].TodayWordNum = 0
	}
	if nv.Chapter.URL == `` ||
		!nv.Chapter.Update.After(b.UpdateTime) {
		// 没有获取到 URL
		// 更新时间不在上次更新时间之后
		// 跳过
		nv.Put()
		return
	}
	if nv.Chapter.URL == b.RecordURL {
		// 没有更新，跳过
		nv.Put()
		return
	}
	if len(nv.Protagonists) == 0 {
		nv.Protagonists = b.Protagonists
	}
	done := make(chan struct{})
	// 异步发送更新消息
	var paths []fio.Path
	if nv.CoverURL != `` {
		paths = append(paths, fio.NewPath(nv.CoverURL))
	}
	if nv.HeadURL != `` {
		paths = append(paths, fio.NewPath(nv.HeadURL))
	}
	novel.TryCommentUpdate(
		handler,
		handler.Image(paths...).Text(nv.Update(handler)).SendMulti(b.Users...),
		b.Users,
		nv,
		done,
	)
	// 写入小说更新数据
	(*c)[i].BookName = nv.Name
	(*c)[i].Writer = nv.Writer
	(*c)[i].RecordURL = nv.Chapter.URL
	(*c)[i].UpdateTime = nv.Chapter.Update
	(*c)[i].Completed = nv.Status == novel.Completed
	log.Info(`更新《`, nv.Name, `》成功喵！`)
	<-done
	nv.Put()
}

// SaveConfig 保存报更
func (c *Books) SaveConfig(ctx context.Context, cu chan Books, p fio.Path, auto bool) error {
	// 按更新时间倒序排列
	c.sortByUpdate()
	if err := p.Save(*c); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case cu <- *c:
	default:
	}
	if auto {
		// 如果是自动保存，不尝试丢弃旧值
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-cu: // 丢弃旧值
	default:
	}
	cu <- *c
	return nil
}

// 按更新时间倒序排列小说
func (c *Books) sortByUpdate() {
	slices.SortFunc(*c, func(j, i Book) int {
		return i.UpdateTime.Compare(j.UpdateTime)
	})
}

// String 实现 fmt.Stringer
func (b Book) String() string {
	e := stale.Check(b.UpdateTime, b.Completed).String()
	return e + `《` + b.BookName + `》` + `
作者：　　	` + b.Writer + `
平台：　　	` + b.Platform + `
书号：　　	` + b.BookID + `
上次更新：	` + func() string {
		if b.UpdateTime.IsZero() {
			return `未知`
		}
		return b.UpdateTime.Format(times.LayoutHeart)
	}() + `
今日字数：	` + strconv.Itoa(b.TodayWordNum) + rank(b.TodayWordNum, b.Completed)
}

func rank(todayWordNum int, completed bool) string {
	if completed {
		return `（已完结）`
	}
	tier := map[int]string{
		10000: `（夯）`,
		6000:  `（顶级）`,
		4000:  `（人上人）`,
		2000:  `（NPC）`,
		0:     `（拉完了）`,
	}
	for k, v := range tier {
		if todayWordNum >= k {
			return v
		}
	}
	return ``
}
