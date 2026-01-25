// Package mio 用于处理涉及消息处理的文件 IO
package mio

import (
	"html"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// Path 是一个表示文件路径的结构体
type Path struct {
	fio.Path
}

// New MsgPath 的构造函数
func New(p fio.Path) Path {
	return Path{p}
}

// NewPath MsgPath 的构造函数
func NewPath[T ~string](elem ...T) Path {
	return Path{fio.NewPath(elem...)}
}

// Image 从图片的相对 | 绝对路径（文件夹），
//
// 或相对 | 绝对路径文件中保存的相对 | 绝对路径，
//
// 或网络路径中加载图片
func (p Path) Image(name fio.Path) (message.Segment, error) {
	var (
		pre = func() string {
			switch runtime.GOOS {
			case `linux`:
				return `file://`
			default:
				return ``
			}
		}()
		fn = `[` + name.Name() + `]`
	)
	if filepath.IsAbs(name.String()) {
		// 传入的是绝对路径
		return New(name).Rand(pre, fn)
	}
	if strings.Contains(name.String(), `://`) {
		// 传入的是网络路径
		return message.Image(name.String(), fn), nil
	}
	// 传入的是相对路径
	path := string(p.Path)
	if strings.Contains(path, `://`) {
		// 请求的是网络路径
		return message.Image(fio.NewPath(p.Path, name).String(), fn), nil
	}
	if filepath.IsAbs(path) {
		// 请求的是绝对路径
		isDir, err := p.IsDir()
		if err != nil {
			return message.Segment{}, err
		}
		if isDir {
			// 请求的是文件夹
			return NewPath(path, name.String()).Rand(pre, fn)
		}
		// 请求的是文件
		np, err := p.LoadPath()
		if err != nil {
			return message.Segment{}, err
		}
		return NewPath(np, name).Rand(pre, fn)
	}
	// 请求的是相对路径
	if isDir, err := p.IsDir(); isDir {
		if err != nil {
			return message.Segment{}, err
		}
		// 请求的是文件夹，需要转换为绝对路径
		abs, err := fio.ProcessPath()
		if err != nil {
			return message.Segment{}, err
		}
		return NewPath(abs, p.Path, name).Rand(pre, fn)
	}
	// 请求的是文件
	np, err := p.LoadPath()
	if err != nil {
		return message.Segment{}, err
	}
	return NewPath(np, name).Rand(pre, fn)
}

// Rand 从绝对路径中随机抽取一张图片
func (p Path) Rand(pre, fn string) (message.Segment, error) {
	isDir, err := p.IsDir()
	if err != nil {
		return message.Segment{}, err
	}
	if !isDir {
		// 传入的是文件
		return message.Image(pre+p.String(), fn), nil
	}
	// 传入的是文件夹，从文件夹中随机抽取一张图片
	img, err := p.Path.Rand()
	if err != nil {
		return message.Segment{}, err
	}
	return message.Image(pre+img.String(), fn), nil
}

// GetImagePath 获取图片路径
func GetImagePath(e message.Segment) fio.Path {
	return fio.Path(e.Data[`file`])
}

// GetImageURL 获取图片 URL
func GetImageURL(e message.Segment) string {
	return html.UnescapeString(e.Data[`url`])
}
