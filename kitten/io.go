package kitten

import (
	"os"
	"path/filepath"

	"github.com/wdvxdr1123/ZeroBot/message"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// FilePath 文件路径构建
func FilePath(elem ...Path) Path {
	s := make([]string, len(elem))
	for k := range elem {
		s[k] = elem[k].String()
	}
	return Path(filepath.Join([]string(s)...))
}

// Read 文件读取
func (path Path) Read() (data []byte) {
	data, err := os.ReadFile(path.String())
	if !Check(err) {
		zap.S().Errorf("打开文件 %s 失败了喵！\n%v", path, err)
	}
	return
}

/*
Write 文件写入

如文件不存在会尝试新建
*/
func (path Path) Write(data []byte) {
	if e, err := path.Exists(); !e {
		// 如果文件或文件夹不存在，或不确定是否存在
		if Check(err) {
			// 如果文件不存在，新建该文件所在的文件夹；如果文件夹不存在，新建该文件夹本身
			if !Check(os.MkdirAll(filepath.Dir(path.String()), 0755)) {
				zap.S().Errorf("新建 %s 失败喵！\n%v", path, err)
			}
		} else {
			// 文件或文件夹不确定是否存在
			zap.S().Errorf("写入时不确定 %s 存在喵！\n%v", path, err)
		}
	} else {
		if err = os.WriteFile(path.String(), data, 0666); !Check(err) {
			zap.S().Errorf("写入文件 %s 失败了喵！\n%v", path, err)
		}
	}
}

/*
Exists 判断文件是否存在

不确定存在的情况下报错
*/
func (path Path) Exists() (bool, error) {
	_, err := os.Stat(path.String())
	if Check(err) {
		// 当 err 为空，文件或文件夹存在
		return true, nil
	}
	if os.IsNotExist(err) {
		// os.IsNotExist(err)为 true，文件或文件夹不存在
		return false, nil
	}
	// 其它类型，不确定是否存在
	return false, err
}

// （私有）判断路径是否文件夹
func (path Path) isDir() bool {
	if s, err := os.Stat(path.String()); Check(err) {
		return s.IsDir()
	} else {
		zap.S().Errorf("识别 %s 失败了喵！\n%v", path, err)
	}
	return false
}

// LoadPath 加载文件中保存的相对路径或绝对路径
func (path Path) LoadPath() Path {
	data, err := os.ReadFile(path.String())
	if !Check(err) {
		zap.S().Errorf("打开文件 %s 失败了喵！\n%v", path, err)
	}
	if filepath.IsAbs(string(data)) {
		return Path(`file://`) + FilePath(Path(data))
	}
	return FilePath(Path(data))
}

// GetImage 从图片的相对/绝对路径，或相对/绝对路径文件中保存的相对/绝对路径加载图片
func (path Path) GetImage(name Path) message.MessageSegment {
	if filepath.IsAbs(path.String()) {
		if path.isDir() {
			return message.Image(`file://` + FilePath(path, name).String())
		}
		return message.Image(`file://` + FilePath(path.LoadPath(), name).String())
	}
	if path.isDir() {
		return message.Image(FilePath(path, name).String())
	}
	return message.Image(FilePath(path.LoadPath(), name).String())
}

/*
Path 类型实现 Stringer 接口，并将路径规范化
*/
func (path Path) String() string {
	return filepath.Clean(filepath.Join(string(path)))
}

// InitFile 初始化文本文件，要求传入路径事先规范化过
func InitFile(name Path, text string) {
	e, err := name.Exists()
	if !Check(err) {
		zap.S().Errorf("初始化时不确定 %s 存在喵！\n%v", name, err)
		return
	}
	// 如果文件不存在，初始化
	if !e {
		name.Write([]byte(text))
	}
}

// LoadMainConfig 加载主配置
func LoadMainConfig() (config Config) {
	if err := yaml.Unmarshal(path.Read(), &config); !Check(err) {
		zap.S().Fatalf("打开 %s 失败了喵！\n%v", path, err)
	}
	return
}
