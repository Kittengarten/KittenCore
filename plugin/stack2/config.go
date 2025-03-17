package stack2

import (
	_ "embed"
	"fmt"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	// 叠猫猫配置文件名
	configFile = core.FilePath(`data`, `Stack2`, `config.yaml`)
	//go:embed config.yaml
	configStr string
	// 叠猫猫配置文件
	stackConfig, err = core.Load[config](configFile, configStr)
	// bot 配置
	botConfig = kitten.MainConfig()
	// 帮助文本
	help = fmt.Sprintf(`%s%s%s %s|%s|%s|%s|%s

摔下去会导致体重减少，压坏会导致体重增加。
没有猫猫时，有(抱枕突破所需体重/当前体重)的概率发生平地摔导致体重变为 e 倍。
清空猫堆有(抱枕突破所需体重/当前体重)的概率触发特效导致体重变为 e 倍。

抱枕、奶猫和绒布球不会导致猫猫摔下去，但成为绒布球的休息时间更长，且会随着摔下去的猫堆高度成倍增加。
猫娘能保护身边的猫猫。随着猫娘的成长，她们的能力也会越来越强。
直接在猫娘以上级别的身上叠猫猫必定不会摔下去；
猫娘萝莉以上可以保护上面一只猫猫不被压坏；
猫娘少女和成年猫娘以上能保护下面一只猫猫不摔下去，且享有分析图片特权。
老虎可以吃猫猫，但被压坏的概率会发生总体不利的变化。
大老虎可以锻炼使休息时间减少到 %d 小时，但每减少 1 小时将减少 0.1 kg 体重。

压坏了别的猫猫；
被别的猫猫压坏；
叠猫猫失败摔下去；
平地摔——
这些情况需要休息 | N(0, 体重²) |（至少为 %d 小时，至多为 [最大休息时间]）后，才能再次加入。`,
		botConfig.CommandPrefix, cStack, cMeow, cIn, cView, cAnalysis, cRank, cOCCat,
		stackConfig.OCMinRestHours,
		stackConfig.MinRestHours)
	// 吃猫猫帮助文本
	helpEat = fmt.Sprintf(`%s%s%s
需要休息 | N(0, (e*体重)²) |（至少为 %d 小时，至多为 [最大休息时间]）后，才能再次加入`,
		botConfig.CommandPrefix, cEat, cMeow,
		stackConfig.MinRestHours)
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            brief,
		Help:             help,
		PublicDataFolder: `Stack2`,
	})
	// 图片路径
	imagePath = core.FilePath(kitten.ImagePath(), replyServiceName)
	// 数据路径
	dataPath = core.FilePath(engine.DataFolder(), dataFile)
	// 缓存路径
	bufferPath = core.FilePath(engine.DataFolder(), bufferFile)
	// 小贴士路径
	tipsPath = core.FilePath(engine.DataFolder(), tipsFile)
)
