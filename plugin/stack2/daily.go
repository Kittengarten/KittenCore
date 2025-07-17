package stack2

import (
	"errors"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
)

const dailyRatio = 1000 // 猫堆高度与加速的比例的比值

// 叠猫猫日常任务，会修改原数据
func (d *data) daily(msgr *kitten.Messager) {
	var (
		_, err = d.pre(msgr)              // 初始化自身
		nre    = needRest(0, 0, 0, false) // 默认错误：需要休息
	)
	times.RandomDelayRange(time.Second, 2*time.Second)
	if !errors.As(err, &nre) {
		// 当前不在休息，不能进行日常任务，直接返回
		sendWithImageFail(msgr, `当前不在休息，不能进行日常任务喵！`)
		return
	}
	if nre.Duration <= 2*time.Hour {
		// 剩余休息时间过短，不能进行日常任务，直接返回
		sendWithImageFail(msgr, `剩余休息时间过短，不能进行日常任务喵！`)
		return
	}
	if equal.IsSameDate4AM((*d)[nre.i].Daily, time.Unix(msgr.Event.Time, 0)) {
		// 已经完成日常任务，不能进行日常任务，直接返回
		sendWithImageFail(msgr, `已经完成日常任务，不能重复进行喵！`)
		return
	}
	// 执行加速
	//nolint:durationcheck
	(*d)[nre.i].Time = (*d)[nre.i].Time.Add(
		-max(time.Duration(stackConfig.MinRestHours)*time.Hour,
			min(nre.Duration-time.Duration(stackConfig.OCMinRestHours)*time.Hour,
				nre.Duration*min(dailyRatio,
					time.Duration(len(d.getStack())))/dailyRatio)))
	(*d)[nre.i].Daily = time.Unix(msgr.Event.Time, 0)
	// 完成日常任务后的猫猫
	after := (*d)[nre.i]
	// 清理过期玩家
	d.clear(msgr, false)
	// 存储叠猫猫数据
	if err := fio.Save(dataPath, d); err != nil {
		sendWithImageFail(msgr, `存储叠猫猫数据时发生错误喵！`, err)
	}
	_ = sendTextf(msgr, true, `执行日常任务成功喵！
你剩余的休息时间变为 %s喵！`,
		times.ConvertTimeDuration(after.Time.Sub(time.Unix(msgr.Event.Time, 0))),
	)
}
