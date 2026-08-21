// Package str 字符串处理
package str

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// TrimTooLong 截断过长内容
func TrimTooLong(content string, maxLen int) string {
	if utf8.RuneCountInString(content) <= maxLen {
		return content
	}
	return substr(content, maxLen-1) + `…`
}

// substr 截取字符串前 n 个字符
func substr(s string, n int) string {
	if n <= 0 {
		return ``
	}
	var i int
	for pos := range s {
		if i == n {
			return s[:pos]
		}
		i++
	}
	return s
}

// Clean 清理字符串中全部不必要内容
//
//	lf 控制是否换行
func Clean(s string, lf bool) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r':
			if lf {
				return r
			}
			// 移除换行符
			return -1
		}
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			// 移除换行符之外的控制字符和空白字符
			return -1
		}
		switch r {
		case '\u061c', '\u200e', '\u200f', '\u202a', '\u202b',
			'\u202c', '\u202d', '\u202e', '\u2066', '\u2067',
			'\u2068', '\u2069', '\ufffd':
			// 移除可能引发排版和显示错误的字符
			return -1
		}
		return r
	}, s))
}

// ClearRuneBytes 移除特定长度的 UTF-8 码点
func ClearRuneBytes(s string, n ...int) string {
	return strings.Map(func(r rune) rune {
		if slices.Contains(n, utf8.RuneLen(r)) {
			return -1
		}
		return r
	}, s)
}

// ComposeAuto 不需要提供 strings.Builder 的排版
func ComposeAuto(s string) string {
	b := new(strings.Builder)
	b.Grow(len(s))
	return Compose(b, s)
}

// Compose 排版
//
//	必须提供 strings.Builder，否则无效。应当进行预增长。
func Compose(b *strings.Builder, s string) string {
	if b == nil {
		return s
	}
	for line := range strings.SplitSeq(strings.TrimSpace(s), "\n") {
		if line = strings.TrimSpace(line); line == `` {
			continue
		}
		fmt.Fprintf(b, "　　%s\n", line)
	}
	return b.String()
}

// FirstHan 获取第一段中文字符码点组成的字符串
func FirstHan(s string) string {
	return First(s, isChinese)
}

// FirstText 获取第一段常用文字（不包括标点）组成的字符串
func FirstText(s string) string {
	return First(s,
		isChinese, isJapanese, isKorean,
		isRussian, isFrench, isArabic, isGreek,
		isN, unicode.IsLetter,
	)
}

// 是中文
func isChinese(r rune) bool {
	return unicode.Is(unicode.Han, r)
}

// 是假名
func isJapanese(r rune) bool {
	return unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r)
}

// 是谚文
func isKorean(r rune) bool {
	return unicode.Is(unicode.Hangul, r)
}

// 是西里尔字母
func isRussian(r rune) bool {
	return unicode.Is(unicode.Cyrillic, r)
}

// 是拉丁字母
func isFrench(r rune) bool {
	return unicode.Is(unicode.Latin, r)
}

// 是阿拉伯字母
func isArabic(r rune) bool {
	return unicode.Is(unicode.Arabic, r)
}

// 是希腊字母
func isGreek(r rune) bool {
	return unicode.Is(unicode.Greek, r)
}

// 是数字
func isN(r rune) bool {
	return unicode.Is(unicode.Number, r) ||
		unicode.Is(unicode.Digit, r) ||
		unicode.IsNumber(r)
}

// First 获取第一段满足条件的码点组成的字符串
func First(s string, f ...func(r rune) bool) string {
	pre := -1
ru:
	for i, r := range s {
		for _, v := range f {
			if v(r) {
				if pre < 0 {
					pre = i
				}
				continue ru
			}
		}
		if pre >= 0 {
			return s[pre:i]
		}
	}
	if pre >= 0 {
		return s[pre:]
	}
	return ``
}

// Mid 获取中间最长字符串，前缀后缀为空则忽略
//
//	pre 为前缀（不包含），suf 为后缀（不包含），str 为整个字符串
func Mid(pre, suf, str string) string {
	return mid(pre, suf, str, false)
}

// 找不到的字符串
const canNotFound = "\x00"

// MidMin 获取中间最短字符串，前缀后缀为空或找不到则忽略
//
//	pre 为前缀（不包含），suf 为后缀（不包含），str 为整个字符串
func MidMin(pre, suf, str string) string {
	return mid(
		cmp.Or(pre, canNotFound),
		cmp.Or(suf, canNotFound),
		str, true,
	)
}

