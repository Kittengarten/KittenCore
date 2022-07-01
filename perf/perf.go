package perf

import (
	"fmt"
	"kitten/kitten"
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
	replyServiceName = "Kitten_PerformanceBP" // 插件名
)

func init() {
	config := kitten.LoadConfig()
	help := "发送\n" + config.CommandPrefix + "查看" + config.NickName[0] + "，可获取服务器运行状况"
	engine := control.Register(replyServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             help,
	}) // 注册插件

	engine.OnCommand("查看").Handle(func(ctx *zero.Ctx) {
		config := kitten.LoadConfig()
		who := ctx.State["args"].(string)
		var str string
		switch who {
		case config.NickName[0]:
			str = "CPU使用率：" + Decimal(GetCpuPercent()) +
				"%，内存使用：" + Decimal(GetMemPercent()) +
				"%（" + GetMemUsed() + "），系统盘使用：" + Decimal(GetDiskPercent()) + "%（" + GetDiskUsed() + "），体温：" + GetCPUTemperature() + "℃" //查看性能页
		}
		ctx.Send(str)
	}) // 查看功能
}

func GetCpuPercent() float64 {
	percent, err := cpu.Percent(time.Second, false)
	if !kitten.Check(err) {
		log.Warn("获取CPU使用率失败了喵！")
	}
	return percent[0]
} // CPU使用率%

func GetMem() *mem.VirtualMemoryStat {
	memInfo, err := mem.VirtualMemory()
	if !kitten.Check(err) {
		log.Warn("获取内存使用失败了喵！")
	}
	return memInfo
} // 内存使用调用

func GetMemPercent() float64 {
	return GetMem().UsedPercent
} // 内存使用率%

func GetMemUsed() string {
	used := Decimal(float64(GetMem().Used)/1024/1024/1024) + " GiB"
	total := Decimal(float64(GetMem().Total)/1024/1024/1024) + " GiB"
	str := used + "/" + total
	return str
} // 内存使用情况

func GetDisk() *disk.UsageStat {
	parts, err1 := disk.Partitions(true)
	diskInfo, err2 := disk.Usage(parts[0].Mountpoint)
	if !(kitten.Check(err1) && kitten.Check(err2)) {
		log.Warn("获取磁盘使用失败了喵！")
	}
	return diskInfo
} // 磁盘使用调用

func GetDiskPercent() float64 {
	return GetDisk().UsedPercent
} // 系统盘使用率%

func GetDiskUsed() string {
	used := Decimal(float64(GetDisk().Used)/1024/1024/1024) + " GiB"
	total := Decimal(float64(GetDisk().Total)/1024/1024/1024) + " GiB"
	str := used + "/" + total
	return str
} // 系统盘使用情况

func Decimal(value float64) string {
	str := fmt.Sprintf("%.2f", value)
	return str
} // 转换为十进制数字字符串，并保留两位小数

func DecimalInt(value float64) string {
	str := fmt.Sprintf("%.0f", value)
	return str
} // 转换为十进制数字字符串，以整数形式输出

func GetCPUTemperature() string {
	filePath := "C:\\Program Files (x86)\\MSI Afterburner\\HardwareMonitoring.hml"
	os.Remove(filePath)
	time.Sleep(1 * time.Second)
	file, err := os.ReadFile(filePath)
	if !kitten.Check(err) {
		log.Warn("获取CPU温度日志失败了喵！")
	}
	CPUTemperature := string(file)[329:331]
	return CPUTemperature
} // 获取CPU温度
