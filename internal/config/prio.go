package config

import (
	"bufio"
	_ "embed"
	"log/slog"
	"regexp"
	"strings"

	"github.com/FloatTech/zbputils/control"
)

//go:embed prio/main.go
var mainData string

// 加载插件优先级
func loadPrio() {
	var (
		re       = regexp.MustCompile(`^\t_ "github\.com/Kittengarten/KittenCore/internal/(\w+)"\s+// `)
		pluginRe = regexp.MustCompile(`^\t_ "github\.com/Kittengarten/KittenCore/plugin/(\w+)"\s+// `)
		zbpRe    = regexp.MustCompile(`^\t_ "github\.com/FloatTech/ZeroBot-Plugin/plugin/(\w+)"\s+// `)
	)
	prio := make(map[string]uint64, 0<<8)
	for scanner, i := bufio.NewScanner(strings.NewReader(mainData)), uint64(0); scanner.Scan(); i++ {
		var (
			line    = scanner.Text()
			matches = re.FindStringSubmatch(line)
		)
		if len(matches) >= 1 {
			prio[matches[1]] = i << 6
			continue
		}
		matches = pluginRe.FindStringSubmatch(line)
		if len(matches) >= 1 {
			prio[matches[1]] = i << 6
			continue
		}
		matches = zbpRe.FindStringSubmatch(line)
		if len(matches) >= 1 {
			prio[matches[1]] = i << 6
		}
	}
	control.LoadCustomPriority(prio)
	slog.Info(`插件优先级加载完成`, slog.Any(`配置`, prio))
}
