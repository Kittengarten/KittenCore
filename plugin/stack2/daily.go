package stack2

import (
	"errors"
	"math/big"
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/msg"

	"github.com/wdvxdr1123/ZeroBot/message"
)

const dailyRatio = 1000 // 猫堆高度与加速的比例的比值的标准差

// 叠猫猫日常任务，会修改原数据
func (d *data) daily(handler *msg.Handler, r ...replacer) message.ID {
	if len(r) == 0 {
		r = []replacer{l10n(cat)}
	}
	m, err := d.pre(handler, r...) // 初始化自身
	select {
	case <-times.RandDelayRange(time.Second, 2*time.Second):
	// 初始化的错误与后面的错误需要有延迟
	case <-handler.Done():
		return handler.SendWithImageFail(handler.Err())
	}
	nre, ok := errors.AsType[*needRestError](err)
	if !ok {
		// 当前不在休息，不能进行日常任务，恢复后返回
		*d = slices.Concat(*d, data{m})
		return sendWithImageFail(handler, r[0], `当前不在休息，不能进行日常任务喵！`)
	}
	mrh := time.Duration(stackConfig.MinRestHours) * time.Hour
	if nre.Duration <= 2*mrh {
		// 剩余休息时间过短，不能进行日常任务，直接返回
		return sendWithImageFail(handler, r[0], `剩余休息时间过短，不能进行日常任务喵！`)
	}
	now := handler.Time()
	if equal.IsSameDate4AM((*d)[nre.i].Daily, now) {
		// 已经完成日常任务，不能进行日常任务，直接返回
		return sendWithImageFail(handler, r[0], `已经完成日常任务，不能重复进行喵！`)
	}
	// 执行加速，每 1 猫堆高度的加速幅度标准差为 dailyRatio 的倒数
	ratio := max(1, normal[int64](dailyRatio))
	//nolint:durationcheck
	(*d)[nre.i].Time = (*d)[nre.i].Time.Add(
		-max(mrh, min(nre.Duration-mrh,
			time.Duration(new(big.Int).Mul(big.NewInt(int64(nre.Duration)),
				big.NewInt(min(ratio, int64(len(d.getStack()))))).Int64()/ratio))))
	(*d)[nre.i].Daily = now
	// 完成日常任务后的猫猫
	after := (*d)[nre.i]
	// 清理过期玩家
	d.clear(handler, false)
	// 存储叠猫猫数据
	if err := dataPath.SaveWithContext(handler, d); err != nil {
		return sendWithImageFail(handler, r[0], `存储叠猫猫数据时发生错误喵！`, err)
	}
	return sendTextf(handler, r[0], true, `执行日常任务成功喵！
你剩余的休息时间变为 %s喵！`,
		times.ConvertTimeDuration(after.Time.Sub(now)),
	)
}
