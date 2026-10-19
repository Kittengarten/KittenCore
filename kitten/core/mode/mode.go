// Package mode 运行模式
package mode

import (
	"os"
	"strings"
)

// Test 是否为测试模式
func Test() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, `-test.`) {
			return true
		}
	}
	return false
}
