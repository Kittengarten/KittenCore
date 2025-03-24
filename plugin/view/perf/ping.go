package perf

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"

	probing "github.com/prometheus-community/pro-bing"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// Ping
func Ping(msgr *kitten.Messager) message.ID {
	pingURL := msgr.Args()
	pg, err := probing.NewPinger(pingURL)
	if err != nil {
		return msgr.Reply().AtLf().Image(`哈.png`).Text(err).Send()
	}
	pg.Count = 4                                                              // 检测 4 次
	pg.Timeout = time.Duration(pg.Count) * shttp.TimeOutSeconds * time.Second // 超时时间设置
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
