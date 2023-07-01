// Package stack 叠猫猫
package stack

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/Kittengarten/KittenCore/kitten"

	log "github.com/sirupsen/logrus"
)

const (
	// ReplyServiceName 插件名
	ReplyServiceName             = `stack`
	brief                        = `一起来玩叠猫猫`
	dataFile         kitten.Path = `data.yaml` // 叠猫猫数据文件
	exitFile         kitten.Path = `exit.yaml` // 叠猫猫退出日志文件
)

var (
	configFile  = kitten.FilePath(`stack`, `config.yaml`)                    // 叠猫猫配置文件名
	imagePath   = kitten.FilePath(kitten.Path(kitten.Configs.Path), `image`) // 图片路径
	stackConfig Config                                                       // 叠猫猫配置文件
	mu          sync.Mutex
)

func init() {
	var (
		help = strings.Join([]string{`发送`,
			fmt.Sprintf(`%s叠猫猫 [参数]`, kitten.Configs.CommandPrefix),
			`参数可选：加入|退出|查看`,
			fmt.Sprintf(`叠猫猫每层高度有 %d%% 概率会失败`, stackConfig.FailPercent),
			fmt.Sprintf(`最多可以叠 %d 只猫猫哦`, stackConfig.MaxStack),
			fmt.Sprintf(`在叠猫猫队列中超过 %d 小时后，会自动退出`, stackConfig.MaxTime),
			fmt.Sprintf(`主动退出叠猫猫；试图压别的猫猫；被压超过 %d 次且位于下半部分；叠猫猫失败摔下来——这些情况需要 %d 小时后，才能再次加入`, stackConfig.MaxCount, stackConfig.GapTime),
		}, "\n")
		// 注册插件
		engine = control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
			DisableOnDefault:  false,
			Brief:             brief,
			Help:              help,
			PrivateDataFolder: `stack`,
		})
	)
	kitten.InitFile(configFile, `maxstack: 10 # 叠猫猫队列上限
maxtime: 2   # 叠猫猫时间上限（小时数）
gaptime: 1   # 叠猫猫主动退出或者被压坏后重新加入所需的时间（小时数）
outofstack: "不能再叠了，下面的猫猫会被压坏的喵！" # 叠猫猫队列已满的回复
maxcount: 5 # 被压次数上限
failpercent: 1 # 叠猫猫每层失败概率百分数`)
	stackConfig := loadConfig(configFile) // 加载叠猫猫配置文件
	go autoExit(dataFile, stackConfig, engine)
	go autoExit(exitFile, stackConfig, engine)

	engine.OnCommand(`叠猫猫`).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Minute, 5).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		mu.Lock()
		defer mu.Unlock()
		var (
			ag       = ctx.State[`args`].(string)
			data     = loadData(kitten.FilePath(kitten.Path(engine.DataFolder()), dataFile))
			dataExit = loadData(kitten.FilePath(kitten.Path(engine.DataFolder()), exitFile))
		)
		switch ag {
		case `加入`:
			data.in(dataExit, stackConfig, ctx, engine)
		case `退出`:
			data.exit(ctx, engine)
		case `查看`:
			data.view(ctx)
		default:
			ctx.Send(help)
		}
	})
}

// 加载叠猫猫数据
func loadData(path kitten.Path) (stackData Data) {
	yaml.Unmarshal(path.Read(), &stackData)
	return
}

