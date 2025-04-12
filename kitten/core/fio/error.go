package fio

import (
	"errors"
	"io/fs"
	"os"
)

var (
	// ErrInvalid 无效的参数喵！
	ErrInvalid = errors.New(`无效的参数喵！`)
	// ErrPermission 没有权限喵！
	ErrPermission = errors.New(`没有权限喵！`)
	// ErrExist 文件已存在喵！
	ErrExist = errors.New(`文件已存在喵！`)
	// ErrNotExist 文件不存在喵！
	ErrNotExist = errors.New(`文件不存在喵！`)
	// ErrClosed 文件已关闭喵！
	ErrClosed = errors.New(`文件已关闭喵！`)
)

func init() {
	fs.ErrInvalid = ErrInvalid
	fs.ErrPermission = ErrPermission
	fs.ErrExist = ErrExist
	fs.ErrNotExist = ErrNotExist
	fs.ErrClosed = ErrClosed
	os.ErrInvalid = ErrInvalid
	os.ErrPermission = ErrPermission
	os.ErrExist = ErrExist
	os.ErrNotExist = ErrNotExist
	os.ErrClosed = ErrClosed
}
