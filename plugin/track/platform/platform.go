package platform

import (
	"time"
)

// 所有小说平台
var platforms []Platform

// Register 注册小说平台
func Register(p Platform) {
	platforms = append(platforms, p)
}

// Get 获取小说平台
func Get(platform string) (Platform, error) {
	for _, p := range platforms {
		if p.String() == platform {
			return p, nil
		}
	}
	return nil, NotSupported(platform)
}

// NotSupported *NotSupportedError 的构造函数，不支持的平台
func NotSupported(platform string) *NotSupportedError {
	return &NotSupportedError{
		platform: platform,
	}
}

// Error 实现 error
func (e *NotSupportedError) Error() string {
	return e.platform + `不是受支持的小说平台喵！`
}

// ParseTime 解析时间
func ParseTime(p Platform, str string) (time.Time, error) {
	return time.Parse(p.Layout(), str)
}