// 加入叠猫猫
func (data Data) in(esc Data, stackConfig Config, ctx *zero.Ctx, e *control.Engine) {
	var (
		permit          = true
		ID              = ctx.Event.UserID
		report, reports string
	)
	for _, meow := range esc {
		if ID == meow.ID {
			report = fmt.Sprintf(`休息不足 %d 小时，不能加入喵！`, stackConfig.GapTime)
			permit = false
			log.Info(strconv.FormatInt(ID, 10) + report)
		}
	}
	for _, meow := range data {
		if ID == meow.ID {
			report = `已经加入叠猫猫了喵！`
			permit = false
			log.Info(strconv.FormatInt(ID, 10) + ` ` + report)
		}
	}
	if permit {
		if len(data) >= stackConfig.MaxStack {
			report = stackConfig.OutOfStack
			permit = false
			// 压猫猫
			if exitLabel := -1; checkStack(len(data) + 1) {
				// 只有下半的猫猫会被压坏
				for i := range data {
					if data[i].Count++; data[i].Count > stackConfig.MaxCount && i < len(data)/2 {
						exitLabel = i // 最上面一只压坏的猫猫的位置
					}
				}
				// 如果有猫猫被压坏
				if 0 <= exitLabel {
					exitData := data[:exitLabel+1]
					reports = strings.Join(exitData.reverse(), "\n")
					// 将被压坏的的猫猫记录至退出日志
					for _, kitten := range exitData {
						logExit(kitten.ID, ctx, e)
					}
					data = data[exitLabel+1:]
					save(data, kitten.FilePath(kitten.Path(e.DataFolder()), dataFile))
					report += fmt.Sprintf("\n\n压猫猫成功，下面的猫猫对你的好感度下降了！你在 %d 小时内无法加入叠猫猫。", stackConfig.GapTime)
					report += fmt.Sprintf("\n\n有 %d 只猫猫被压坏了喵！需要休息 %d 小时。\n%s", exitLabel+1, stackConfig.GapTime, reports)
				} else {
					save(data, kitten.FilePath(kitten.Path(e.DataFolder()), dataFile))
					report += fmt.Sprintf("\n\n压猫猫成功，下面的猫猫对你的好感度下降了！你在 %d 小时内无法加入叠猫猫。", stackConfig.GapTime)
				}
			} else {
				report += fmt.Sprintf("\n\n压猫猫失败了喵！你在 %d 小时内无法加入叠猫猫。", stackConfig.GapTime)
			}
			log.Info(strconv.FormatInt(ID, 10) + report)
			logExit(ID, ctx, e) // 将压猫猫的猫猫记录至退出日志
		} else if checkStack(len(data) + 1) {
			// 如果叠猫猫成功
			meow := Kitten{
				ID:   ID,
				Name: kitten.QQ(ID).GetTitle(ctx) + ctx.CardOrNickName(ID),
				Time: time.Unix(ctx.Event.Time, 0),
			}
			data = append(data, meow)
			save(data, kitten.FilePath(kitten.Path(e.DataFolder()), dataFile))
			report = fmt.Sprintf(`叠猫猫成功，目前处于队列中第 %d 位喵～`, len(data))
			log.Info(strconv.FormatInt(ID, 10) + report)
		} else {
			// 如果叠猫猫失败
			// 如果不是平地摔
			if len(data) != 0 {
				exitCount := int(math.Ceil(float64(len(data)) * rand.Float64()))
				if exitCount == 0 {
					exitCount = 1
				}
				exitData := data[len(data)-exitCount:]
				// 将摔下来的的猫猫记录至退出日志
				for _, kitten := range exitData {
					logExit(kitten.ID, ctx, e)
				}
				data = data[:len(data)-exitCount]
				save(data, kitten.FilePath(kitten.Path(e.DataFolder()), dataFile))
				report = fmt.Sprintf("叠猫猫失败，上面 %d 只猫猫摔下来了喵！需要休息 %d 小时。\n%s",
					exitCount, stackConfig.GapTime, strings.Join(exitData.reverse(), "\n"))
			} else {
				// 如果是平地摔
				report = fmt.Sprintf("叠猫猫失败，你平地摔了喵！需要休息 %d 小时。", stackConfig.GapTime)
			}
			permit = false
			log.Info(strconv.FormatInt(ID, 10) + report)
			logExit(ID, ctx, e) // 将叠猫猫失败的猫猫记录至退出日志
		}
	}
	send(ID, permit, ctx, report)
}

// 退出叠猫猫
func (data Data) exit(ctx *zero.Ctx, e *control.Engine) {
	var (
		permit  = true
		ID      = ctx.Event.UserID
		dataNew = data
		report  string
	)
	for i, meow := range data {
		if ID == meow.ID {
			dataNew = append(data[:i], data[i+1:]...)
		}
	}
	if len(dataNew) == len(data) {
		report = `没有加入叠猫猫，不能退出喵！`
		permit = false
		log.Warn(strconv.FormatInt(ID, 10) + report)
	} else {
		stackData, err := yaml.Marshal(dataNew)
		kitten.FilePath(kitten.Path(e.DataFolder()), dataFile).Write(stackData)
		if kitten.Check(err) {
			report = `退出叠猫猫成功喵！`
			log.Info(strconv.FormatInt(ID, 10) + report)
			logExit(ID, ctx, e)
		} else {
			report = `退出叠猫猫失败喵！`
			permit = false
			log.Warn(strconv.FormatInt(ID, 10) + report)
		}
		send(ID, permit, ctx, report)
	}
}

