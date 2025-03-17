package utils

import (
	"bufio"
	"cmp"
	"fmt"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	human "github.com/dustin/go-humanize"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/sensors"

	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	defaultTemperature = `45` // 默认温度
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

// Check 检查
var Check = func(name, info string) string {
	// 默认为空实现
	return ``
}

// 返回查看字符串
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
			t, Weight(),
			getWTA(msgr))
	)
	return s + Check(name, s)
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
			fmt.Fprintf(&s, strings.TrimSpace(v.ModelName), `，`)
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
	p, err := cpu.Percent(core.TimeOutSeconds*time.Second, false)
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

// 全部磁盘使用情况
func diskUsedAll() string {
	var (
		b strings.Builder
		d = getDisk()
	)
	for i, s := range d {
		fmt.Fprintf(&b, "磁盘 %d：	%.1f%%	（%s / %s，%s）\n",
			i, s.UsedPercent, human.IBytes(s.Used), human.IBytes(s.Total), s.Fstype)
	}
	return b.String()[:b.Len()-1]
}

// 获取 CPU 温度（默认为所有传感器温度中最高的）
func cpuTemperature(l core.Path) string {
	t, err := sensors.SensorsTemperatures()
	if err != nil {
		kitten.Warn(err)
		switch runtime.GOOS {
		case `windows`:
			return cpuTemperatureOnWindows(l)
		default:
			return defaultTemperature
		}
	}
	return strconv.FormatFloat(slices.MaxFunc(t, func(i, j sensors.TemperatureStat) int {
		return cmp.Compare(i.Temperature, j.Temperature)
	}).Temperature, 'f', 2, 64)
}

// Windows 系统下获取 CPU 温度，通过微星小飞机（需要自行安装配置，并确保温度在其 log 中的位置）
func cpuTemperatureOnWindows(l core.Path) string {
	if err := l.Delete(); err != nil {
		kitten.Error(err)
		return err.Error()
	}
	<-time.NewTimer(time.Second).C
	file, err := l.Load(false)
	if err != nil {
		return err.Error()
	}
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	for index := 0; fileScanner.Scan(); {
		const offset = 2
		switch s := fileScanner.Text(); {
		case strings.HasPrefix(s, `02`):
			for i, v := range strings.Split(s, `,`) {
				if strings.TrimSpace(v) == `CPU temperature` {
					index = i - offset
					break
				}
			}
		case strings.HasPrefix(s, `80`):
			return strings.TrimSpace(strings.Split(s, `,`)[index+offset])
		}
	}
	return defaultTemperature
}

// 返回状态等级
func getPerf(cpu float64, mem float64, ts string) int {
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

// Ping
func Ping(msgr *kitten.Messager) message.ID {
	pingURL := msgr.Args()
	pg, err := probing.NewPinger(pingURL)
	if err != nil {
		return msgr.Reply().AtLf().Image(`哈——？.png`).Text(err).Send()
	}
	pg.Count = 4                                                             // 检测 4 次
	pg.Timeout = time.Duration(pg.Count) * core.TimeOutSeconds * time.Second // 超时时间设置
	var nbytes int
	pg.OnSend = func(pkt *probing.Packet) {
		nbytes = pkt.Nbytes
	}
	var pm strings.Builder
	pg.OnRecv = func(pkt *probing.Packet) {
		fmt.Fprintf(&pm, `来自 %s 的回复：字节=%d 时间=%dms TTL=%d
`, pkt.IPAddr, pkt.Nbytes, pkt.Rtt.Milliseconds(), pkt.TTL)
	}
	var r strings.Builder
	r.Grow(32 + 32*pg.Count)
	pg.OnFinish = func(st *probing.Statistics) {
		fmt.Fprintf(&r, `正在 Ping %s [%s] 具有 %d 字节的数据：
%v
%s 的 Ping 统计信息：
数据包：已发送 = %d，已接收 = %d，丢失 = %d（%.0f%% 丢失），
`,
			pingURL, st.IPAddr, nbytes,
			&pm,
			st.IPAddr,
			st.PacketsSent, st.PacketsRecv, st.PacketsSent-st.PacketsRecv, st.PacketLoss)
		if st.PacketLoss < 100 {
			fmt.Fprintf(&r, `往返行程的估计时间：
最短 = %dms，最长 = %dms，平均 = %dms`,
				st.MinRtt.Milliseconds(), st.MaxRtt.Milliseconds(), st.AvgRtt.Milliseconds())
		}
	}
	if err := pg.Run(); err != nil {
		return msgr.SendWithImageFail(err)
	}
	return msgr.Reply().AtLf().Text(&r).Send()
}
