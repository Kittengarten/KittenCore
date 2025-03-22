// Package mio 用于处理涉及消息处理的文件 IO
package mio

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core/io"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// Path 是一个表示文件路径的结构体
type Path struct {
	io.Path
}

// New MsgPath 的构造函数
func New(p io.Path) Path {
	return Path{p}
}

/*
Image 从图片的相对 | 绝对路径（文件夹），

或相对 | 绝对路径文件中保存的相对 | 绝对路径，

或网络路径中加载图片
*/
func (p Path) Image(name io.Path) (message.Segment, error) {
	var (
		pre = func() string {
			switch runtime.GOOS {
			case `linux`:
				return `file://`
			default:
				return ``
			}
		}()
		fn = `[` + name.FileName() + `]`
	)
	if filepath.IsAbs(name.String()) {
		// 传入的是绝对路径
		if isDir, err := name.IsDir(); err == nil && !isDir {
			// 传入的是文件
			return message.Image(pre+name.String(), fn), nil
		}
	}
	if strings.Contains(name.String(), `://`) {
		// 传入的是网络路径
		return message.Image(name.String(), fn), nil
	}
	// 传入的是相对路径
	path := string(p.Path)
	if strings.Contains(path, `://`) {
		// 请求的是网络路径
		return message.Image(io.NewPath(p.String(), name.String()).String(), fn), nil
	}
	if filepath.IsAbs(path) {
		// 请求的是绝对路径
		isDir, err := p.IsDir()
		if err != nil {
			return message.Segment{}, err
		}
		if isDir {
			// 请求的是文件夹
			return message.Image(pre+io.NewPath(path, name.String()).String(), fn), nil
		}
		// 请求的是文件
		np, err := p.LoadPath()
		return message.Image(pre+io.NewPath(np, name).String(), fn), err
	}
	// 请求的是相对路径
	if isDir, err := p.IsDir(); isDir {
		if err != nil {
			return message.Segment{}, err
		}
		// 请求的是文件夹，需要转换为绝对路径
		abs, err := io.ProcessPath()
		return message.Image(io.NewPath(pre, abs.String(), p.String(), name.String()).String(), fn), err
	}
	// 请求的是文件
	np, err := p.LoadPath()
	return message.Image(io.NewPath(np, name).String(), fn), err
}
