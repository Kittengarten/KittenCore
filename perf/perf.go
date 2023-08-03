// Package perf 查看服务器运行状况
package perf

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	probing "github.com/prometheus-community/pro-bing"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
	"go.uber.org/zap"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/Kittengarten/KittenCore/kitten"
)

const (
	_ = 1 << (10 * iota)
	// KiB 表示 KiB 所含字节数的变量
	KiB
	// MiB 表示 MiB 所含字节数的变量
	MiB
	// GiB 表示 GiB 所含字节数的变量
	GiB
	// TiB 表示 TiB 所含字节数的变量
	TiB
	// PiB 表示 PiB 所含字节数的变量
	PiB
	// EiB 表示 EiB 所含字节数的变量
	EiB
	// ZiB 表示 ZiB 所含字节数的变量
	ZiB
	// YiB 表示 YiB 所含字节数的变量
	YiB
	// ReplyServiceName 插件名
	ReplyServiceName             = `perf`
	brief                        = `查看运行状况`
	filePath         kitten.Path = `file.txt` // 保存微星小飞机温度配置文件路径的文件，非 Windows 系统或不使用可以忽略
	randMax                      = 100        // 随机数上限（不包含）
)

var (
	imagePath = kitten.FilePath(kitten.Path(kitten.Configs.Path), ReplyServiceName, `image`) // 图片路径
	poke      = rate.NewManager[int64](5*time.Minute, 9)                                     // 戳一戳
	nickname  = kitten.Configs.NickName[0]                                                   // 昵称
)

func init() {
	var (
		help = strings.Join([]string{`发送`,
			fmt.Sprintf(`%s查看%s，可获取服务器运行状况`, kitten.Configs.CommandPrefix, kitten.Configs.NickName[0]),
		}, "\n")
		// 注册插件
		engine = control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
			DisableOnDefault: false,
			Brief:            brief,
			Help:             help,
		}).ApplySingle(ctxext.DefaultSingle)
	)

	// 查看功能
	engine.OnCommand(`查看`).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Hour, 5).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		var (
			who    = ctx.State[`args`].(string)
			str    string
			report message.Message
		)
		for k := range zero.BotConfig.NickName {
			if who == zero.BotConfig.NickName[k] {
				who = nickname
			}
		}
		switch who {
		case nickname:
			var (
				cpu                                      = getCPUPercent()
				mem                                      = getMemPercent()
				t                                        = `45`
				annoStr, flower, elemental, imagery, err = kitten.GetWTAAnno()
				reportAnno                               string
			)
			// 查看性能页
			if !kitten.Check(err) {
				zap.S().Errorf("报时失败喵！\n%v", err)
				reportAnno = `喵？`
			} else {
				reportAnno = strings.Join([]string{fmt.Sprintf(`%s报时：现在是%s`, zero.BotConfig.NickName[0], annoStr),
					fmt.Sprintf(`花卉：%s`, flower),
					fmt.Sprintf(`～%s元灵之%s～`, elemental, imagery),
				}, "\n")
			}
			switch runtime.GOOS {
			case `windows`:
				t = getCPUTemperatureOnWindows(engine)
				str = strings.Join([]string{fmt.Sprintf(`CPU 使用率：%.2f%%`, cpu),
					fmt.Sprintf(`内存使用：%.0f%%（%s）`, mem, getMemUsed()),
					fmt.Sprintf(`系统盘使用：%.2f%%（%s）`, getDiskPercent(), getDiskUsed()),
					fmt.Sprintf(`体温：%s℃`, t),
					reportAnno,
				}, "\n")
			case `linux`:
				str = strings.Join([]string{fmt.Sprintf(`CPU 使用率：%.2f%%`, cpu),
					fmt.Sprintf(`内存使用：%.0f%%（%s）`, mem, getMemUsed()),
					fmt.Sprintf(`系统盘使用：%.2f%%（%s）`, getDiskPercent(), getDiskUsed()),
					reportAnno,
				}, "\n")
			default:
				str = strings.Join([]string{fmt.Sprintf(`CPU 使用率：%.2f%%`, cpu),
					fmt.Sprintf(`内存使用：%.0f%%（%s）`, mem, getMemUsed()),
					reportAnno,
				}, "\n")
			}
			report = message.Message{imagePath.GetImage(kitten.Path(strconv.Itoa(getPerf(cpu, mem, t)) + `.png`)), message.Text(str)}
			ctx.Send(report)
		default:
			kitten.DoNotKnow(ctx)
		}
	})

	// Ping 功能
	engine.OnCommandGroup([]string{`Ping`, `ping`}, zero.AdminPermission).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Minute, 1).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		var (
			pingURL = ctx.State[`args`].(string)
			report  string
			pingMsg string
			nbytes  int
		)
		pinger, err := probing.NewPinger(pingURL)
		pinger.Count = 4                  // 检测 4 次
		pinger.Timeout = 16 * time.Second // 超时时间设置
		if !kitten.Check(err) {
			kitten.DoNotKnow(ctx)
			return
		}
		pinger.OnSend = func(pkt *probing.Packet) {
			nbytes = pkt.Nbytes
		}
		pinger.OnRecv = func(pkt *probing.Packet) {
			pingMsg = strings.Join([]string{pingMsg,
				fmt.Sprintf(`来自 %s 的回复：字节=%d 时间=%dms TTL=%v`, pkt.IPAddr, pkt.Nbytes, pkt.Rtt.Milliseconds(), pkt.TTL),
			}, "\n")
		}
		pinger.OnFinish = func(stats *probing.Statistics) {
			report = strings.Join([]string{fmt.Sprintf(`正在 Ping %s [%s] 具有 %d 字节的数据：`, pingURL, stats.IPAddr, nbytes),
				pingMsg,
				``,
				fmt.Sprintf(`%s 的 Ping 统计信息：`, stats.IPAddr),
				fmt.Sprintf(`    数据包：已发送 = %d，已接收 = %d，丢失 = %d（%.0f%% 丢失）`,
					stats.PacketsSent, stats.PacketsRecv, stats.PacketsSent-stats.PacketsRecv, stats.PacketLoss),
			}, "\n")
			if 100 > stats.PacketLoss {
				report += strings.Join([]string{"\n往返行程的估计时间：",
					fmt.Sprintf(`    最短 = %dms，最长 = %dms，平均 = %dms`,
						stats.MinRtt.Milliseconds(), stats.MaxRtt.Milliseconds(), stats.AvgRtt.Milliseconds())}, "\n")
			}
		}
		err = pinger.Run()
		if kitten.Check(err) {
			kitten.SendText(ctx, true, report)
			return
		}
		kitten.SendTextOf(ctx, true, "Ping 出现错误：\n%v", err)
		zap.S().Warnf("Ping 出现错误：\n%v", err)
	})

	// 戳一戳
	engine.On(`notice/notify/poke`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var (
			g int64              // 本群的群号
			u = ctx.Event.UserID // 发出 poke 的 QQ 号
		)
		if `private` == ctx.Event.DetailType {
			g = -ctx.Event.UserID
		} else {
			g = ctx.Event.GroupID
		}
		switch {
		case poke.Load(g).AcquireN(5):
			// 5 分钟共 8 块命令牌 一次消耗 5 块命令牌
			ctx.SendChain(message.Poke(u))
		case poke.Load(g).AcquireN(3):
			// 5 分钟共 8 块命令牌 一次消耗 3 块命令牌
			kitten.SendTextOf(ctx, false, `请不要拍%s >_<`, nickname)
		case poke.Load(g).Acquire():
			// 5 分钟共 8 块命令牌 一次消耗 1 块命令牌
			kitten.SendTextOf(ctx, false, "喂(#`O′) 拍%s干嘛！\n（好感 - %d）", nickname, rand.Intn(randMax)+1)
		default:
			// 频繁触发，不回复
		}
	})

	// 图片，用于让 Bot 发送图片，可通过 CQ 码、链接等，为防止滥用，仅管理员可用
	zero.OnCommand(`图片`, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.Send(message.Image(ctx.State[`args`].(string)))
	})
}

