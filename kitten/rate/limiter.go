// Package rate 限速器
package rate

import (
	"log/slog"
	"time"

	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
)

type (
	limiterBy   bool                              // 限速器基准
	limiterType byte                              // 限速器类型
	Limiter     func(ctx *zero.Ctx) *rate.Limiter // Limiter 限速器函数
)

const (
	ByGroup limiterBy = true  // 群内限速
	ByUser  limiterBy = false // 个人限速
)

const (
	Unknown     limiterType = iota // Unknown 未知限速器类型
	GroupNormal                    // GroupNormal 群内限速，每 3 分钟 1 次
	GroupFast                      // GroupFast 群内防刷屏限速，每 12 秒 1 次
	GroupSlow                      // GroupSlow 群内慢限速，每小时 1 次
	User                           // User 个人限速，每 12 分钟 1 次
)

var limiterStore = map[limiterType]Limiter{
	GroupNormal: ctxext.NewLimiterManager(3*time.Minute, 5).LimitByGroup,
	GroupFast:   ctxext.NewLimiterManager(12*time.Second, 5).LimitByGroup,
	GroupSlow:   ctxext.NewLimiterManager(time.Hour, 5).LimitByGroup,
	User:        ctxext.NewLimiterManager(12*time.Minute, 5).LimitByUser,
} // 共通限速器

func init() {
	ctxext.SetDefaultLimiterManagerParam(12*time.Second, 5)
}

// Get 获取共通限速器，o 为限速器类型
func (o limiterType) Get() Limiter {
	if lmt, ok := limiterStore[o]; ok {
		return lmt
	}
	// 如果获取限速器失败，则返回默认的个人限速器
	slog.Error(`获取限速器失败，请检查限速器类型喵！`)
	return ctxext.LimitByUser
}

// New 创建限速器，b 为限速器基准，interval 为限速器间隔，burst 为限速器容量
func New(b limiterBy, interval time.Duration, burst int) Limiter {
	m := ctxext.NewLimiterManager(interval, burst)
	switch b {
	case ByUser:
		return m.LimitByUser
	case ByGroup:
		return m.LimitByGroup
	default:
		// 死码，为了防止编译器报 missing return 而保留
		return m.LimitByUser
	}
}
