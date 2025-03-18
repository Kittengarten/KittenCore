package io

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core/http"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"

	"gopkg.in/yaml.v3"
)

type (
	// Path 文件路径
	Path interface {
		FileName() string
		Delete() error
		Load(write bool) (f *os.File, err error)
		ReadBytes() (*bytes.Buffer, error)
		ReadString() (string, error)
		WriteBytes(b []byte) error
		WriteString(s string) error
		Len() int
		Size() (int64, error)
		TryMakeDir() error
		String() string
		DownloadImage(url string) (int64, error)
		InitFile(def string) error
		Get(def FilePath) FilePath
		GetString(def string) string
	}

	// FilePath 是一个表示文件路径的字符串
	FilePath string
)

const (
	Empty = `[]` // Empty YAML 空数组（slice）
	Blank = `{}` // Blank YAML 空集合（map）
)

func init() {
	fs.ErrInvalid = errors.New(`无效的参数喵！`)
	fs.ErrPermission = errors.New(`没有权限喵！`)
	fs.ErrExist = errors.New(`文件已存在喵！`)
	fs.ErrNotExist = errors.New(`文件不存在喵！`)
	fs.ErrClosed = errors.New(`文件已关闭喵！`)
	os.ErrInvalid = fs.ErrInvalid
	os.ErrPermission = fs.ErrPermission
	os.ErrExist = fs.ErrExist
	os.ErrNotExist = fs.ErrNotExist
	os.ErrClosed = fs.ErrClosed
}

// 加载 YAML 配置文件，def 为默认值（加载不到的时候会尝试初始化）
func Load[T any](p Path, def string) (c T, err error) {
	if err = p.InitFile(def); err != nil {
		err = fmt.Errorf(`初始化 %s 时失败喵！%w`, p, err)
		return
	}
	f, err := p.Load(false)
	if err != nil {
		err = fmt.Errorf(`加载 %s 时失败喵！%w`, p, err)
		return
	}
	defer f.Close()
	err = yaml.NewDecoder(f).Decode(&c)
	return
}

// 保存 YAML 配置文件
func Save[T any](p Path, c T) error {
	f, err := p.Load(true)
	if err != nil {
		return err
	}
	defer f.Close()
	err = yaml.NewEncoder(f).Encode(c)
	return err
}

// NewPath 文件路径构建
func NewPath[T ~string](elem ...T) FilePath {
	if len(elem) == 0 {
		return ``
	}
	const symbol, symbol_ = `://`, `%symbol%`
	elem[0] = T(strings.Replace(string(elem[0]), symbol, symbol_, 1))
	return FilePath(
		strings.Replace(
			filepath.Join(
				utils.ConvertSlice(
					elem,
					func(e T) string {
						return string(e)
					},
				)...,
			),
			symbol_,
			symbol,
			1,
		),
	)
}

// FileName 文件名
func (p FilePath) FileName() string {
	return filepath.Base(p.String())
}

// Delete 删除文件
func (p FilePath) Delete() error {
	return os.Remove(p.String())
}

// 载入文件以供操作，当 write 为 false 时只读
func (p FilePath) Load(write bool) (f *os.File, err error) {
	// 检查其父文件夹是否存在，不存在则创建
	if err := p.TryMakeDir(); err != nil {
		err = fmt.Errorf(`创建 %s 的父文件夹失败喵！%w`, filepath.Dir(p.String()), err)
		return nil, err
	}
	if !write {
		// 只读，打开文件
		f, err = os.Open(p.String())
		if err == nil {
			return f, nil
		}
	}
	// 需要写入或不存在，尝试创建文件
	return os.Create(p.String())
}

// ReadBytes 从文件读取字节切片
func (p FilePath) ReadBytes() (*bytes.Buffer, error) {
	f, err := p.Load(false)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var b bytes.Buffer
	_, err = io.Copy(&b, f)
	return &b, err
}

