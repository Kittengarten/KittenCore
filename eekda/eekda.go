package eekda

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/FloatTech/floatbox/math"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/Kittengarten/KittenCore/kitten"

	log "github.com/sirupsen/logrus"
)

const (
	// ReplyServiceName 插件名
	ReplyServiceName = "eekda"
	brief            = "今天吃什么"
	namePath         = "eekda/name.txt" // 保存名字的文件
	todayFile        = "today.yaml"     // 保存今天吃什么的文件
	statFile         = "stat.yaml"      // 保存统计数据的文件
)

func init() {
	var (
		name = getName()
		help = strings.Join([]string{"发送",
			fmt.Sprintf("%s%s今天吃什么", kitten.Configs.CommandPrefix, name),
			fmt.Sprintf("获取%s今日食谱", name),
			fmt.Sprintf("%s查询被吃次数", kitten.Configs.CommandPrefix),
			"查询本人被吃次数",
		}, "\n")
		// 注册插件
		engine = control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
			DisableOnDefault:  true,
			Brief:             brief,
			Help:              help,
			PrivateDataFolder: "eekda",
		})
	)

	engine.OnCommand(fmt.Sprintf("%s今天吃什么", name)).Handle(func(ctx *zero.Ctx) {
	re:
		var (
			today, err = kitten.Path(engine.DataFolder() + todayFile).Read()
			todayData  Today
		)
		if kitten.Check(err) {
			yaml.Unmarshal(today, &todayData)
			if kitten.IsSameDate(time.Now(), todayData.Time) {
				report(todayData, name, ctx)
			} else {
				// 生成今天吃什么
				list := ctx.GetThisGroupMemberListNoCache().Array()
				sort.SliceStable(list, func(i, j int) bool {
					return list[i].Get("last_sent_time").Int() < list[j].Get("last_sent_time").Int()
				})
				list = list[math.Max(0, len(list)-50):]
				nums := generateRandomNumber(0, len((list)), 5)
				todayData.Time = time.Now()
				todayData.Breakfast = kitten.QQ(list[nums[0]].Get("user_id").Int())
				todayData.Lunch = kitten.QQ(list[nums[1]].Get("user_id").Int())
				todayData.LowTea = kitten.QQ(list[nums[2]].Get("user_id").Int())
				todayData.Dinner = kitten.QQ(list[nums[3]].Get("user_id").Int())
				todayData.Supper = kitten.QQ(list[nums[4]].Get("user_id").Int())
				today, err1 := yaml.Marshal(todayData)
				err2 := kitten.Path(engine.DataFolder() + todayFile).Write(today)
				if !kitten.Check(err1) || !kitten.Check(err2) {
					log.Errorf("生成今天吃什么发生错误：\n%v\n%v", err1, err2)
				}
				report(todayData, name, ctx)

				// 存储饮食统计数据
				var (
					stat, err = kitten.Path(engine.DataFolder() + statFile).Read()
					statData  Stat
					isNew     = map[string]bool{
						"Breakfast": true,
						"Lunch":     true,
						"LowTea":    true,
						"Dinner":    true,
						"Supper":    true,
					}
					statMap = make(map[kitten.QQ]Kitten) // QQ:猫猫集合
				)
				if !kitten.Check(yaml.Unmarshal(stat, &statData)) {
					log.Warn("饮食统计数据损坏了喵！")
				} else if !kitten.Check(err) {
					log.Error("饮食统计数据无法访问喵！")
				}
				// 加载数据
				for _, v := range statData {
					statMap[v.ID] = v
				}
				bf, bfok := statMap[todayData.Breakfast]
				l, lok := statMap[todayData.Lunch]
				lt, ltok := statMap[todayData.LowTea]
				d, dok := statMap[todayData.Dinner]
				s, sok := statMap[todayData.Supper]
				isNew = map[string]bool{
					"Breakfast": !bfok,
					"Lunch":     !lok,
					"LowTea":    !ltok,
					"Dinner":    !dok,
					"Supper":    !sok,
				}
				// 修改数据
				switch true {
				case bfok:
					bf.Breakfast++
					statMap[todayData.Breakfast] = bf
					fallthrough
				case lok:
					l.Lunch++
					statMap[todayData.Lunch] = l
					fallthrough
				case ltok:
					lt.LowTea++
					statMap[todayData.LowTea] = lt
					fallthrough
				case dok:
					d.Dinner++
					statMap[todayData.Dinner] = d
					fallthrough
				case sok:
					s.Supper++
					statMap[todayData.Supper] = s
				}
				// 回写修改的数据
				for k, v := range statData {
					statData[k] = statMap[v.ID]
				}
				for k, v := range isNew {
					var new Kitten
					switch k {
					case "Breakfast":
						new = Kitten{
							ID:   todayData.Breakfast,
							Name: getLine(todayData.Breakfast, ctx),
						}
						if v {
							new.Breakfast = 1
						}
					case "Lunch":
						new = Kitten{
							ID:   todayData.Lunch,
							Name: getLine(todayData.Lunch, ctx),
						}
						if v {
							new.Lunch = 1
						}
					case "LowTea":
						new = Kitten{
							ID:   todayData.LowTea,
							Name: getLine(todayData.LowTea, ctx),
						}
						if v {
							new.LowTea = 1
						}
					case "Dinner":
						new = Kitten{
							ID:   todayData.Dinner,
							Name: getLine(todayData.Dinner, ctx),
						}
						if v {
							new.Dinner = 1
						}
					case "Supper":
						new = Kitten{
							ID:   todayData.Supper,
							Name: getLine(todayData.Supper, ctx),
						}
						if v {
							new.Supper = 1
						}
					}
					if v {
						statData = append(statData, new)
					}
				}
				stat, err3 := yaml.Marshal(statData)
				err4 := kitten.Path(engine.DataFolder() + statFile).Write(stat)
				if !kitten.Check(err3) || !kitten.Check(err4) {
					log.Errorf("写入饮食统计数据发生错误：\n%v\n%v", err1, err2)
				}
			}
		} else if isExist, err := kitten.Path(engine.DataFolder() + todayFile).Exists(); !kitten.Check(err) {
			// 如果不确定文件存在
			doNotKnow(ctx)
		} else if !isExist {
			// 如果文件不存在，创建文件后重新载入命令
			if fp, err := os.Create(engine.DataFolder() + todayFile); kitten.Check(err) {
				fp.WriteString("[]")
				defer fp.Close()
				goto re
			} else {
				doNotKnow(ctx)
			}
		}
	})

	engine.OnCommand("查询被吃次数").Handle(func(ctx *zero.Ctx) {
	re:
		var (
			stat, err = kitten.Path(engine.DataFolder() + statFile).Read()
			statData  Stat
			isGet     bool
		)
		if kitten.Check(err) {
			if !kitten.Check(yaml.Unmarshal(stat, &statData)) {
				doNotKnow(ctx)
			}
			for i, v := range statData {
				if kitten.QQ(ctx.Event.UserID) == v.ID {
					report := strings.Join([]string{fmt.Sprintf("\n%s的被吃次数", getLine(kitten.QQ(ctx.Event.UserID), ctx)),
						fmt.Sprintf("早餐：%d 次", statData[i].Breakfast),
						fmt.Sprintf("午餐：%d 次", statData[i].Lunch),
						fmt.Sprintf("下午茶：%d 次", statData[i].LowTea),
						fmt.Sprintf("晚餐：%d 次", statData[i].Dinner),
						fmt.Sprintf("夜宵：%d 次", statData[i].Supper),
					}, "\n")
					ctx.SendChain(message.At(ctx.Event.UserID), message.Text(report))
					isGet = true
				}
			}
			if !isGet {
				doNotKnow(ctx)
			}
		} else if isExist, err := kitten.Path(engine.DataFolder() + statFile).Exists(); !kitten.Check(err) {
			// 如果不确定文件存在
			doNotKnow(ctx)
		} else if !isExist {
			// 如果文件不存在，创建文件后重新载入命令
			if fp, err := os.Create(engine.DataFolder() + statFile); kitten.Check(err) {
				fp.WriteString("[]")
				defer fp.Close()
				goto re
			} else {
				doNotKnow(ctx)
			}
		}

	})
}