// 查看叠猫猫
func (data Data) view(ctx *zero.Ctx) {
	const report = `【叠猫猫队列】`
	var (
		dataString = data.reverse()                                                // 反序查看
		reports    = fmt.Sprintf("%s\n%s", report, strings.Join(dataString, "\n")) // 生成播报
	)
	if 0 >= len(data) {
		reports = fmt.Sprintf(`%s暂时没有猫猫哦`, reports)
	}
	ctx.Send(reports)
}

// 叠猫猫队列反序并写为字符串数组
func (data Data) reverse() (s []string) {
	for i := len(data) - 1; 0 <= i; i-- {
		s = append(s, fmt.Sprintf(`%s（%d）`, data[i].Name, data[i].ID))
	}
	return
}

// 叠猫猫数据文件存储，成功则返回 True
func save(data Data, path kitten.Path) (ok bool) {
	d, err := yaml.Marshal(data)
	path.Write(d)
	ok = kitten.Check(err)
	if !ok {
		log.Errorf("文件写入错误喵！\n%v", err)
		reciver := kitten.Configs.SuperUsers[0]
		if kitten.Bot != nil {
			kitten.Bot.SendPrivateMessage(reciver, `叠猫猫文件写入错误，请检查日志喵！`)
		}
	}
	return
}

// 自动退出队列
func autoExit(f kitten.Path, c Config, e *control.Engine) {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); !kitten.Check(err) {
			log.Error(err)
		}
	}()
	var limitTime time.Duration
	switch f {
	case dataFile:
		limitTime = time.Duration(c.MaxTime) * time.Hour
	case exitFile:
		limitTime = time.Duration(c.GapTime) * time.Hour
	}
	if limitTime == 0 {
		limitTime = time.Hour
	}
	for {
		mu.Lock()
		kitten.InitFile(kitten.FilePath(kitten.Path(e.DataFolder()), f), `[]`)
		var (
			data     = loadData(kitten.FilePath(kitten.Path(e.DataFolder()), f))
			dataNew  = data
			nextTime = time.Now().Add(limitTime)
		)
		if 0 < len(data) {
			if time.Since(data[0].Time) > limitTime {
				if 1 < len(data) {
					nextTime = data[1].Time.Add(limitTime)
				}
				dataNew = data[1:]
			} else {
				nextTime = data[0].Time.Add(limitTime)
			}
		}
		if len(dataNew) != len(data) {
			save(dataNew, kitten.FilePath(kitten.Path(e.DataFolder()), f))
		}
		mu.Unlock()
		log.Infof(`下次定时退出 %s 时间为：%s`, kitten.FilePath(kitten.Path(e.DataFolder()), f), nextTime.Format(`2006-01-02 15:04:05`))

		time.Sleep(time.Until(nextTime))
	}
}

// 记录至退出日志
func logExit(u int64, ctx *zero.Ctx, e *control.Engine) {
	var (
		dataExit = loadData(kitten.FilePath(kitten.Path(e.DataFolder()), exitFile))
		meowExit = Kitten{
			ID:   u,
			Time: time.Unix(ctx.Event.Time, 0),
			Name: kitten.QQ(u).GetTitle(ctx) + ctx.CardOrNickName(u),
		}
	)
	dataExit = append(dataExit, meowExit)
	save(dataExit, kitten.FilePath(kitten.Path(e.DataFolder()), exitFile))
}

// 根据高度 h 检查压猫猫或叠猫猫是否成功
func checkStack(h int) bool {
	if rand.Float64() < 0.01*float64(h*stackConfig.FailPercent) {
		return false
	}
	return true
}

// 发送叠猫猫结果
func send(u int64, p bool, ctx *zero.Ctx, r string) {
	if ctx.Event.DetailType == "private" {
		if p {
			ctx.Send(message.Text(r))
		} else {
			ctx.SendChain(imagePath.GetImage(`no.png`), message.Text(r))
		}
	} else {
		if p {
			ctx.SendChain(message.At(u), message.Text(r))
		} else {
			ctx.SendChain(message.At(u), imagePath.GetImage(`no.png`), message.Text(r))
		}
	}
}

// 加载叠猫猫配置
func loadConfig(configFile kitten.Path) (c Config) {
	yaml.Unmarshal(configFile.Read(), &c)
	return
}
