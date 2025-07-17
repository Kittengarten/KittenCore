package book

import (
	"net/url"
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/plugin/track/novel"
	"github.com/Kittengarten/KittenCore/plugin/track/platform"
	"github.com/Kittengarten/KittenCore/plugin/track/platform/fanqie"
	"github.com/tidwall/gjson"
)

// Report 执行报更
func (c *Books) Report(
	msgr *kitten.Messager,
	cu chan Books,
	path fio.PathRWMutex,
	cycle time.Duration,
	t, st *time.Ticker,
) {
	for i, b := range *c {
		<-t.C
		switch b.Platform {
		case fanqie.Platform.String():
			u, err := url.Parse(fanqie.APIHOST)
			if err != nil {
				// API 无法访问，使用网页模式
				// 接收到专用的慢速定时器信号才释放
				kitten.Error(`番茄 API 解析错误：`, err)
				<-st.C
				break
			}
			res, err := shttp.GETDataURL(u.JoinPath(`self-info`))
			if err != nil {
				// API 无法访问，使用网页模式
				// 接收到专用的慢速定时器信号才释放
				kitten.Error(`番茄 API 不可用：`, err)
				<-st.C
				break
			}
			if gjson.GetBytes(res, `message`).String() == `SUCCESS` {
				// API 可以访问，切换为 API 模式
				b.Platform = fanqie.API.String()
			}
		}
		p, err := platform.Get(b.Platform)
		if err != nil {
			kitten.Error(`平台错误：`, err)
			continue
		}
		nv, err := novel.Init(p, b.BookID)
		if err != nil {
			kitten.Error(`初始化小说错误：`, err)
			continue
		}
		switch b.Platform {
		case fanqie.Platform.String(), fanqie.API.String():
			if fanqie.IsUpdate(nv.Chapter.URL, b.RecordURL) {
				// 如果番茄（API 和 网页不同途径之间的比较）没有更新，则跳过
				novel.Pool.Put(nv)
				continue
			}
		default:
			if nv.Chapter.URL == b.RecordURL {
				// 如果没有更新，则跳过
				novel.Pool.Put(nv)
				continue
			}
		}
		if len(nv.Protagonists) == 0 {
			nv.Protagonists = b.Protagonists
		}
		done := make(chan struct{}, 1)
		// 发送更新消息
		go novel.TryCommentUpdate(
			msgr,
			msgr.Image(
				fio.NewPath(nv.CoverURL),
				fio.NewPath(nv.HeadURL),
			).Text(nv.Update()).SendMulti(b.Users...),
			b.Users,
			nv,
			done,
		)
		// 写入小说更新数据
		(*c)[i].BookName = nv.Name
		(*c)[i].Writer = nv.Writer
		(*c)[i].RecordURL = nv.Chapter.URL
		(*c)[i].UpdateTime = nv.Chapter.Update
		// 按更新时间倒序排列
		c.sortByUpdate()
		// 异步保存配置
		save := make(chan error, 1)
		go func() {
			path.Lock()
			defer path.Unlock()
			save <- c.SaveConfig(cu, path.Path)
		}()
		if err := <-save; err != nil {
			kitten.Error(ErrSave, err)
			continue
		}
		close(save)
		kitten.Info(`更新《`, nv.Name, `》成功喵！`)
		<-done
		novel.Pool.Put(nv)
	}
}

// SaveConfig 保存报更
func (c *Books) SaveConfig(cu chan Books, path fio.Path) error {
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
	return `《` + b.BookName + `》` + `
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
