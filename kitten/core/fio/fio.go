// Package fio 处理文件 IO
package fio

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/str"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"

	"gopkg.in/yaml.v3"
)

type (
	// Path 是一个表示文件路径的字符串
	Path string
	// PathMutex Path 的互斥锁版本
	PathMutex struct {
		Path
		*sync.Mutex
	}
	// PathRWMutex Path 的读写锁版本
	PathRWMutex struct {
		Path
		*sync.RWMutex
	}
)

const (
	Empty = `[]` // Empty YAML 空数组（slice）
	Blank = `{}` // Blank YAML 空集合（map）
)

// WithMutex 附加互斥锁
func (p Path) WithMutex() PathMutex {
	return PathMutex{Path: p, Mutex: new(sync.Mutex)}
}

// WithRWMutex 附加读写锁
func (p Path) WithRWMutex() PathRWMutex {
	return PathRWMutex{Path: p, RWMutex: new(sync.RWMutex)}
}

// Load 加载 YAML 配置文件，def 为默认值（加载不到的时候会尝试初始化）
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

// Save 保存 YAML 配置文件
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
func NewPath[T ~string](elem ...T) Path {
	if len(elem) == 0 {
		return ``
	}
	const symbol, symbol_ = `://`, `%symbol%`
	elem[0] = T(strings.Replace(string(elem[0]), symbol, symbol_, 1))
	return Path(
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

// NoDuplicate 生成不重复的文件或文件夹名
func NoDuplicate(names []string, name string) string {
	for slices.Contains(names, name) {
		name = str.Rename(name)
	}
	return name
}

// NoDuplicate 生成不重复的文件或文件夹名
func (p Path) NoDuplicate() (Path, error) {
	names, err := p.Names()
	return NewPath(p.Dir(), NoDuplicate(names, p.Name())), err
}

/*
Names 获取文件夹下所有文件（夹）名

如果是文件，获取父文件夹下的所有文件（夹）名
*/
func (p Path) Names() ([]string, error) {
	isDir, err := p.IsDir()
	if err != nil {
		return nil, err
	}
	dir := func() string {
		if isDir {
			return p.String()
		}
		return p.Dir()
	}()
	// 读取目录内容
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	return utils.ConvertSlice(
		entries, func(e fs.DirEntry) string {
			return e.Name()
		},
	), nil
}

// Dir 父文件夹名
func (p Path) Dir() string {
	return filepath.Dir(string(p))
}

// Name 文件或文件夹名
func (p Path) Name() string {
	return filepath.Base(string(p))
}

// Split 分割为文件夹和文件名
func (p Path) Split() (dir, file string) {
	return filepath.Split(string(p))
}

// Delete 删除文件
func (p Path) Delete() error {
	return os.Remove(p.String())
}

// 载入文件以供操作，当 write 为 false 时只读
func (p Path) Load(write bool) (f *os.File, err error) {
	// 检查其父文件夹是否存在，不存在则创建
	if err := p.TryMakeDir(); err != nil {
		err = fmt.Errorf(`创建 %s 失败喵！%w`, p.Dir(), err)
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
func (p Path) ReadBytes() (*bytes.Buffer, error) {
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
func (p Path) ReadString() (string, error) {
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
func (p Path) WriteBytes(b []byte) error {
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
func (p Path) WriteString(s string) error {
	f, err := p.Load(true)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprint(f, s)
	return err
}

// Len 获取路径长度
func (p Path) Len() int {
	return len(p)
}

// Size 获取文件大小
func (p Path) Size() (int64, error) {
	info, err := os.Stat(p.String())
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// TryMakeDir 检查其父文件夹是否存在，不存在则创建
func (p Path) TryMakeDir() error {
	return os.MkdirAll(p.Dir(), 0o750)
}

// Exists 判断文件或文件夹是否存在
func (p Path) Exists() bool {
	_, err := os.Stat(p.String())
	return err == nil || os.IsExist(err)
}

// IsDir 判断路径是否文件夹
func (p Path) IsDir() (bool, error) {
	info, err := os.Stat(p.String())
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// String 实现 fmt.Stringer，返回路径规范化后的字符串表示
func (p Path) String() string {
	const sep = `://`
	path := string(p)
	if before, after, found := strings.Cut(path, sep); found {
		// Windows → 类 Unix 跨平台处理
		return before + sep + strings.ReplaceAll(after, `\`, `/`)
	}
	return filepath.Clean(path)
}

// LoadPath 加载文件中保存的相对路径或绝对路径
func (p Path) LoadPath() (Path, error) {
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
func (p Path) DownloadImage(url string) (int64, error) {
	// 获取 HTTP 响应体，失败则返回
	b, err := shttp.GET(url)
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
func (p Path) InitFile(def string) error {
	// 如果文件存在，直接返回，以免覆盖文件
	if p.Exists() {
		return nil
	}
	// 如果文件不存在，初始化该文件
	return p.WriteString(def)
}

// Get 从文件获取路径，def 为默认值（加载不到的时候会尝试初始化）
func (p Path) Get(def Path) Path {
	return NewPath(p.GetString(string(def)))
}

// GetString 从文件获取字符串，def 为默认值（加载不到的时候会尝试初始化）
func (p Path) GetString(def string) string {
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

// Copy 复制文件
func (p Path) Copy(src Path) (size int64, err error) {
	// 打开源文件
	source, err := src.Load(false)
	if err != nil {
		return
	}
	// 打开目标文件
	destination, err := p.Load(true)
	if err != nil {
		return
	}
	// 保存图片
	size, err = io.Copy(destination, source)
	if err != nil {
		return
	}
	return size, errors.Join(source.Close(), destination.Close())
}

// SHA512 ...
type SHA512 [sha512.Size]byte

// String 实现 fmt.Stringer，返回十六进制哈希值（128 位数字）
func (s SHA512) String() string {
	return hex.EncodeToString(s[:])
}

// SHA512 获取文件 SHA512 哈希值
func (p Path) SHA512() (SHA512, error) {
	if !p.Exists() {
		return SHA512{}, os.ErrNotExist
	}
	b, err := p.ReadBytes()
	if err != nil {
		return SHA512{}, err
	}
	return SHA512(sha512.Sum512(b.Bytes())), nil
}

// ProcessPath 获取程序运行的绝对路径
func ProcessPath() (Path, error) {
	ex, err := os.Executable()
	if err != nil {
		return ``, err
	}
	return NewPath(filepath.Dir(ex)), nil
}
