package platform

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/plugin/track/chapter"
	"github.com/Kittengarten/KittenCore/plugin/track/novel"
	"golang.org/x/net/html"
)

// 所有小说平台
var platforms []Platform

// Register 注册小说平台
func Register(p Platform) {
	platforms = append(platforms, p)
}

func init() {
	chapter.GetChapterSource = func(platform string) (chapter.Source, error) {
		return Get(platform)
	}
	novel.GetChapterIDSource = func(platform string) (novel.ChapterIDSource, error) {
		return Get(platform)
	}
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

// NotSupported 不支持的平台
func NotSupported(platform string) error {
	return fmt.Errorf(`%s不是受支持的小说平台喵！%w`,
		platform, errors.ErrUnsupported)
}

// ParseTime 解析时间
func ParseTime(p Platform, str string) (time.Time, error) {
	return time.ParseInLocation(p.Layout(), str, times.Location)
}

// Doc 获取 HTML
func Doc(ctx context.Context, urlStr string, cache bool) (*html.Node, error) {
	if cache {
		return shttp.UpdateURLWithContext(ctx, urlStr)
	}
	return shttp.LoadURLWithContext(ctx, urlStr)
}
