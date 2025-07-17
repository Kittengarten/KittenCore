// Package perf 主机性能
package perf

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/plugin/view/text"

	human "github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// 默认温度
const defaultTemperature = `45`

// 状态等级下界
var perfBelowBound = [...]float64{
	0: 0,
	1: 0.1,
	2: 0.15,
	3: 0.2,
	4: 0.25,
	5: 0.3,
}

// Level 返回状态等级
func Level(logFile fio.Path) int {
	return level(cpuPercent(), percent(getMem()), cpuTemperature(logFile))
}

// 获取状态等级
func level(cpu float64, mem float64, ts string) int {
	ti, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		kitten.Warn(err)
		return 5
	}
	if ti <= 0 || 100 <= ti {
		return 5
	}
	perf := 0.00005 * (cpu + mem) * ti
	for p, b := range slices.Backward(perfBelowBound[:]) {
		if perf > b {
			return p
		}
	}
	return 0
}

// ViewString 返回查看字符串
func ViewString(msgr *kitten.Messager, name string, logFile fio.Path) string {
	return viewString(msgr, name, cpuTemperature(logFile))
}

// 获取查看字符串
func viewString(msgr *kitten.Messager, name, t string) string {
	var (
		mem = getMem()
		s   = fmt.Sprintf(`系统：  	%s
CPU：   	%.2f%%  （%s）
内存：  	%.1f%%  （%s）
%s
体温：  	%s℃%s

%s`,
			osInfo(),
			cpuPercent(), cpuInfo(),
			percent(mem), use(mem),
			diskUsedAll(),
			t, text.Weight(),
			text.GetWTA(msgr))
	)
	return s + func() string {
		if text.Export.Checker == nil {
			return ``
		}
		return text.Export.Check(name, s)
	}()
}

// 系统信息
func osInfo() string {
	i, err := host.Info()
	if err != nil {
		return err.Error()
	}
	return i.Platform + `（` + i.PlatformVersion + `）`
}

// CPU 信息
func cpuInfo() string {
	getCPUs := func(logical bool) string {
		i, err := cpu.Counts(logical)
		if err != nil {
			return err.Error()
		}
		return strconv.Itoa(i)
	}
	return func() string {
		c, err := cpu.Info()
		if err != nil {
			return err.Error()
		}
		var s strings.Builder
		for _, v := range c {
			fmt.Fprint(&s, strings.TrimSpace(v.ModelName), `，`)
		}
		return s.String()
	}() + getCPUs(false) + `C` + getCPUs(true) + `T，` + func() string {
		i, err := host.Info()
		if err != nil {
			return ``
		}
		return i.KernelArch
	}()
}

// CPU 使用率 %
func cpuPercent() float64 {
	p, err := cpu.Percent(shttp.TimeOut, false)
	if err != nil {
		kitten.Warnln(`获取 CPU 使用率失败了喵！`, err)
		return 0
	}
	var avg float64
	for _, c := range p {
		avg += c
	}
	return avg / float64(len(p))
}

// 内存使用调用
func getMem() *mem.VirtualMemoryStat {
	m, err := mem.VirtualMemory()
	if err != nil {
		kitten.Warnln(`获取内存使用失败了喵！`, err)
		return &mem.VirtualMemoryStat{}
	}
	return m
}

// 内存使用率 %
func percent(m *mem.VirtualMemoryStat) float64 {
	return 100 * float64(m.Total-m.Free) / float64(m.Total)
}

// 内存使用情况
func use(m *mem.VirtualMemoryStat) string {
	return human.IBytes(m.Total-m.Free) + ` / ` + human.IBytes(m.Total)
}

// 全部磁盘使用情况
func diskUsedAll() string {
	var s strings.Builder
	for i, u := range getDisk() {
		fmt.Fprintf(&s, "磁盘 %d：	%.1f%%	（%s / %s，%s）\n",
			i, u.UsedPercent, human.IBytes(u.Used), human.IBytes(u.Total), u.Fstype)
	}
	return s.String()[:s.Len()-1]
}

// 磁盘使用调用
func getDisk() (d []*disk.UsageStat) {
	p, err := disk.Partitions(false)
	if err != nil {
		kitten.Warnln(`获取磁盘分区失败了喵！`, err)
		return
	}
	for _, s := range p {
		u, err := disk.Usage(s.Mountpoint)
		if err != nil {
			kitten.Warnln(`获取磁盘信息失败了喵！`, err)
			continue
		}
		u.Fstype = s.Fstype
		d = append(d, u)
	}
	return
}