// ReadString 从文件读取字符串
func (p FilePath) ReadString() (string, error) {
	f, err := p.Load(false)
	if err != nil {
		return ``, err
	}
	defer f.Close()
	var s strings.Builder
	_, err = io.Copy(&s, f)
	return s.String(), err
}

/*
WriteBytes 向文件写入字节切片（会从头覆盖文件）

如文件不存在会尝试新建
*/
func (p FilePath) WriteBytes(b []byte) error {
	f, err := p.Load(true)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(b)
	return err
}

/*
WriteString 向文件写入字符串

如文件不存在会尝试新建
*/
func (p FilePath) WriteString(s string) error {
	f, err := p.Load(true)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprint(f, s)
	return err
}

// Len 获取路径长度
func (p FilePath) Len() int {
	return len(p)
}

// Size 获取文件大小
func (p FilePath) Size() (int64, error) {
	info, err := os.Stat(p.String())
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// TryMakeDir 检查其父文件夹是否存在，不存在则创建
func (p FilePath) TryMakeDir() error {
	return os.MkdirAll(filepath.Dir(p.String()), 0o755)
}

// Exists 判断文件或文件夹是否存在
func (p FilePath) Exists() bool {
	_, err := os.Stat(p.String())
	return err == nil || os.IsExist(err)
}

// IsDir 判断路径是否文件夹
func (p FilePath) IsDir() (bool, error) {
	info, err := os.Stat(p.String())
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// String 实现 fmt.Stringer，返回路径规范化后的字符串表示
func (p FilePath) String() string {
	path := string(p)
	if strings.Contains(path, `://`) {
		// Windows → 类 Unix 跨平台处理
		return strings.ReplaceAll(path, `\`, `/`)
	}
	return filepath.Clean(path)
}

// LoadPath 加载文件中保存的相对路径或绝对路径
func (p FilePath) LoadPath() (FilePath, error) {
	s, err := p.ReadString()
	if err != nil {
		return p, err
	}
	if s = strings.TrimSpace(s); filepath.IsAbs(s) {
		return NewPath(`file://`, s), nil
	}
	return NewPath(s), nil
}

// DownloadImage 从 url 下载图片到 path
func (p FilePath) DownloadImage(url string) (int64, error) {
	// 获取 HTTP 响应体，失败则返回
	b, err := http.GET(url)
	if err != nil {
		return 0, err
	}
	f, err := p.Load(true)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.ReadFrom(b)
}

// InitFile 初始化文本文件，要求传入路径事先规范化过
func (p FilePath) InitFile(def string) error {
	// 如果文件存在，直接返回，以免覆盖文件
	if p.Exists() {
		return nil
	}
	// 如果文件不存在，初始化该文件
	return p.WriteString(def)
}

// Get 从文件获取路径，def 为默认值（加载不到的时候会尝试初始化）
func (p FilePath) Get(def FilePath) FilePath {
	return NewPath(p.GetString(def.String()))
}

// GetString 从文件获取字符串，def 为默认值（加载不到的时候会尝试初始化）
func (p FilePath) GetString(def string) string {
	if err := p.InitFile(def); err != nil {
		slog.Error(`初始化文件失败了喵！`, slog.Any(`路径`, p), slog.Any(`错误`, err))
		return def
	}
	s, err := p.ReadString()
	if err != nil {
		slog.Error(`打开文件失败了喵！`, slog.Any(`路径`, p), slog.Any(`错误`, err))
		return def
	}
	return s
}

// ProcessPath 获取程序运行的绝对路径
func ProcessPath() (FilePath, error) {
	ex, err := os.Executable()
	if err != nil {
		return ``, err
	}
	return NewPath(filepath.Dir(ex)), nil
}

// HandleFileName 处理文件名中不支持的字符
func HandleFileName(filename string) string {
	return strings.NewReplacer(
		`\`, `_`,
		`/`, `_`,
		`:`, `_`,
		`*`, `_`,
		`?`, `_`,
		`<`, `_`,
		`>`, `_`,
		`|`, `_`,
	).Replace(filename)
}
