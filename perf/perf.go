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
	kittenConfig := kitten.LoadConfig()
	// 注册插件
	help := strings.Join([]string{"发送",
		fmt.Sprintf("%s查看%s，可获取服务器运行状况", config.CommandPrefix, kittenConfig.NickName[0]),
	}, "\n")
	engine := control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             help,
	})

	// 查看功能
	engine.OnCommand("查看").Handle(func(ctx *zero.Ctx) {
		who := ctx.State["args"].(string)
		var str string
		var report message.Message
		switch who {
		case kittenConfig.NickName[0]:
			cpu := getCPUPercent()
			mem := getMemPercent()
			t := getCPUTemperature()
			// 查看性能页
			str = strings.Join([]string{fmt.Sprintf("CPU使用率：%.2f%%", getCPUPercent()),
				fmt.Sprintf("内存使用：%.0f%%（%s）", getMemPercent(), getMemUsed()),
				fmt.Sprintf("系统盘使用：%.2f%%（%s）", getDiskPercent(), getDiskUsed()),
				fmt.Sprintf("体温：%s℃", getCPUTemperature()),
			}, "\n")
			report = make(message.Message, 2)
			perf := getPerf(cpu, mem, t)
			report[0] = kitten.GetImage(imagePath, strconv.Itoa(perf)+".png")
			report[1] = message.Text(str)
		}
		ctx.Send(report)
	})
}

// CPU使用率%
func getCPUPercent() float64 {
	percent, err := cpu.Percent(time.Second, false)
	if !kitten.Check(err) {
		log.Warn("获取CPU使用率失败了喵！")
	}
	return percent[0]
}

// 内存使用调用
func getMem() *mem.VirtualMemoryStat {
	memInfo, err := mem.VirtualMemory()
	if !kitten.Check(err) {
		log.Warn("获取内存使用失败了喵！")
	}
	return memInfo
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
func getDisk() *disk.UsageStat {
	parts, err1 := disk.Partitions(true)
	diskInfo, err2 := disk.Usage(parts[0].Mountpoint)
	if !(kitten.Check(err1) && kitten.Check(err2)) {
		log.Warn("获取磁盘使用失败了喵！")
	}
	return diskInfo
}

// 系统盘使用率%
func getDiskPercent() float64 {
	return getDisk().UsedPercent
}

// 系统盘使用情况
func getDiskUsed() string {
	used := fmt.Sprintf("%.2f GiB", float64(getDisk().Used)/GiB)
	total := fmt.Sprintf("%.2f GiB", float64(getDisk().Total)/GiB)
	str := used + "/" + total
	return str
}

// 获取CPU温度
func getCPUTemperature() string {
	os.Remove(filePath)
	time.Sleep(1 * time.Second)
	file, err := os.ReadFile(filePath)
	if !kitten.Check(err) {
		log.Warn("获取CPU温度日志失败了喵！")
	}
	CPUTemperature := string(file)[329:331]
	return CPUTemperature
}

// 返回状态等级
func getPerf(cpu float64, mem float64, t string) int {
	tt := float64(kitten.Atoi(t))
	if 0 < tt && tt < 100 {
		perf := (cpu + mem) * tt / 20000
		log.Trace(perf)
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
