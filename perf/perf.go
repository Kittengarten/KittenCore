package perf

import (
	"fmt"
	"kitten/kitten"
	"strings"

	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"

	log "github.com/sirupsen/logrus"
)

const (
	replyServiceName = "Kitten_PerformanceBP"                                             // 插件名
	filePath         = "C:\\Program Files (x86)\\MSI Afterburner\\HardwareMonitoring.hml" // 温度配置文件路径
)

func init() {
	config := zero.BotConfig
	kittenConfig := kitten.LoadConfig()
	// 注册插件
	help := strings.Join([]string{"发送",
		fmt.Sprintf("%s查看%s，可获取服务器运行状况", config.CommandPrefix, kittenConfig.NickName[0]),
	}, "\n")
	engine := control.Register(replyServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             help,
	})

	// 查看功能
	engine.OnCommand("查看").Handle(func(ctx *zero.Ctx) {
		who := ctx.State["args"].(string)
		var str string
		switch who {
		case kittenConfig.NickName[0]:
			// 查看性能页
			str = strings.Join([]string{fmt.Sprintf("CPU使用率：%s%%", Decimal(GetCpuPercent())),
				fmt.Sprintf("内存使用：%s%%（%s）", DecimalInt(GetMemPercent()), GetMemUsed()),
				fmt.Sprintf("系统盘使用：%s%%（%s）", Decimal(GetDiskPercent()), GetDiskUsed()),
				fmt.Sprintf("体温：%s℃", GetCPUTemperature()),
			}, "\n")
		}
		ctx.Send(str)
	})
}

// CPU使用率%
func GetCpuPercent() float64 {
	percent, err := cpu.Percent(time.Second, false)
	if !kitten.Check(err) {
		log.Warn("获取CPU使用率失败了喵！")
	}
	return percent[0]
}

// 内存使用调用
func GetMem() *mem.VirtualMemoryStat {
	memInfo, err := mem.VirtualMemory()
	if !kitten.Check(err) {
		log.Warn("获取内存使用失败了喵！")
	}
	return memInfo
}

// 内存使用率%
func GetMemPercent() float64 {
	return GetMem().UsedPercent
}

// 内存使用情况
func GetMemUsed() string {
	used := Decimal(float64(GetMem().Used)/1024/1024/1024) + " GiB"
	total := Decimal(float64(GetMem().Total)/1024/1024/1024) + " GiB"
	str := used + "/" + total
	return str
}

// 磁盘使用调用
func GetDisk() *disk.UsageStat {
	parts, err1 := disk.Partitions(true)
	diskInfo, err2 := disk.Usage(parts[0].Mountpoint)
	if !(kitten.Check(err1) && kitten.Check(err2)) {
		log.Warn("获取磁盘使用失败了喵！")
	}
	return diskInfo
}

// 系统盘使用率%
func GetDiskPercent() float64 {
	return GetDisk().UsedPercent
}

// 系统盘使用情况
func GetDiskUsed() string {
	used := Decimal(float64(GetDisk().Used)/1024/1024/1024) + " GiB"
	total := Decimal(float64(GetDisk().Total)/1024/1024/1024) + " GiB"
	str := used + "/" + total
	return str
}

// 转换为十进制数字字符串，并保留两位小数
func Decimal(value float64) string {
	str := fmt.Sprintf("%.2f", value)
	return str
}

// 转换为十进制数字字符串，以整数形式输出
func DecimalInt(value float64) string {
	str := fmt.Sprintf("%.0f", value)
	return str
}

// 获取CPU温度
func GetCPUTemperature() string {
	os.Remove(filePath)
	time.Sleep(1 * time.Second)
	file, err := os.ReadFile(filePath)
	if !kitten.Check(err) {
		log.Warn("获取CPU温度日志失败了喵！")
	}
	CPUTemperature := string(file)[329:331]
	return CPUTemperature
}
