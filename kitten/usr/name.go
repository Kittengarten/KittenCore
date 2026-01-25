package usr

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Kittengarten/KittenCore/internal/config"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
)

type name map[string]string // 昵称配置，key 为 QQ，value 为昵称

// 当前昵称文件
var nameFile = fio.NewPath(`data`, `zbp`, `name.yaml`).WithRWMutex()

// ErrNotDefaultName 不是预设的昵称喵！
var ErrNotDefaultName = errors.New(`不是预设的昵称喵！`)

// Name 获取当前的 bot 昵称
func (u QQ) Name(ctx context.Context) (string, error) {
	nameFile.RLock()
	defer nameFile.RUnlock()
	n, err := fio.LoadWithContext[name](ctx, nameFile.Path, fio.Blank)
	if err != nil {
		return ``, err
	}
	return cmp.Or(append([]string{n[u.Str()]}, config.Name()...)...), nil
}

// SetName 设置当前的 bot 昵称
func (u QQ) SetName(ctx context.Context, nickname string) error {
	if !slices.Contains(config.Name(), nickname) {
		return fmt.Errorf(`“%s”%w`, nickname, ErrNotDefaultName)
	}
	nameFile.Lock()
	defer nameFile.Unlock()
	n, err := fio.LoadWithContext[name](ctx, nameFile.Path, fio.Blank)
	if err != nil {
		return err
	}
	n[u.Str()] = nickname
	return fio.SaveWithContext(ctx, nameFile.Path, n)
}
