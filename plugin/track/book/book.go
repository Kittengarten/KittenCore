package book

import (
	"net/url"
	"slices"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/stat"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/core/times/repeat"
	"github.com/Kittengarten/KittenCore/kitten/core/times/stale"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/plugin/track/check"
	"github.com/Kittengarten/KittenCore/plugin/track/novel"
	"github.com/Kittengarten/KittenCore/plugin/track/platform"
	"github.com/Kittengarten/KittenCore/plugin/track/platform/fanqie"

	"github.com/tidwall/gjson"
)

const cycle = shttp.Timeout // 最小循环间隔 10 秒（最大可翻倍）

// Report 执行报更
func (c *Books) Report(
	handler *msg.Handler,
	cu chan Books,
	path fio.PathRWMutex,
) {
	if err:= repeat.IterS(
		handler,
		repeat.New(0, cycle, 2*cycle),
		*c,
		func(i int, b Book) error {
			b.Report(handler, cu, c, i, path)
			return nil
		},
	); err != nil {
		log.Error(err)
	}
}

// Report 执行单本书的报更
func (b Book) Report(
	handler *msg.Handler,
	cu chan Books,
	c *Books, i int,
	path fio.PathRWMutex,
) {
	switch b.Platform {
	case fanqie.Platform.String():
		u, err := url.Parse(stat.APIHOST[stat.Fanqie])
		if err != nil {
			// API 无法访问，使用网页模式
			// 接收到专用的慢速定时器信号才释放
			log.Error(`番茄 API 解析错误：`, err)
			<-times.RandDelayRange(60*cycle, 120*cycle)
			break
		}
		res, err := shttp.GETDataURLWithContext(handler, u.JoinPath(`self-info`))
		if err != nil {
			// API 无法访问，使用网页模式
			// 接收到专用的慢速定时器信号才释放
			log.Error(`番茄 API 不可用：`, err)
			<-times.RandDelayRange(60*cycle, 120*cycle)
			break
		}
		if gjson.GetBytes(res, `message`).String() == `SUCCESS` {
			// API 可以访问，切换为 API 模式
			b.Platform = fanqie.API.String()
		}
	}
	p, err := platform.Get(b.Platform)
	if err != nil {
		log.Error(`平台错误：`, err)
		return
	}
	nv, err := novel.Init(handler, p, b.BookID)
	defer novel.Pool.Put(nv)
	if err != nil {
		log.Error(`初始化小说错误：`, err)
		return
	}
	check.RecordTime()
	if nv.Chapter.URL == `` {
		// 如果没有获取到 URL，则跳过
		return
	}
	switch b.Platform {
	case fanqie.Platform.String(), fanqie.API.String():
		if !fanqie.ShouldUpdate(nv.Chapter.URL, b.RecordURL) {
			// 如果番茄（API 和 网页不同途径之间的比较）不应更新，则跳过
			return
		}
	default:
		if nv.Chapter.URL == b.RecordURL {
			// 如果没有更新，则跳过
			return
		}
	}
	if len(nv.Protagonists) == 0 {
		nv.Protagonists = b.Protagonists
	}
	// 异步发送更新消息
	novel.TryCommentUpdate(
		handler,
		handler.Image(
			fio.NewPath(nv.CoverURL),
			fio.NewPath(nv.HeadURL),
		).Text(nv.Update(handler)).SendMulti(b.Users...),
		b.Users,
		nv,
	)
	// 写入小说更新数据
	(*c)[i].BookName = nv.Name
	(*c)[i].Writer = nv.Writer
	(*c)[i].RecordURL = nv.Chapter.URL
	(*c)[i].UpdateTime = nv.Chapter.Update
	// 异步保存配置
	save := make(chan error, 1)
	utils.Go(`保存报更配置`, func() {
		path.Lock()
		defer path.Unlock()
		save <- c.SaveConfig(cu, path.Path)
		close(save)
	})
	if err := <-save; err != nil {
		log.Error(ErrSave, err)
		return
	}
	log.Info(`更新《`, nv.Name, `》成功喵！`)
}

// SaveConfig 保存报更
func (c *Books) SaveConfig(cu chan<- Books, path fio.Path) error {
	// 按更新时间倒序排列
	c.sortByUpdate()
	if err := fio.Save(path, *c); err != nil {
		return err
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
	e := stale.Check(b.UpdateTime).String()
	return e + `《` + b.BookName + `》` + `
作者：　　	` + b.Writer + `
平台：　　	` + b.Platform + `
书号：　　	` + b.BookID + `
上次更新：	` + func() string {
		if b.UpdateTime.IsZero() {
			return `未知`
		}
		return b.UpdateTime.Format(times.LayoutHeart)
	}()
}
