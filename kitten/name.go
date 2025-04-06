package kitten

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"

	"golang.org/x/exp/constraints"
)

type name map[QQ]string // 昵称配置

// 当前昵称文件
var nameFile = fio.NewPath(`data`, `zbp`, `name.yaml`).WithRWMutex()

// ErrNotDefaultName 不是预设的昵称喵！
var ErrNotDefaultName = errors.New(`不是预设的昵称喵！`)

// Name 获取当前的 bot 昵称
func (u *QQ) Name() (string, error) {
	nameFile.RLock()
	defer nameFile.RUnlock()
	n, err := fio.Load[name](nameFile.Path, fio.Blank)
	if err != nil {
		return ``, err
	}
	return cmp.Or(n[*u], botConfig.NickName[0]), nil
}

// ReplaceCard 替换当前 bot 的群昵称
func (m *Messager) ReplaceCard(n string) error {
	o, err := m.Object()
	if err != nil {
		return err
	}
	if err := o.SetName(n); err != nil {
		return err
	}
	m.SetCard(-1)
	return nil
}

// SetName 设置当前的 bot 昵称
func (u *QQ) SetName(nickname string) error {
	if !slices.Contains(botConfig.NickName, nickname) {
		return fmt.Errorf(`“%s”%w`, nickname, ErrNotDefaultName)
	}
	nameFile.Lock()
	defer nameFile.Unlock()
	n, err := fio.Load[name](nameFile.Path, fio.Blank)
	if err != nil {
		return err
	}
	n[*u] = nickname
	return fio.Save(nameFile.Path, n)
}

// SetCardThisGroup 在本群设置自己的群昵称，h 为猫堆高度
func (m *Messager) SetCardThisGroup(h ...int) {
	if len(h) == 0 {
		// 默认高度为 -1
		h = []int{-1}
	}
	if m.Event.DetailType == Group {
		m.SetThisGroupCard(botConfig.SelfID.Int(), m.card(h[0]))
	}
}

/*
SetCard 设置自己的群昵称

h 为猫堆高度，g 为发送对象，若无则使用当前发送对象
*/
func (m *Messager) SetCard(h int, g ...QQ) {
	if len(g) == 0 {
		// 当前群
		m.SetCardThisGroup(h)
		return
	}
	for _, v := range g {
		if !v.IsGroup() {
			continue
		}
		m.SetGroupCard(v.Int(), botConfig.SelfID.Int(), m.card(h))
	}
}

/*
根据发送对象生成群昵称

h 为猫堆高度，g 为发送对象，若无则使用当前发送对象
*/

func (m *Messager) card(h int, g ...QQ) string {
	if len(g) == 0 {
		g = []QQ{*NewQQGroup(m.Event.GroupID)}
	}
	n, err := g[0].Name()
	if err != nil {
		Error(err)
		return ``
	}
	return card(n, botConfig.SelfID.Age(m), h)
}

// 生成群昵称，h 为猫堆高度
func card[T constraints.Integer](name string, age T, h int) string {
	return fmt.Sprintf(`%s（%d岁）（%s）`,
		name,
		age,
		func() string {
			switch cmp.Compare(0, h) {
			case -1:
				return `猫堆高度：` + strconv.Itoa(h)
			case 0:
				return `猫堆已清空`
			case 1:
				return `内置冷却，禁止调戏`
			default:
				// 死码，为了防止编译器报 missing return 而保留
				return ``
			}
		}(),
	)
}
