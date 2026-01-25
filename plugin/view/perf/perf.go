// Package perf 主机性能
package perf

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/Kittengarten/KittenCore/kitten/core"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/stat"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/plugin/view/text"

	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

// 默认 API
const (
	port   = `:8086`
	urlStr = `http://localhost` + port
	api    = `/perf`
	view   = api + `/view`
)

// 状态等级下界
var perfBelowBound = [...]float64{
	0: 0,
	1: 0.1,
	2: 0.15,
	3: 0.2,
	4: 0.25,
	5: 0.3,
}

// ErrNoData 没有数据喵！
var ErrNoData = errors.New(`没有数据喵！`)

var (
	server bool // 工作模式
	work   = sync.OnceFunc(func() {
		// 设置超时
		shttp.SetTimeout(core.Timeout)
		// 恢复超时
		defer shttp.SetTimeout(shttp.Timeout)
		res, err := shttp.GET(urlStr + view)
		if err == nil {
			defer shttp.Clear(res)
			log.Info(`本地性能 API 服务端正常，本机作为客户端工作喵！`)
			return
		}
		server = true
		log.Info(`本地性能 API 服务端错误，本机作为服务端工作喵！`, err)
		utils.Go(`性能 API 服务端`, func() {
			if err := http.ListenAndServe(port, nil); err != nil {
				log.Error(err)
				return
			}
		})
		http.HandleFunc(view, viewHandler)
	}) // 工作初始化
)

func init() {
	utils.Go(`性能 API 初始化`, work)
}

// LogFile 日志文件
var LogFile fio.Path = `C:\Program Files (x86)\MSI Afterburner\HardwareMonitoring.hml`

// 本地查看服务端
func viewHandler(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte(viewString(context.Background(), CPUTemperature(LogFile))))
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

// Level 返回状态等级
func Level(ctx context.Context, ts string) int {
	return level(cpuPercent(ctx), percent(getMem(ctx)), ts)
}

// 获取状态等级
func level(cpu float64, mem float64, ts string) int {
	if ts != ErrNoData.Error() {
		log.Warn(ts)
		return 5
	}
	ti, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		ti = stat.Temperature
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
func ViewString(ctx context.Context, name, t, w string) string {
	info := viewString(ctx, t) + w + `

` + text.GetWTA(name)
	return info + func() string {
		if text.Export.Checker == nil {
			return ``
		}
		return text.Export.Check(ctx, name, info)
	}()
}

// 查看字符串
func viewString(ctx context.Context, t string) string {
	if !server {
		// 设置超时
		shttp.SetTimeout(core.Timeout)
		// 恢复超时
		defer shttp.SetTimeout(shttp.Timeout)
		res, err := shttp.GETData(urlStr + view)
		if err != nil {
			return err.Error()
		}
		return string(res)
	}
	mem := getMem(ctx)
	return fmt.Sprintf(`系统：  	%s
CPU：   	%.2f%%	（%s）
线程：  	%d
内存：  	%.1f%%	（%s）
%s
体温：  	%s℃`,
		osInfo(ctx),
		cpuPercent(ctx), cpuInfo(ctx),
		processCount(ctx),
		percent(mem), use(mem),
		diskUsedAll(ctx),
		t,
	)
}

// 系统信息
func osInfo(ctx context.Context) string {
	i, err := host.InfoWithContext(ctx)
	if err != nil {
		return err.Error()
	}
	return i.Platform + `（` + i.PlatformVersion + `）`
}

// CPU 信息
func cpuInfo(ctx context.Context) string {
	getCPUs := func(logical bool) string {
		i, err := cpu.CountsWithContext(ctx, logical)
		if err != nil {
			return err.Error()
		}
		return strconv.Itoa(i)
	}
	return func() string {
		c, err := cpu.InfoWithContext(ctx)
		if err != nil {
			return err.Error()
		}
		s := new(strings.Builder)
		s.Grow(32 * len(c))
		for _, v := range c {
			fmt.Fprint(s, strings.TrimSpace(v.ModelName), `，`)
		}
		return s.String()
	}() + getCPUs(false) + `C` + getCPUs(true) + `T，` + func() string {
		i, err := host.InfoWithContext(ctx)
		if err != nil {
			return ``
		}
		return i.KernelArch
	}()
}

// CPU 使用率 %
func cpuPercent(ctx context.Context) float64 {
	p, err := cpu.PercentWithContext(ctx, shttp.Timeout, false)
	if err != nil {
		log.Warnln(`获取 CPU 使用率失败了喵！`, err)
		return 0
	}
	var avg float64
	for _, c := range p {
		avg += c
	}
	return avg / float64(len(p))
}

// 进程数
func processCount(ctx context.Context) int {
	p, err := process.ProcessesWithContext(ctx)
	if err != nil {
		log.Warnln(`获取进程数失败了喵！`, err)
		return 0
	}
	return len(p)
}

// 内存使用调用
func getMem(ctx context.Context) *mem.VirtualMemoryStat {
	m, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		log.Warnln(`获取内存使用失败了喵！`, err)
		return m
	}
	return m
}

// 内存使用率 %
func percent(m *mem.VirtualMemoryStat) float64 {
	return 100 * float64(m.Total-m.Free) / float64(m.Total)
}

// 内存使用情况
func use(m *mem.VirtualMemoryStat) string {
	return humanize.IBytes(m.Total-m.Free) + ` / ` + humanize.IBytes(m.Total)
}

// 全部磁盘使用情况
func diskUsedAll(ctx context.Context) string {
	var (
		d = getDisk(ctx)
		s = new(strings.Builder)
	)
	s.Grow(32 * len(d))
	for i, u := range d {
		fmt.Fprintf(s, "磁盘 %d：	%.1f%%	（%s / %s，%s）\n",
			i, u.UsedPercent, humanize.IBytes(u.Used), humanize.IBytes(u.Total), u.Fstype)
	}
	return s.String()[:s.Len()-1]
}

// 磁盘使用调用
func getDisk(ctx context.Context) (d []*disk.UsageStat) {
	p, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		log.Warnln(`获取磁盘分区失败了喵！`, err)
		return
	}
	for _, s := range p {
		u, err := disk.UsageWithContext(ctx, s.Mountpoint)
		if err != nil {
			log.Warnln(`获取磁盘信息失败了喵！`, err)
			continue
		}
		u.Fstype = s.Fstype
		d = append(d, u)
	}
	return
}
