package stack

import (
	"kitten/kitten"
	"strconv"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"

	"gopkg.in/yaml.v3"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	log "github.com/sirupsen/logrus"
)

const (
	replyServiceName = "Kitten_StackBP" // 插件名
)

func init() {
	go AutoExit("stack/data.yaml", LoadConfig())
	go AutoExit("stack/exit.yaml", LoadConfig())

	stackConfig := LoadConfig()
	config := kitten.LoadConfig()
	help := "发送\n" + config.CommandPrefix + "叠猫猫 [参数]" +
		"\n参数可选：加入|退出|查看" +
		"\n最多可以叠" + strconv.Itoa(stackConfig.MaxStack) + "只猫猫哦" +
		"\n在叠猫猫队列中超过" + strconv.Itoa(stackConfig.MaxTime) + "小时后，会自动退出" +
		"\n主动退出叠猫猫，需要" + strconv.Itoa(stackConfig.GapTime) + "小时后，才能再次加入"
	// 注册插件
	engine := control.Register(replyServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             help,
	})

	engine.OnCommand("叠猫猫").Handle(func(ctx *zero.Ctx) {
		ag := ctx.State["args"].(string)
		data := LoadData("stack/data.yaml")
		dataExit := LoadData("stack/exit.yaml")
		switch ag {
		case "加入":
			In(data, dataExit, stackConfig, ctx)
		case "退出":
			Exit(data, dataExit, ctx)
		case "查看":
			View(data, ctx)
		default:
			ctx.SendGroupMessage(ctx.Event.GroupID, help)
		}
	})
}

// 加载叠猫猫配置
func LoadConfig() (stackConfig Config) {
	yaml.Unmarshal(kitten.FileRead("stack/config.yaml"), &stackConfig)
	return stackConfig
}

// 加载叠猫猫数据
func LoadData(path string) (stackData Data) {
	yaml.Unmarshal(kitten.FileRead(path), &stackData)
	return stackData
}

// 加入叠猫猫
func In(data Data, dataExit Data, stackConfig Config, ctx *zero.Ctx) {
	permit := true
	id := ctx.Event.UserID
	var report string

	for _, meow := range dataExit {
		if id == meow.Id {
			report = "退出叠猫猫不足" + strconv.Itoa(stackConfig.GapTime) + "小时，不能加入喵！"
			permit = false
			log.Info(strconv.FormatInt(id, 10) + report)
		}
	}
	for _, meow := range data {
		if id == meow.Id {
			report = "已经加入叠猫猫了喵！"
			permit = false
			log.Info(strconv.FormatInt(id, 10) + report)
		}
	}

	if permit {
		if len(data) >= stackConfig.MaxStack {
			log.Info(strconv.FormatInt(id, 10) + stackConfig.OutOfStack)
			report = stackConfig.OutOfStack
		} else {
			var meow Kitten
			meow.Id = id
			meow.Name = ctx.Event.Sender.Card
			if meow.Name == "" {
				meow.Name = ctx.Event.Sender.NickName
			}
			meow.Time = time.Unix(ctx.Event.Time, 0)
			data = append(data, meow)
			stackData, err1 := yaml.Marshal(data)
			err2 := kitten.FileWrite("stack/data.yaml", stackData)
			if !kitten.Check(err1) || !kitten.Check(err2) {
				report = "叠猫猫失败了喵！"
				log.Warn(strconv.FormatInt(id, 10) + report)
			} else {
				report = "叠猫猫成功，目前处于队列中第" + strconv.Itoa(len(data)) + "位喵～"
				log.Info(strconv.FormatInt(id, 10) + report)
			}
		}
	}
	messageToSend := message.Text(report)
	ctx.SendChain(message.At(id), messageToSend)
}

