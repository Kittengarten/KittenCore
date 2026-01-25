package stack2

import (
	"errors"
	"math"
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/msg"
)

// 加速叠猫猫，会修改原数据
func (d *data) oc(handler *msg.Handler) {
	var (
		m, err  = d.pre(handler)                        // 初始化自身
		nre, ok = errors.AsType[*needRestError](err) // 错误是否需要休息
	)
	if !ok {
		// 当前不在休息，不需要加速，恢复后返回
		*d = slices.Concat(*d, data{m})
		sendWithImageFail(handler, `当前不在休息，不能进行锻炼喵！`)
		return
	}
	if (*d)[nre.i].getTypeID(handler) < 大老虎 {
		// 不是大老虎以上，不能加速，直接返回
		sendWithImageFail(handler, `大老虎以上才可以锻炼——`)
		return
	}
	hours := int(math.RoundToEven(
		float64(nre.Duration-time.Hour*time.Duration(stackConfig.OCMinRestHours)) /
			float64(time.Hour))) // 加速的小时数
	if hours <= 0 {
		// 加速的小时数不大于 0，则不能加速，直接返回
		sendWithImageFail(handler, `剩余休息时间过短，不能锻炼喵！`)
		return
	}
	if (*d)[nre.i].Weight-hours < 1 {
		// 体重不足，则不能加速，直接返回
		sendWithImageFail(handler, `猫猫体重不足，锻炼失败喵！`)
		return
	}
	// 加速的代价
	(*d)[nre.i].Weight -= hours
	// 执行加速
	(*d)[nre.i].Time = (*d)[nre.i].Time.Add(time.Hour * time.Duration(-hours))
	// 付出加速代价后的猫猫
	after := (*d)[nre.i]
	// 清理过期玩家
	d.clear(handler, false)
	// 存储叠猫猫数据
	if err := fio.SaveWithContext(handler, dataPath, d); err != nil {
		sendWithImageFail(handler, `存储叠猫猫数据时发生错误喵！`, err)
	}
	_ = sendTextf(handler, true, `锻炼成功喵！
你剩余的休息时间变为 %s喵！
你的体重减少至 %.1f kg 喵！`,
		times.ConvertTimeDuration(after.Time.Sub(time.Unix(handler.Event().Time, 0))),
		i2f(after.Weight),
	)
}
