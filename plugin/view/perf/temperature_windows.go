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
	time.Sleep(time.Second)
	file, err := l.Load(false)
	if err != nil {
		return err.Error()
	}
	fileScanner := bufio.NewScanner(file)
	if err := file.Close(); err != nil {
		return err.Error()
	}
	fileScanner.Split(bufio.ScanLines)
	for index := 0; fileScanner.Scan(); {
		const offset = 2
		switch line := fileScanner.Text(); {
		case strings.HasPrefix(line, `02`):
			for i, v := range strings.Split(line, `,`) {
				if strings.TrimSpace(v) == `CPU temperature` {
					index = i - offset
					break
				}
			}
		case strings.HasPrefix(line, `80`):
			return strings.TrimSpace(strings.Split(line, `,`)[index+offset])
		}
	}
	return defaultTemperature
}
