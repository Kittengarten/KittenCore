package prio

import (
	"bufio"
	_ "embed"
	"log/slog"
	"regexp"
	"strings"

	"github.com/FloatTech/zbputils/control"
)

//go:embed main.prio
var mainData string

// 加载插件优先级
func init() {
	slog.Info(`插件自定义优先级开始加载`)
	var (
		re   = regexp.MustCompile(`^\t_ "(.+/)?([\w\-\.]+)"\s+// `)
		prio = make(map[string]uint64, 0<<8)
	)
	for scanner, i := bufio.NewScanner(strings.NewReader(mainData)), uint64(0); scanner.Scan(); i++ {
		if matches := re.FindStringSubmatch(scanner.Text()); len(matches) >= 2 {
			prio[matches[2]] = i << 6
		}
	}
	control.LoadCustomPriority(prio)
	slog.Info(`插件自定义优先级加载完成`, slog.Any(`数量`, len(prio)), slog.Any(`配置`, prio))
}