// 获取中间字符串
//
//	pre 为前缀（不包含），suf 为后缀（不包含），str 为整个字符串
func mid(pre, suf, str string, isMin bool) string {
	var (
		low = func() int {
			// 截掉前缀及之前部分
			if i := func() int {
				if isMin {
					return strings.LastIndex(str, pre)
				}
				return strings.Index(str, pre)
			}(); i != -1 {
				return i + len(pre)
			}
			return 0
		}() // 下界
		up = func() int {
			// 截掉后缀及之后部分
			if i := func() int {
				if isMin {
					return strings.Index(str, suf)
				}
				return strings.LastIndex(str, suf)
			}(); i != -1 {
				return i
			}
			return len(str)
		}() // 上界
	)
	if low >= up {
		// 如果上界不高于下界，则返回空字符串
		return ``
	}
	return str[low:up]
}

// Similarity 计算两个字符串的相似程度
func Similarity(a, b string) float64 {
	return 0.5 * (Equal(a, b) + Common(a, b))
}

// Equal 计算两个字符串的未更改字符占比
func Equal(a, b string) float64 {
	return float64(equalN(diffmatchpatch.New().DiffMain(a, b, false))) /
		float64(max(utf8.RuneCountInString(a), utf8.RuneCountInString(b)))
}

// Common 计算两个字符串的公共字符占比
func Common(a, b string) float64 {
	return float64(commonN(a, b)) /
		float64(utf8.RuneCountInString(a)+utf8.RuneCountInString(b))
}

// 计算两个字符串中未更改字符的个数
func equalN(diffs []diffmatchpatch.Diff) (n int) {
	for _, d := range diffs {
		if d.Type == diffmatchpatch.DiffEqual {
			n += utf8.RuneCountInString(d.Text)
		}
	}
	return
}

// 计算两个字符串中公共字符的个数
func commonN(a, b string) (n int) {
	for _, r := range a {
		if strings.ContainsRune(b, r) {
			n++
		}
	}
	for _, r := range b {
		if strings.ContainsRune(a, r) {
			n++
		}
	}
	return
}

// SimilarityChinese 计算两个汉字字符串的相似程度
//
//	Deprecated: 结果不具备足够的参考意义
func SimilarityChinese(a, b string) float64 {
	const avg = `盒` // 汉字平均码点值
	var (
		ar = []rune(a)
		br = []rune(b)
	)
	switch ac, bc := len(ar), len(br); cmp.Compare(ac, bc) {
	case -1:
		ar = []rune(a + strings.Repeat(avg, bc-ac))
	case 1:
		br = []rune(b + strings.Repeat(avg, ac-bc))
	}
	var sum, s1, s2 float64
	for i := range ar {
		sum += float64(ar[i]) * float64(br[i])
		s1 += math.Pow(float64(ar[i]), 2)
		s2 += math.Pow(float64(br[i]), 2)
	}
	if s1 == 0 || s2 == 0 {
		return 0
	}
	return sum / math.Sqrt(s1*s2)
}

// Levenshtein 计算两个字符串之间的编辑距离
func Levenshtein(a, b string) int {
	var (
		ar   = []rune(a)
		br   = []rune(b)
		lenA = len(ar)
		lenB = len(br)
		dp   = make([][]int, lenA+1) // 创建二维切片
	)
	for i := range lenA + 1 {
		dp[i] = make([]int, lenB+1)
	}
	// 初始化
	for i := range lenA + 1 {
		dp[i][0] = i
	}
	for j := range lenB + 1 {
		dp[0][j] = j
	}
	// 动态规划计算编辑距离
	for i := range lenA {
		for j := range lenB {
			if ar[i] == br[j] {
				dp[i+1][j+1] = dp[i][j]
				continue
			}
			dp[i+1][j+1] = min(dp[i][j], dp[i][j+1], dp[i+1][j]) + 1
		}
	}
	return dp[lenA][lenB]
}

// AddNumSpace 在数字前后添加空格
func AddNumSpace(s string) string {
	b := new(strings.Builder)
	b.Grow(len(s) + 4)
	n := len(s)
	for i := range n {
		c := s[i]
		isDigit := c >= '0' && c <= '9'
		if isDigit && i > 0 && (s[i-1] < '0' || s[i-1] > '9') {
			b.WriteByte(' ')
		}
		b.WriteByte(c)
		if isDigit && i+1 < n && (s[i+1] < '0' || s[i+1] > '9') {
			b.WriteByte(' ')
		}
	}
	return b.String()
}
