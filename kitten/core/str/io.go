package str

import (
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// GetFileName 返回文件名
func GetFileName(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// SplitNumber 分离文件名末尾的数字
func SplitNumber(filename string) (string, int) {
	if filename == `` {
		return filename, 0
	}
	var (
		s   = []rune(filename)
		end = len(s) // 记录数字的结束索引
	)
	for i := len(s) - 1; i >= 0; i-- {
		if !unicode.IsDigit(s[i]) {
			break // 遇到非数字停止
		}
		end = i
	}
	num, err := strconv.Atoi(string(s[end:]))
	if err != nil {
		return filename, 0
	}
	return string(s[:end]), num // 返回数字部分
}

// NoDuplicate 在已知重复的情况下，生成不重复的文件名
func NoDuplicate(filename string) string {
	filename, num := SplitNumber(filename)
	if num == 0 {
		return filename + `1`
	}
	return filename + strconv.Itoa(num+1)
}

// HandleFileName 处理文件名中不支持的字符
func HandleFileName(filename string) string {
	return strings.NewReplacer(
		`\`, `_`,
		`/`, `_`,
		`:`, `_`,
		`*`, `_`,
		`?`, `_`,
		`"`, `_`,
		`<`, `_`,
		`>`, `_`,
		`|`, `_`,
	).Replace(filename)
}
