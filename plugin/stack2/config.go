package stack2

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/internal/config"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"

	"github.com/goccy/go-yaml"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	baseChanceStr    = `[基础概率]` // 抱枕突破所需体重 ÷ 当前体重
	currentHeightStr = `[当前猫堆高度]`
	baseRestTimeStr  = `[基础休息时间]`
	lorryRestTimeStr = `[大运休息时间]`
	maxRestTimeStr   = `[最大休息时间]`
)

var (
	// 叠猫猫配置文件
	stackConfig, err = fio.NewPath(`data`, `Stack2`, `config.yaml`).
		Load[cfg](func() string {
			s := new(strings.Builder)
			s.Grow(61)
			if err := yaml.NewEncoder(s).Encode(cfg{
				RestHoursPerKG: 1,
				MinRestHours:   1,
				OCMinRestHours: 24,
			}); err != nil {
				log.Panic(err)
			}
			return s.String()
		}())
	// 指令前缀
	p = config.CommandPrefix()
	// 帮助文本
	help = fmt.Sprintf(`%s%s%s %s|%s|%s|%s|%s|%s

摔下去会导致体重减少，压坏会导致体重增加。
没有猫猫时，有`+baseChanceStr+`的概率发生平地摔导致体重变为 e 倍。
清空猫堆有`+baseChanceStr+`的概率触发特效导致体重变为 e 倍。

抱枕、奶猫和绒布球不会导致猫猫摔下去，但成为绒布球的休息时间更长，且会随着摔下去的猫堆高度成倍增加。
猫娘能保护身边的猫猫。随着猫娘的成长，她们的能力也会越来越强。
直接在猫娘以上级别的身上叠猫猫必定不会摔下去；
猫娘萝莉以上可以保护上面一只猫猫不被压坏；
猫娘少女和成年猫娘以上能保护下面一只猫猫不摔下去，且享有分析图片特权。
老虎可以吃猫猫，但被压坏的概率会发生总体不利的变化。
日常任务可以使休息时间减少 (%s / | N(0, %d²) |)（至少为 %d 小时，至多为剩余 %d 小时）。
大老虎可以锻炼使休息时间减少到 %d 小时，但每减少 1 小时将减少 0.1 kg 体重。

压坏了别的猫猫；
被别的猫猫压坏——
这些情况需要休息 | `+lorryRestTimeStr+` |（至少为 %d 小时，至多为 `+maxRestTimeStr+`）后，才能再次加入。

叠猫猫失败摔下去；
平地摔——
这些情况需要休息 | `+baseRestTimeStr+` |（至少为 %d 小时，至多为 `+maxRestTimeStr+`）后，才能再次加入。`,
		p, cStack, cMeow, cIn, cView, cAnalysis, cRank, cOC, cDaily,
		currentHeightStr, dailyRatio, stackConfig.MinRestHours, stackConfig.MinRestHours,
		stackConfig.OCMinRestHours,
		stackConfig.MinRestHours,
		stackConfig.MinRestHours)
	// 吃猫猫帮助文本
	helpEat = fmt.Sprintf(`%s%s%s
需要休息 | `+lorryRestTimeStr+` |（至少为 %d 小时，至多为 `+maxRestTimeStr+`）后，才能再次加入。`,
		p, cEat, cMeow,
		stackConfig.MinRestHours)
	// 撞大运帮助文本
	helpLorry = fmt.Sprintf(`%s%s
需要休息 | `+lorryRestTimeStr+` |（至少为 %d 小时，至多为 `+maxRestTimeStr+`）后，才能再次加入。`,
		p, cLorry,
		stackConfig.MinRestHours)
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            brief,
		Help:             help,
		PublicDataFolder: `Stack2`,
	})
	// 图片路径
	imagePath = fio.NewPath(replyServiceName)
	// 数据路径
	dataPath = fio.NewPath(engine.DataFolder(), dataFile)
	// 缓存路径
	bufferPath = fio.NewPath(engine.DataFolder(), bufferFile)
	// 小贴士路径
	tipsPath = fio.NewPath(engine.DataFolder(), tipsFile)
)

func init() {
	// 备份数据文件
	n, err := fio.NewPath(`data`, `Stack2`,
		`data-`+time.Now().Format(time.DateOnly)+`.yaml`).Copy(dataPath)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infoln(`备份配置文件`, n, `字节喵！`)
}
