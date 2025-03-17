package core

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode"
)

type str interface {
	~string | ~[]rune | ~[]byte
}

/*
CleanAll 清理字符串中全部不必要内容

lf 控制是否换行
*/
func CleanAll[T str](s T, lf bool) T {
	return T(
		strings.TrimSpace(
			strings.Map(func(r rune) rune {
				if remove := unicode.IsControl(r) || unicode.IsSpace(r);
				// 如果不换行，移除包括换行符在内的控制字符和空白字符
				(!lf && remove) ||
					// 如果换行，移除换行符以外的控制字符和空白字符
					(lf && remove && !strings.ContainsRune("\n\r", r)) ||
					// 移除可能引发排版和显示错误的字符
					strings.ContainsRune("\u061c\u200e\u200f\u202a\u202b\u202c\u202d\u202e\u2066\u2067\u2068\u2069\ufffd", r) {
					return -1
				}
				return r
			},
				string(s),
			),
		),
	)
}

// 获取 rune 对应的 UTF-8 字节数
func runeBytes(r rune) int {
	return len([]byte(string([]rune{r})))
}

// ClearRuneBytes 移除特定长度的 UTF-8 码点
func ClearRuneBytes[T str](s T, n ...int) T {
	return T(strings.Map(func(r rune) rune {
		if slices.Contains(n, runeBytes(r)) {
			return -1
		}
		return r
	}, string(s)))
}

// Compose 排版
func Compose(b *strings.Builder, s string) string {
	if b == nil {
		b = &strings.Builder{}
	}
	for line := range strings.Lines(strings.TrimSpace(s)) {
		if line == `` {
			continue
		}
		b.WriteString(`　　`)
		fmt.Fprintln(b, strings.TrimSpace(line))
	}
	return b.String()
}

// First 获取第一段满足条件的码点组成的字符串
func First[T str](s T, f ...func(r rune) bool) T {
	var (
		pre, count int
		ok         bool
	)
ru:
	for _, r := range string(s) {
		for _, v := range f {
			if v(r) {
				ok = true
				count++
				continue ru
			}
		}
		if ok {
			return T(string([]rune(string(s))[pre : pre+count]))
		}
		pre++
	}
	return T(string([]rune(string(s))[pre : pre+count]))
}

// 是中文
func isChinese(ch rune) bool {
	return unicode.Is(unicode.Han, ch)
}

// 是假名
func isJapanese(ch rune) bool {
	return unicode.Is(unicode.Hiragana, ch) ||
		unicode.Is(unicode.Katakana, ch)
}

// 是谚文
func isKorean(ch rune) bool {
	return unicode.Is(unicode.Hangul, ch)
}

// 是西里尔字母
func isRussian(ch rune) bool {
	return unicode.Is(unicode.Cyrillic, ch)
}

// 是拉丁字母
func isFrench(ch rune) bool {
	return unicode.Is(unicode.Latin, ch)
}

// 是阿拉伯字母
func isArabic(ch rune) bool {
	return unicode.Is(unicode.Arabic, ch)
}

// 是希腊字母
func isGreek(ch rune) bool {
	return unicode.Is(unicode.Greek, ch)
}

// 是数字
func isN(ch rune) bool {
	return unicode.Is(unicode.Number, ch) ||
		unicode.Is(unicode.Digit, ch) ||
		unicode.IsNumber(ch)
}

// FirstHan 获取第一段中文字符码点组成的字符串
func FirstHan[T str](s T) T {
	return First(s, isChinese)
}

// FirstText 获取第一段常用文字（不包括标点）组成的字符串
func FirstText[T str](s T) T {
	return First(s,
		isChinese, isJapanese, isKorean,
		isRussian, isFrench, isArabic, isGreek,
		isN, unicode.IsLetter)
}

/*
MidText 获取中间最长字符串，前缀后缀为空则忽略

pre 为前缀（不包含），suf 为后缀（不包含），str 为整个字符串
*/
func MidText(pre, suf, str string) string {
	return midText(pre, suf, str, false)
}

/*
MidTextMin 获取中间最短字符串，前缀后缀找不到则忽略（建议使用 "\u0000"）

pre 为前缀（不包含），suf 为后缀（不包含），str 为整个字符串
*/
func MidTextMin(pre, suf, str string) string {
	return midText(pre, suf, str, true)
}

/*
获取中间字符串

pre 为前缀（不包含），suf 为后缀（不包含），str 为整个字符串
*/
func midText(pre, suf, str string, min bool) string {
	var (
		low = func() int {
			// 截掉前缀及之前部分
			if i := func() int {
				if min {
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
				if min {
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

// SimilarityChinese 返回两个汉字字符串的相似程度
func SimilarityChinese(x, y string) float64 {
	const avg = `盒` // 汉字平均码点值
	switch xc, yc := len([]rune(x)), len([]rune(y)); cmp.Compare(xc, yc) {
	case -1:
		x += strings.Repeat(avg, yc-xc)
	case 1:
		y += strings.Repeat(avg, xc-yc)
	}
	var sum, s1, s2 float64
	for i, r := range []rune(x) {
		sum += float64(r) * float64([]rune(y)[i])
		s1 += math.Pow(float64(r), 2)
		s2 += math.Pow(float64([]rune(y)[i]), 2)
	}
	if s1 == 0 || s2 == 0 {
		return 0
	}
	return sum / math.Sqrt(s1*s2)
}

// Levenshtein 计算两个字符串之间的编辑距离
func Levenshtein(a, b string) int {
	var (
		lenA = len([]rune(a))
		lenB = len([]rune(b))
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
			if []rune(a)[i] == []rune(b)[j] {
				dp[i+1][j+1] = dp[i][j]
				continue
			}
			dp[i+1][j+1] = min(dp[i][j], dp[i][j+1], dp[i+1][j]) + 1
		}
	}
	return dp[lenA][lenB]
}
