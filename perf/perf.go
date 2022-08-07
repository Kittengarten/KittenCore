// Package perf 查看服务器运行状况
package perf

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/go-ping/ping"
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
	ReplyServiceName = "查看"
	filePath         = "C:\\Program Files (x86)\\MSI Afterburner\\HardwareMonitoring.hml" // 温度配置文件路径
	imagePath        = "perf/path.txt"                                                    // 保存图片路径的文件
)

func init() {
	config := kitten.LoadConfig()
	// 注册插件
	help := strings.Join([]string{"发送",
		fmt.Sprintf("%s查看%s，可获取服务器运行状况", config.CommandPrefix, config.NickName[0]),
	}, "\n")
	engine := control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             help,
	})

	// 查看功能
	engine.OnCommand("查看").Handle(func(ctx *zero.Ctx) {
		who := ctx.State["args"].(string)
		var str, pingMessage string
		var report message.Message
		switch who {
		case config.NickName[0]:
			cpu := getCPUPercent()
			mem := getMemPercent()
			t := getCPUTemperature()
			ping := checkServer(config.WebSocket.URL)
			if ping <= 0 {
				pingMessage = "连接超时喵！"
			} else if ping < 1 {
				pingMessage = "延迟：< 1 ms"
			} else {
				pingMessage = fmt.Sprintf("延迟：%d ms", ping)
			}
			// 查看性能页
			annoStr, err := kitten.GetWTAAnno()
			var reportAnno string
			if !kitten.Check(err) {
				log.Error("报时失败喵！", err)
				reportAnno = "喵？"
			} else {
				reportAnno = fmt.Sprintf("喵喵报时：现在是%s", annoStr)
			}
			str = strings.Join([]string{fmt.Sprintf("CPU 使用率：%.2f%%", cpu),
				fmt.Sprintf("内存使用：%.0f%%（%s）", mem, getMemUsed()),
				fmt.Sprintf("系统盘使用：%.2f%%（%s）", getDiskPercent(), getDiskUsed()),
				fmt.Sprintf("体温：%s℃", t),
				pingMessage,
				reportAnno,
			}, "\n")
			perf := getPerf(cpu, mem, t)
			report = message.Message{kitten.GetImage(imagePath, strconv.Itoa(perf)+".png"), message.Text(str)}
		}
		ctx.Send(report)
	})
}

// CPU使用率%
func getCPUPercent() float64 {
	percent, err := cpu.Percent(time.Second, false)
	if !kitten.Check(err) {
		log.Warn("获取 CPU 使用率失败了喵！", err)
	}
	return percent[0]
}

// 内存使用调用
func getMem() (memInfo *mem.VirtualMemoryStat) {
	memInfo, err := mem.VirtualMemory()
	if !kitten.Check(err) {
		log.Warn("获取内存使用失败了喵！", err)
	}
	return
}

// 内存使用率%
func getMemPercent() float64 {
	return getMem().UsedPercent
}

// 内存使用情况
func getMemUsed() string {
	used := fmt.Sprintf("%.2f MiB", float64(getMem().Used)/MiB)
	total := fmt.Sprintf("%.2f MiB", float64(getMem().Total)/MiB)
	str := used + "/" + total
	return str
}

// 磁盘使用调用
func getDisk() (diskInfo *disk.UsageStat) {
	parts, err1 := disk.Partitions(true)
	diskInfo, err2 := disk.Usage(parts[0].Mountpoint)
	if !(kitten.Check(err1) && kitten.Check(err2)) {
		log.Warn("获取磁盘使用失败了喵！", err1, err2)
	}
	return
}

// 系统盘使用率%
func getDiskPercent() float64 {
	return getDisk().UsedPercent
}

// 系统盘使用情况
func getDiskUsed() (str string) {
	used := fmt.Sprintf("%.2f GiB", float64(getDisk().Used)/GiB)
	total := fmt.Sprintf("%.2f GiB", float64(getDisk().Total)/GiB)
	str = used + "/" + total
	return
}

// 获取CPU温度
func getCPUTemperature() (CPUTemperature string) {
	os.Remove(filePath)
	time.Sleep(1 * time.Second)
	file, err := os.ReadFile(filePath)
	if !kitten.Check(err) {
		log.Warn("获取 CPU 温度日志失败了喵！", err)
	}
	CPUTemperature = string(file)[329:331]
	return
}

// 返回状态等级
func getPerf(cpu float64, mem float64, t string) int {
	tt := float64(kitten.Atoi(t))
	if 0 < tt && tt < 100 {
		perf := (cpu + mem) * tt / 20000
		log.Tracef("喵喵的负荷评分是 %f……", perf)
		switch {
		case perf < 0.1:
			return 0
		case perf < 0.15:
			return 1
		case perf < 0.2:
			return 2
		case perf < 0.25:
			return 3
		case perf < 0.3:
			return 4
		}
	}
	return 5
}

// 检查连接状况，错误则返回 -1，正常则返回延迟的毫秒数
func checkServer(url string) int64 {
	url = kitten.GetMidText("//", ":", url)
	log.Tracef("正在 Ping %s 喵……", url)
	pinger, err := ping.NewPinger(url)
	if !kitten.Check(err) {
		log.Warnf("Ping 出现错误了喵！", err)
		return -1
	}
	pinger.Count = 1             // 检测 1 次
	pinger.Timeout = time.Second // 超时为 1 秒
	pinger.SetPrivileged(true)
	pinger.Run() // 直到完成之前，阻塞
	return pinger.Statistics().AvgRtt.Milliseconds()
}