// 退出叠猫猫
func Exit(data Data, dataExit Data, ctx *zero.Ctx) {
	dataNew := data
	for idx, meow := range data {
		if ctx.Event.UserID == meow.Id {
			dataNew = append(data[:idx], data[idx+1:]...)
		}
	}
	if len(dataNew) == len(data) {
		report := "没有加入叠猫猫，不能退出喵！"
		log.Warn(strconv.FormatInt(ctx.Event.UserID, 10) + report)
		messageToSend := message.Text(report)
		ctx.SendChain(message.At(ctx.Event.UserID), messageToSend)
	} else {
		stackData, err1 := yaml.Marshal(dataNew)
		err2 := kitten.FileWrite("stack/data.yaml", stackData)
		if !kitten.Check(err1) || !kitten.Check(err2) {
			report := "退出叠猫猫失败喵！"
			log.Warn(strconv.FormatInt(ctx.Event.UserID, 10) + report)
			messageToSend := message.Text(report)
			ctx.SendChain(message.At(ctx.Event.UserID), messageToSend)
		} else {
			report := "退出叠猫猫成功喵！"
			log.Info(strconv.FormatInt(ctx.Event.UserID, 10) + report)
			messageToSend := message.Text(report)
			ctx.SendChain(message.At(ctx.Event.UserID), messageToSend)

			// 记录至退出日志
			var meowExit Kitten
			meowExit.Id = ctx.Event.UserID
			meowExit.Time = time.Unix(ctx.Event.Time, 0)
			dataExit = append(dataExit, meowExit)
			exitData, err1 := yaml.Marshal(dataExit)
			err2 := kitten.FileWrite("stack/exit.yaml", exitData)
			if !kitten.Check(err1) || !kitten.Check(err2) {
				report := "记录至退出日志失败了喵！"
				log.Warn(strconv.FormatInt(meowExit.Id, 10) + report)
			} else {
				report := "记录至退出日志成功喵！"
				log.Info(strconv.FormatInt(meowExit.Id, 10) + report)
			}
		}
	}
}

// 查看叠猫猫
func View(data Data, ctx *zero.Ctx) {
	reports := "【叠猫猫队列】"
	for idx := len(data) - 1; idx >= 0; idx-- {
		// data.yaml以时间正序存储，但以时间倒序查看
		report := "\n" + data[idx].Name + "（" + strconv.FormatInt(data[idx].Id, 10) + "）"
		reports += report
	}
	if reports == "【叠猫猫队列】" {
		reports += "\n暂时没有猫猫哦"
	}
	ctx.SendGroupMessage(ctx.Event.GroupID, reports)
}

// 自动退出队列
func AutoExit(path string, config Config) {
	// 处理panic，防止程序崩溃
	defer func() {
		if err := recover(); !kitten.Check(err) {
			log.Error(err)
		}
	}()

	var limitTimeHours int
	switch path {
	case "stack/data.yaml":
		limitTimeHours = config.MaxTime
	case "stack/exit.yaml":
		limitTimeHours = config.GapTime
	}

	for {
		data := LoadData(path)
		dataNew := data
		limitTime, _ := time.ParseDuration(strconv.Itoa(limitTimeHours) + "h")
		nextTime := time.Now().Add(limitTime)
		if len(data) > 0 {
			if time.Since(data[0].Time).Hours() > float64(limitTimeHours) {
				if len(data) > 1 {
					nextTime = data[1].Time.Add(limitTime)
				}
				dataNew = data[1:]
			}
		}
		if len(dataNew) != len(data) {
			stackData, err1 := yaml.Marshal(dataNew)
			err2 := kitten.FileWrite(path, stackData)
			if !kitten.Check(err1) || !kitten.Check(err2) {
				log.Warn("定时退出" + path + "失败喵！")
			} else {
				log.Info("定时退出" + path + "成功喵！")
			}
		}
		log.Info("下次定时退出时间为：" + nextTime.Format("2006-01-02 15:04:05"))
		time.Sleep(time.Until(nextTime))
	}
}
