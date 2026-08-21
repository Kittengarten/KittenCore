package stack2

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"
)

type (
	// 已经加入叠猫猫
	alreadyJoinedError utils.Object

	// 需要休息
	needRestError struct {
		time.Duration     // 剩余的休息时间
		w             int // 叠入猫猫的体重
		i             int // 叠入猫猫的索引
		bool              // 日常任务是否完成
	}

	// 叠猫猫失败
	stackError struct {
		strings.Builder        // 错误内容
		*msg.Handler           // 上下文
		*meow                  // 失败的猫猫
		l               int    // 叠猫猫队列高度
		n               int    // 造成别的猫猫退出的数量
		r               result // 退出原因
	}
)

// Error 实现 error
func (*alreadyJoinedError) Error() string {
	return `已经加入叠猫猫了喵！`
}

// Error 实现 error
func (e *needRestError) Error() string {
	return fmt.Sprintf(`还需要休息 %s才能活动喵！
你的当前体重为 %.1f kg。
日常任务%s完成。`,
		times.ConvertTimeDuration(e.Duration),
		i2f(e.w),
		func() string {
			if e.bool {
				return `已`
			}
			return `未`
		}())
}

// *needRest 的构造函数，需要休息
func needRest(t time.Duration, w, i int, b bool) *needRestError {
	return &needRestError{
		Duration: t,
		w:        w,
		i:        i,
		bool:     b,
	}
}

// Error 实现 error
func (e *stackError) Error() string {
	if e.Len() != 0 {
		return e.Builder.String()
	}
	w := e.Weight // 叠猫猫前的体重
	e.Grow(1 << 7)
	_, _ = e.WriteString(`叠猫猫失败，杂鱼～杂鱼❤`)
	switch e.r {
	case flat:
		// 如果平地摔
		exit(e.Handler, e.meow, e.r, e.n) // 让失败的猫猫退出
		_, _ = fmt.Fprintf(e, `你平地摔了喵！
需要休息 %s。
你的体重由 %.1f kg 变为 %.1f kg。`,
			times.ConvertTimeDuration(e.Time.Sub(time.Unix(e.Event().Time, 0))),
			i2f(w), i2f(e.Weight))
	case press:
		// 压坏了别的猫猫
		exit(e.Handler, e.meow, e.r, e.n) // 让失败的猫猫退出
		_, _ = fmt.Fprintf(e, `有 %d 只猫猫被压坏了喵！
需要休息一段时间。
你的休息时长为 %s。`,
			e.n,
			times.ConvertTimeDuration(e.Time.Sub(time.Unix(e.Event().Time, 0))))
		doClear(e.Handler, e.l, e.n, w, e.meow, &e.Builder)
		for range e.n {
			_, _ = e.WriteRune('🙀')
		}
	case fall:
		// 摔坏了别的猫猫
		exit(e.Handler, e.meow, e.r, e.l) // 让失败的猫猫退出
		_, _ = fmt.Fprintf(e, `上面 %d 只猫猫摔下去了喵！
需要休息一段时间。
你的休息时长为 %s。`,
			e.n,
			times.ConvertTimeDuration(e.Time.Sub(time.Unix(e.Event().Time, 0))))
		doClear(e.Handler, e.l, e.n, w, e.meow, &e.Builder)
		for range e.n {
			_, _ = e.WriteRune('😿')
		}
	default:
		_, _ = e.WriteString(`未知错误喵！`)
	}
	return e.Builder.String()
}

// *stackErr 的构造函数
//
//	// 叠猫猫失败
//	type stackError struct {
//		strings.Builder         // 错误内容
//		*msg.Handler        // 上下文
//		*meow                   // 失败的猫猫
//		l                int    // 叠猫猫队列高度
//		n                int    // 造成别的猫猫退出的数量
//		r                result // 退出原因
//	}
func stack(handler *msg.Handler, m *meow, l, n int, r result) *stackError {
	utils.Go(`叠猫猫失败设置群昵称`, func() { setCard(handler, l-n) })
	return &stackError{
		Handler: handler,
		meow:    m,
		l:       l,
		n:       n,
		r:       r,
	}
}
