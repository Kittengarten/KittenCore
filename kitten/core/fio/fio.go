// Package fio 处理文件 IO
package fio

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/Kittengarten/KittenCore/kitten/core/str"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"

	"github.com/goccy/go-yaml"
)

type (
	// Path 是一个表示文件路径的字符串
	Path string
	// PathMutex Path 的互斥锁版本
	PathMutex struct {
		Path        // 路径
		*sync.Mutex // 互斥锁
	}
	// PathRWMutex Path 的读写锁版本
	PathRWMutex struct {
		Path          // 路径
		*sync.RWMutex // 读写锁
	}
)

const (
	Empty = `[]` // Empty YAML 空数组（slice）
	Blank = `{}` // Blank YAML 空映射（map）
)

// WithMutex 附加互斥锁
func (p Path) WithMutex() PathMutex {
	return PathMutex{Path: p, Mutex: new(sync.Mutex)}
}

// WithRWMutex 附加读写锁
func (p Path) WithRWMutex() PathRWMutex {
	return PathRWMutex{Path: p, RWMutex: new(sync.RWMutex)}
}

// 记录错误，s 为描述，p 为路径
func logError(s string, path Path, err error) {
	p := slog.Any(`路径`, path)
	if err == nil {
		slog.Error(s, p)
		return
	}
	slog.Error(s, p, slog.Any(`错误`, err))
}

// Load 加载 YAML 配置文件，def 为默认值（加载不到的时候会尝试初始化）
//
//	即使 err == nil，加载的映射或切片也可能为 nil，需要判断
func Load[T any, S str.Str](p Path, def S, opts ...yaml.DecodeOption) (c T, err error) {
	return LoadWithContext[T](context.Background(), p, def, opts...)
}

// LoadWithContext 加载 YAML 配置文件（带上下文），def 为默认值（加载不到的时候会尝试初始化）
//
//	即使 err == nil，加载的映射或切片也可能为 nil，需要判断
func LoadWithContext[T any, S str.Str](ctx context.Context, p Path, def S, opts ...yaml.DecodeOption) (c T, err error) {
	if err = p.InitFileText(string(def)); err != nil {
		err = fmt.Errorf(`初始化 %s 时失败喵！%w`, p, err)
		return
	}
	f, err := p.Load(false)
	return c, errors.Join(err, yaml.NewDecoder(f, opts...).DecodeContext(ctx, &c), f.Close())
}

// Save 保存 YAML 配置文件
func Save[T any](p Path, c T, opts ...yaml.EncodeOption) error {
	return SaveWithContext(context.Background(), p, c, opts...)
}

// SaveWithContext 保存 YAML 配置文件（带上下文）
func SaveWithContext[T any](ctx context.Context, p Path, c T, opts ...yaml.EncodeOption) error {
	f, err := p.Load(true)
	return errors.Join(err, yaml.NewEncoder(f, opts...).EncodeContext(ctx, c), f.Close())
}

// NoDuplicate 生成不重复的文件或文件夹名
func (p Path) NoDuplicate() (Path, error) {
	names, err := p.Names()
	return NewPath(p.Dir(), NoDuplicate(names, p.Name())), err
}

// Name 文件或文件夹名
func (p Path) Name() string {
	return filepath.Base(string(p))
}

// NoDuplicate 生成不重复的文件或文件夹名
func NoDuplicate(names []string, name string) string {
	for slices.Contains(names, name) {
		name = str.Rename(name)
	}
	return name
}

// Rand 从绝对路径（文件夹）随机抽取一个文件
func (p Path) Rand() (Path, error) {
	names, err := p.Names()
	if err != nil {
		return p, err
	}
	//nolint:gosec
	return NewPath(p.String(), names[rand.N(len(names))]), nil
}

// Names 获取文件夹下所有文件（夹）名
//
//	如果是文件，获取父文件夹下的所有文件（夹）名
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

// Split 分割为文件夹和文件名
func (p Path) Split() (dir, file string) {
	return filepath.Split(string(p))
}