// 获取名字
func getName() string {
	res, err := os.Open(namePath)
	if !kitten.Check(err) {
		log.Warnf("打开文件 %s 失败了喵！\n%v", namePath, err)
	} else {
		defer res.Close()
	}
	data, err := io.ReadAll(res)
	if kitten.Check(err) {
		return string(data)
	}
	log.Warnf("打开文件 %s 失败了喵！\n%v", namePath, err)
	return ""
}

// 获取条目，u 为 QQ
func getLine(u kitten.QQ, ctx *zero.Ctx) string {
	info := Kitten{
		ID:   u,
		Name: u.GetTitle(*ctx) + ctx.CardOrNickName(int64(u)),
	}
	return fmt.Sprintf("%s（%d）", info.Name, info.ID)
}

// 播报今天吃什么，t 为今日数据
func report(t Today, name string, ctx *zero.Ctx) {
	report := strings.Join([]string{fmt.Sprintf("【%s今天吃什么】", name),
		fmt.Sprintf("早餐：%s", getLine(t.Breakfast, ctx)),
		fmt.Sprintf("午餐：%s", getLine(t.Lunch, ctx)),
		fmt.Sprintf("下午茶：%s", getLine(t.LowTea, ctx)),
		fmt.Sprintf("晚餐：%s", getLine(t.Dinner, ctx)),
		fmt.Sprintf("夜宵：%s", getLine(t.Supper, ctx)),
	}, "\n")
	ctx.Send(report)
}

// 喵喵不知道哦
func doNotKnow(ctx *zero.Ctx) {
	ctx.Send(fmt.Sprintf("%s不知道哦", zero.BotConfig.NickName[0]))
}

// 生成 count 个 [start, end) 范围的不重复的随机数
func generateRandomNumber(start int, end int, count int) []int {
	// 范围检查
	if end < start || (end-start) < count {
		return nil
	}
	// 存放结果的 slice
	nums := make([]int, 0)
	// 随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		var (
			// 生成随机数
			num = r.Intn((end - start)) + start
			// 查重
			exist bool
		)
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}
		if !exist {
			nums = append(nums, num)
		}
	}
	return nums
}
