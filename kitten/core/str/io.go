package str

import (
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// GetFileName 返回文件名
func GetFileName(file string) (filename, ext string) {
	ext = filepath.Ext(file)
	return strings.TrimSuffix(file, ext), ext
}

// SplitNumber 分离文件名（不含拓展名）末尾的数字
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

/*
Rename 在已知文件名重复的情况下，生成新的文件名

不能保证生成的文件名不重复，需要自行判断
*/
func Rename(file string) string {
	filename, ext := GetFileName(file)
	filename, num := SplitNumber(filename)
	if num == 0 {
		return filename + `1` + ext
	}
	return filename + strconv.Itoa(num+1) + ext
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