// CPU使用率%
func getCPUPercent() float64 {
	var avg float64
	percent, err := cpu.Percent(time.Second, false)
	if !kitten.Check(err) {
		zap.S().Warnf("获取 CPU 使用率失败了喵！\n%v", err)
	}
	for k := range percent {
		avg += percent[k]
	}
	return avg / float64(len(percent))
}

// 内存使用调用
func getMem() (memInfo *mem.VirtualMemoryStat) {
	memInfo, err := mem.VirtualMemory()
	if !kitten.Check(err) {
		zap.S().Warnf("获取内存使用失败了喵！\n%v", err)
	}
	return
}

// 内存使用率%
func getMemPercent() float64 {
	return getMem().UsedPercent
}

// 内存使用情况
func getMemUsed() (str string) {
	var (
		used  = fmt.Sprintf(`%.2f MiB`, float64(getMem().Used)/MiB)
		total = fmt.Sprintf(`%.2f MiB`, float64(getMem().Total)/MiB)
	)
	str = used + `/` + total
	return
}

// 磁盘使用调用
func getDisk() (diskInfo *disk.UsageStat) {
	parts, err1 := disk.Partitions(false)
	diskInfo, err2 := disk.Usage(parts[0].Mountpoint)
	if !(kitten.Check(err1) && kitten.Check(err2)) {
		zap.S().Warnf("获取磁盘使用失败了喵！\n%v\n%v", err1, err2)
	}
	return
}

// 系统盘使用率%
func getDiskPercent() float64 {
	return getDisk().UsedPercent
}

// 系统盘使用情况
func getDiskUsed() (str string) {
	var (
		used  = fmt.Sprintf(`%.2f GiB`, float64(getDisk().Used)/GiB)
		total = fmt.Sprintf(`%.2f GiB`, float64(getDisk().Total)/GiB)
	)
	str = used + `/` + total
	return
}

// Windows 系统下获取 CPU 温度，通过微星小飞机（需要自行安装配置，并确保温度在其 log 中的位置）
func getCPUTemperatureOnWindows(e *control.Engine) (CPUTemperature string) {
	kitten.InitFile(kitten.FilePath(kitten.Path(e.DataFolder()), filePath), `C:\Program Files (x86)\MSI Afterburner\HardwareMonitoring.hml`)
	os.Remove(filePath.LoadPath().String())
	time.Sleep(1 * time.Second)
	file, err := os.ReadFile(filePath.LoadPath().String())
	if !kitten.Check(err) {
		zap.S().Warnf("获取 CPU 温度日志失败了喵！\n%v", err)
	}
	CPUTemperature = string(file[329:331]) // 此处为温度在微星小飞机 log 中的位置
	return
}

// 返回状态等级
func getPerf(cpu float64, mem float64, ts string) int {
	ti, err := strconv.Atoi(ts)
	if 0 < ti && 100 > ti && kitten.Check(err) {
		perf := 0.00005 * (cpu + mem) * float64(ti)
		zap.S().Debugf(`%s的负荷评分是 %f……`, zero.BotConfig.NickName[0], perf)
		switch {
		case 0.1 > perf:
			return 0
		case 0.15 > perf:
			return 1
		case 0.2 > perf:
			return 2
		case 0.25 > perf:
			return 3
		case 0.3 > perf:
			return 4
		}
	}
	zap.S().Warn(err)
	return 5
}
