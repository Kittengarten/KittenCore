//go:build windows

package perf

import (
	"bufio"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
)

// Windows 系统下获取 CPU 温度，通过微星小飞机（需要自行安装配置，并确保温度在其 log 中的位置）
func cpuTemperature(l fio.Path) string {
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
