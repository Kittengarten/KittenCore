// Package mode 运行模式
package mode

import (
	"os"
	"strings"
)

// 是否为测试模式
var isTest bool

func init() {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, `-test.`) {
			isTest = true
			break
		}
	}
}

// Test 是否为测试模式
func Test() bool {
	return isTest
}