// Delete 删除文件
func (p Path) Delete() error {
	return os.Remove(p.String())
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

// IsDir 判断路径是否文件夹
func (p Path) IsDir() (bool, error) {
	info, err := os.Stat(p.String())
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
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

// Get 从文件获取路径，def 为默认值（加载不到的时候会尝试初始化）
func (p Path) Get(def Path) Path {
	return NewPath(p.GetString(string(def)))
}

// GetString 从文件获取字符串，def 为默认值（加载不到的时候会尝试初始化）
func (p Path) GetString(def string) string {
	if err := p.InitFileText(def); err != nil {
		logError(`初始化文件失败了喵！`, p, err)
		return def
	}
	s, err := p.ReadString()
	if err != nil {
		logError(`打开文件失败了喵！`, p, err)
		return def
	}
	return s
}

// InitFileText 初始化文本文件，要求传入路径事先规范化过
func (p Path) InitFileText(def string) error {
	// 如果文件存在，直接返回，以免覆盖文件
	if p.Exists() {
		return nil
	}
	// 如果文件不存在，初始化该文件
	return p.WriteString(def)
}

// InitFile 初始化文件，要求传入路径事先规范化过
func (p Path) InitFile(def ...byte) error {
	// 如果文件存在，直接返回，以免覆盖文件
	if p.Exists() {
		return nil
	}
	// 如果文件不存在，初始化该文件
	return p.WriteBytes(def)
}

// WriteString 向文件写入字符串
//
//	如文件不存在会尝试新建
func (p Path) WriteString(s string) error {
	f, err := p.Load(true)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(f, s)
	return errors.Join(err, f.Close())
}

// WriteBytes 向文件写入字节切片（会从头覆盖文件）
//
//	如文件不存在会尝试新建
func (p Path) WriteBytes(b []byte) error {
	f, err := p.Load(true)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return errors.Join(err, f.Close())
}

// ReadString 从文件一次性读取全部内容，返回字符串
func (p Path) ReadString() (string, error) {
	f, err := p.Load(false)
	if err != nil {
		return ``, err
	}
	info, err := f.Stat()
	if err != nil {
		return ``, errors.Join(err, f.Close())
	}
	s := new(strings.Builder)
	s.Grow(int(info.Size()))
	_, err = io.Copy(s, f)
	return s.String(), errors.Join(err, f.Close())
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
		return size, source.Close()
	}
	// 完成复制
	size, err = io.Copy(destination, source)
	return size, errors.Join(err, source.Close(), destination.Close())
}

// Hash 获取文件哈希值
func (p Path) Hash(h hash.Hash) ([]byte, error) {
	if !p.Exists() {
		return nil, os.ErrNotExist
	}
	f, err := p.Load(false)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(h, f)
	if err != nil {
		return nil, errors.Join(err, f.Close())
	}
	return h.Sum(nil), f.Close()
}

// HashStr 获取文件十六进制哈希值
func (p Path) HashStr(h hash.Hash) (string, error) {
	hash, err := p.Hash(h)
	return hex.EncodeToString(hash), err
}

// Exists 判断文件或文件夹是否存在
func (p Path) Exists() bool {
	_, err := os.Stat(p.String())
	return err == nil || os.IsExist(err)
}

// ReadBytes 从文件一次性读取全部内容，返回字节切片
func (p Path) ReadBytes() ([]byte, error) {
	f, err := p.Load(false)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		return nil, errors.Join(err, f.Close())
	}
	b := make([]byte, int(info.Size()))
	_, err = io.ReadFull(f, b)
	return b, errors.Join(err, f.Close())
}

// 载入文件以供操作，当 write 为 false 时只读
func (p Path) Load(write bool) (f *os.File, err error) {
	if !write {
		// 只读，打开文件
		f, err = os.Open(p.String())
		if err == nil {
			return f, nil
		}
	}
	// 需要写入或不存在，尝试创建文件
	// 检查其父文件夹是否存在，不存在则创建
	if err := p.TryMakeDir(); err != nil {
		err = fmt.Errorf(`创建 %s 失败喵！%w`, p.Dir(), err)
		return nil, err
	}
	return os.Create(p.String())
}

// TryMakeDir 检查其父文件夹是否存在，不存在则创建
func (p Path) TryMakeDir() error {
	return os.MkdirAll(p.Dir(), 0o750)
}

// Dir 父文件夹名
func (p Path) Dir() string {
	return filepath.Dir(string(p))
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

// ProcessPath 获取程序运行的绝对路径
func ProcessPath() (Path, error) {
	ex, err := os.Executable()
	return NewPath(filepath.Dir(ex)), err
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
