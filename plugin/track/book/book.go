package book

import (
	"net/http"
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/io"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/plugin/track/novel"
	"github.com/Kittengarten/KittenCore/plugin/track/platform"
	"github.com/Kittengarten/KittenCore/plugin/track/platform/fanqie"
)

const (
	Without   = `这里没有添加小说报更喵～`
	ErrConfig = `报更配置文件错误喵！`
	ErrLoad   = `加载` + ErrConfig
	ErrSave   = `保存` + ErrConfig
	Unknown   = `未知`
)

type (
	// 多项小说报更项目的数据集组成的切片
	Books []Book
	// 小说报更项目的数据集
	Book struct {
		Platform   string      // 报更平台
		BookID     string      // 报更书号（为了未来兼容性，不使用数值）
		BookName   string      // 报更书名
		Writer     string      // 小说作者
		Users      []kitten.QQ // 用户，正数代表 QQ 号，负数代表群号
		RecordURL  string      // 上次更新链接
		UpdateTime time.Time   // 上次更新时间
	}
)

// 按更新时间倒序排列小说
func (c *Books) sortByUpdate() {
	slices.SortFunc(*c, func(j, i Book) int {
		return i.UpdateTime.Compare(j.UpdateTime)
	})
}

// SaveConfig 保存报更
func (c *Books) SaveConfig(cu chan Books, path io.Path) error {
	c.sortByUpdate()
	if err := io.Save(path, *c); err != nil {
		return err
	}
	cu <- *c
	return nil
}

// Report 执行报更
func (c *Books) Report(
	msgr *kitten.Messager,
	cu chan Books,
	path io.Path,
	cycle time.Duration,
	st *time.Ticker,
) {
	for i, b := range *c {
		times.RandomDelayRange(cycle, 2*cycle)
		switch b.Platform {
		case fanqie.Platform.String():
			res, err := http.Get(fanqie.APIHOST)
			if err != nil {
				// API 无法访问，使用网页模式
				// 接收到专用的慢速定时器信号才释放
				<-st.C
				break
			}
			if res != nil && res.StatusCode == http.StatusOK {
				res.Body.Close()
				// API 可以访问，切换为 API 模式
				b.Platform = fanqie.API.String()
			}
		}
		nv, err := platform.Get(b.Platform).Init(b.BookID)
		if err != nil {
			kitten.Error(err)
			continue
		}
		switch b.Platform {
		case fanqie.Platform.String(), fanqie.API.String():
			if fanqie.IsUpdate(nv.Chapter.URL, b.RecordURL) {
				// 如果番茄（API 和 网页不同途径之间的比较）没有更新，则跳过
				novel.Pool.Put(nv)
				continue
			}
		}
		if nv.Chapter.URL == b.RecordURL {
			// 如果没有更新，则跳过
			novel.Pool.Put(nv)
			continue
		}
		// 发送更新消息
		go novel.TryCommentUpdate(
			msgr,
			msgr.Image(
				io.NewPath(nv.CoverURL),
				io.NewPath(nv.HeadURL),
			).Text(nv.Update()).SendMulti(b.Users...),
			b.Users,
			nv,
			platform.Get(nv.Platform).ChapterID(nv.Chapter.URL))
		// 写入小说更新数据
		(*c)[i].BookName = nv.Name
		(*c)[i].Writer = nv.Writer
		(*c)[i].RecordURL = nv.Chapter.URL
		(*c)[i].UpdateTime = nv.Time
		novel.Pool.Put(nv)
		// 按更新时间倒序排列
		c.sortByUpdate()
		// 异步保存配置
		go func() {
			err = c.SaveConfig(cu, path)
		}()
		if err != nil {
			kitten.Error(ErrSave, err)
			continue
		}
		kitten.Info(`更新《`, nv.Name, `》成功喵！`)
	}
}

// String 实现 fmt.Stringer
func (b Book) String() string {
	return `《` + b.BookName + `》` + `
作者：　　	` + b.Writer + `
平台：　　	` + b.Platform + `
书号：　　	` + b.BookID + `
上次更新：	` + func() string {
		if b.UpdateTime.IsZero() {
			return Unknown
		}
		return b.UpdateTime.Format(times.Layout)
	}()
}
