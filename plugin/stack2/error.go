package stack2

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
)

type (
	// 已经加入叠猫猫
	alreadyJoinedError struct{}

	// 需要休息
	needRestError struct {
		time.Duration     // 剩余的休息时间
		w             int // 叠入猫猫的体重
		i             int // 叠入猫猫的索引
		bool              // 日常任务是否完成
	}

	// 叠猫猫失败
	stackError struct {
		*kitten.Messager
		m *meow
		strings.Builder
		l int
		n int
		r result
	}
)

// Error 实现 error
func (*alreadyJoinedError) Error() string {
	return `已经加入叠猫猫了喵！`
}

// *alreadyJoinedErr 的构造函数，已经加入叠猫猫
func alreadyJoined() *alreadyJoinedError {
	return new(alreadyJoinedError)
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
		return e.String()
	}
	w := e.m.Weight // 叠猫猫前的体重
	e.Grow(128)
	_, _ = e.WriteString(`叠猫猫失败，杂鱼～杂鱼❤`)
	switch e.r {
	case flat:
		// 如果平地摔
		exit(e.Messager, e.m, e.r, e.n) // 让失败的猫猫退出
		_, _ = fmt.Fprintf(e, `你平地摔了喵！需要休息 %s。
你的体重由 %.1f kg 变为 %.1f kg。`,
			times.ConvertTimeDuration(e.m.Time.Sub(time.Unix(e.Event.Time, 0))),
			i2f(w), i2f(e.m.Weight))
	case press:
		// 压坏了别的猫猫
		exit(e.Messager, e.m, e.r, e.n) // 让失败的猫猫退出
		_, _ = fmt.Fprintf(e, `有 %d 只猫猫被压坏了喵！需要休息一段时间。`, e.n)
		doClear(e.Messager, e.l, e.n, w, e.m, &e.Builder)
		for range e.n {
			_, _ = e.WriteRune('🙀')
		}
	case fall:
		// 摔坏了别的猫猫
		exit(e.Messager, e.m, e.r, e.l) // 让失败的猫猫退出
		_, _ = fmt.Fprintf(e, `上面 %d 只猫猫摔下去了喵！需要休息一段时间。`, e.n)
		doClear(e.Messager, e.l, e.n, w, e.m, &e.Builder)
		for range e.n {
			_, _ = e.WriteRune('😿')
		}
	default:
		_, _ = e.WriteString(`未知错误喵！`)
	}
	return e.String()
}

/*
*stackErr 的构造函数

	// 叠猫猫失败
	stackError struct {
		*kitten.Messager        // 上下文
		m                *meow  // 失败的猫猫
		l                int    // 叠猫猫队列高度
		n                int    // 造成别的猫猫退出的数量
		r                result // 退出原因
		strings.Builder         // 错误内容
	}
*/
func stack(msgr *kitten.Messager, m *meow, l, n int, r result) *stackError {
	go setCard(msgr, l-n)
	return &stackError{
		Messager: msgr,
		m:        m,
		l:        l,
		n:        n,
		r:        r,
	}
}
