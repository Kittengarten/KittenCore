// Package perf 查看服务器运行状况
package perf

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/Kittengarten/KittenCore/kitten"

	log "github.com/sirupsen/logrus"
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
	ReplyServiceName             = "查看"
	brief                        = "查看运行状况"
	filePath                     = "C:\\Program Files (x86)\\MSI Afterburner\\HardwareMonitoring.hml" // 温度配置文件路径
	imagePath        kitten.Path = "perf/path.txt"                                                    // 保存图片路径的文件
)

func init() {
	var (
		help = strings.Join([]string{"发送",
			fmt.Sprintf("%s查看%s，可获取服务器运行状况", kitten.Configs.CommandPrefix, kitten.Configs.NickName[0]),
		}, "\n")
		// 注册插件
		engine = control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
			DisableOnDefault: false,
			Brief:            brief,
			Help:             help,
		}).ApplySingle(ctxext.DefaultSingle)
	)

	// 查看功能
	engine.OnCommand("查看").SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Minute, 1).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		var (
			who              = ctx.State["args"].(string)
			str, pingMessage string
			report           message.Message
		)
		switch who {
		case zero.BotConfig.NickName[0]:
			var (
				cpu                                      = getCPUPercent()
				mem                                      = getMemPercent()
				t                                        = getCPUTemperature()
				ping                                     = kitten.LoadConfig().WebSocket.URL.CheckServer()
				annoStr, flower, elemental, imagery, err = kitten.GetWTAAnno()
				reportAnno                               string
			)
			if 0 >= ping {
				pingMessage = "连接超时喵！"
			} else if 1 > ping {
				pingMessage = "延迟：< 1 ms"
			} else {
				pingMessage = fmt.Sprintf("延迟：%d ms", ping)
			}
			// 查看性能页
			if !kitten.Check(err) {
				log.Errorf("报时失败喵！\n%v", err)
				reportAnno = "喵？"
			} else {
				reportAnno = strings.Join([]string{fmt.Sprintf("%s报时：现在是%s", zero.BotConfig.NickName[0], annoStr),
					fmt.Sprintf("花卉：%s", flower),
					fmt.Sprintf("～%s元灵之%s～", elemental, imagery),
				}, "\n")
			}
			str = strings.Join([]string{fmt.Sprintf("CPU 使用率：%.2f%%", cpu),
				fmt.Sprintf("内存使用：%.0f%%（%s）", mem, getMemUsed()),
				fmt.Sprintf("系统盘使用：%.2f%%（%s）", getDiskPercent(), getDiskUsed()),
				fmt.Sprintf("体温：%s℃", t),
				pingMessage,
				reportAnno,
			}, "\n")
			report = message.Message{imagePath.GetImage(strconv.Itoa(getPerf(cpu, mem, t)) + ".png"), message.Text(str)}
			ctx.Send(report)
		}
	})
}

// CPU使用率%
func getCPUPercent() float64 {
	percent, err := cpu.Percent(time.Second, false)
	if !kitten.Check(err) {
		log.Warnf("获取 CPU 使用率失败了喵！\n%v", err)
	}
	return percent[0]
}

// 内存使用调用
func getMem() (memInfo *mem.VirtualMemoryStat) {
	memInfo, err := mem.VirtualMemory()
	if !kitten.Check(err) {
		log.Warnf("获取内存使用失败了喵！\n%v", err)
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
		used  = fmt.Sprintf("%.2f MiB", float64(getMem().Used)/MiB)
		total = fmt.Sprintf("%.2f MiB", float64(getMem().Total)/MiB)
	)
	str = used + "/" + total
	return
}

// 磁盘使用调用
func getDisk() (diskInfo *disk.UsageStat) {
	parts, err1 := disk.Partitions(true)
	diskInfo, err2 := disk.Usage(parts[0].Mountpoint)
	if !(kitten.Check(err1) && kitten.Check(err2)) {
		log.Warnf("获取磁盘使用失败了喵！\n%v\n%v", err1, err2)
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
		used  = fmt.Sprintf("%.2f GiB", float64(getDisk().Used)/GiB)
		total = fmt.Sprintf("%.2f GiB", float64(getDisk().Total)/GiB)
	)
	str = used + "/" + total
	return
}

// 获取CPU温度
func getCPUTemperature() (CPUTemperature kitten.IntString) {
	os.Remove(filePath)
	time.Sleep(1 * time.Second)
	file, err := os.ReadFile(filePath)
	if !kitten.Check(err) {
		log.Warnf("获取 CPU 温度日志失败了喵！\n%v", err)
	}
	CPUTemperature = kitten.IntString(file[329:331])
	return
}

// 返回状态等级
func getPerf(cpu float64, mem float64, t kitten.IntString) int {
	if tt := float64(t.Int()); 0 < tt && 100 > tt {
		perf := 0.00005 * (cpu + mem) * tt
		log.Tracef("%s的负荷评分是 %f……", zero.BotConfig.NickName[0], perf)
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
	return 5
}
