package usr

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"

	"github.com/RomiChan/syncx"
	"github.com/tidwall/gjson"
)

type (
	// QQ 信息
	qqInfo struct {
		gjson.Result // 信息
		time.Time    // 上次更新时间
	}

	qqInfoDisk struct {
		Raw  string
		Time time.Time
	}

	// 群成员列表
	groupList struct {
		List      []gjson.Result // 每个群员的信息
		time.Time                // 上次更新时间
	}

	groupListDisk struct {
		List []string
		Time time.Time
	}
)

const expire = time.Hour // 缓存过期时间 1 小时

var (
	infoFile   = fio.NewPath(`data`, `zbp`, `stranger.json`)
	memberFile = fio.NewPath(`data`, `zbp`, `member.json`)
)

var (
	strangerInfo    syncx.Map[QQ, qqInfo]    // 各陌生人信息缓存
	groupMemberList syncx.Map[QQ, groupList] // 各群的成员列表缓存
)

func init() {
	if err := errors.Join(loadInfo(), loadMember()); err != nil {
		log.Panic(err)
	}
}

func loadInfo() error {
	data := new(map[QQ]qqInfoDisk)
	if err := load(infoFile, data); err != nil {
		return err
	}
	for k, v := range *data {
		strangerInfo.Store(k, qqInfo{
			Result: gjson.Parse(v.Raw),
			Time:   v.Time,
		})
	}
	return nil
}

func loadMember() error {
	data := new(map[QQ]groupListDisk)
	if err := load(memberFile, data); err != nil {
		return err
	}
	for k, v := range *data {
		groupMemberList.Store(k, groupList{
			List: utils.ConvertSlice(v.List, func(s string) gjson.Result {
				return gjson.Parse(s)
			}),
			Time: v.Time,
		})
	}
	return nil
}

func load(p fio.Path, v any) error {
	if err := p.InitFileText(fio.Blank); err != nil {
		return err
	}
	f, err := p.Load(false)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func saveInfo() error {
	infoDump := make(map[QQ]qqInfoDisk)
	strangerInfo.Range(func(k QQ, v qqInfo) bool {
		infoDump[k] = qqInfoDisk{
			Raw:  v.Result.Raw,
			Time: v.Time,
		}
		return true
	})
	return save(infoFile, infoDump)
}

func saveMember() error {
	memberDump := make(map[QQ]groupListDisk)
	groupMemberList.Range(func(k QQ, v groupList) bool {
		memberDump[k] = groupListDisk{
			List: utils.ConvertSlice(v.List, func(raw gjson.Result) string {
				return raw.Raw
			}),
			Time: v.Time,
		}
		return true
	})
	return save(memberFile, memberDump)
}

func save(p fio.Path, v any) error {
	f, err := p.Load(true)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent(``, `  `)
	return enc.Encode(v)
}
