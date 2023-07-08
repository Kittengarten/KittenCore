package kitten

import (
	"os"
	"path/filepath"

	"github.com/wdvxdr1123/ZeroBot/message"
	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

// FilePath 文件路径构建
func FilePath(elem ...Path) Path {
	var s = make([]string, len(elem))
	for k, v := range elem {
		s[k] = v.String()
	}
	return Path(filepath.Join([]string(s)...))
}

// Read 文件读取
func (path Path) Read() (data []byte) {
	data, err := os.ReadFile(path.String())
	if !Check(err) {
		log.Errorf("打开文件 %s 失败了喵！\n%v", path, err)
	}
	return
}

/*
Write 文件写入

如文件不存在会尝试新建
*/
func (path Path) Write(data []byte) {
	e, err := path.Exists()
	if !e {
		// 如果文件或文件夹不存在，或不确定是否存在
		if Check(err) {
			// 如果文件不存在，新建该文件所在的文件夹；如果文件夹不存在，新建该文件夹本身
			os.MkdirAll(filepath.Dir(path.String()), os.ModeDir)
		} else {
			// 文件或文件夹不确定是否存在
			log.Warnf("写入时不确定 %s 存在喵！\n%v", path, err)
		}
	}
	err = os.WriteFile(path.String(), data, 0666)
	if !Check(err) {
		log.Errorf("写入文件 %s 失败了喵！\n%v", path, err)
	}
	return
}

/*
Exists 判断文件是否存在

不确定存在的情况下报错
*/
func (path Path) Exists() (bool, error) {
	_, err := os.Stat(path.String())
	// 当 err 为空，文件或文件夹存在
	if Check(err) {
		return true, nil
	}
	// os.IsNotExist(err)为 true，文件或文件夹不存在
	if os.IsNotExist(err) {
		return false, nil
	}
	// 其它类型，不确定是否存在
	return false, err
}

// （私有）判断路径是否文件夹
func (path Path) isDir() bool {
	s, err := os.Stat(path.String())
	if Check(err) {
		return s.IsDir()
	}
	log.Errorf("识别 %s 失败了喵！\n%v", path, err)
	return false
}

// LoadPath 加载文件中保存的相对路径或绝对路径
func (path Path) LoadPath() Path {
	data, err := os.ReadFile(path.String())
	if !Check(err) {
		log.Errorf("打开文件 %s 失败了喵！\n%v", path, err)
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
	return filepath.Clean(string(FilePath(path)))
}

// InitFile 初始化文本文件，要求传入路径事先规范化过
func InitFile(name Path, text string) {
	e, err := name.Exists()
	if !Check(err) {
		log.Warnf("初始化时不确定 %s 存在喵！\n%v", name, err)
		return
	}
	// 如果文件不存在，初始化
	if !e {
		name.Write([]byte(text))
	}
	return
}

// LoadMainConfig 加载主配置
func LoadMainConfig() (config Config) {
	if err := yaml.Unmarshal(path.Read(), &config); !Check(err) {
		log.Fatalf("打开 %s 失败了喵！\n%v", path, err)
	}
	return
}
