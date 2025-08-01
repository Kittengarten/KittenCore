package stack2

import (
	"strings"
	"time"
)

type loc = byte // 地区

const (
	cat       loc = iota // 叠猫猫
	fox                  // 叠狐狐
	gpu                  // 叠显卡
	cockroach            // 叠蟑螂
)

const (
	cockroachDoNotAnalysis = `蟑螂不会分析，蟑螂只会勇敢地创上去`
)

var (
	// 各地区的字符串，运行时不可修改
	l10nStr = [...]map[loc]string{
		{cat: `猫猫`, fox: `狐狐`, gpu: `显卡`, cockroach: `蟑螂`},
		{cat: `只猫猫`, fox: `只狐狐`, gpu: `张显卡`},
		{cat: ` 只`, gpu: ` 张`},
		{cat: `	只`, gpu: `	张`},
		{cat: `平地摔了喵`, gpu: `超到了5G`, fox: `平地摔了嘤`, cockroach: `翻了个身`},
		{cat: `发生平地摔`, gpu: `超到了5G`, cockroach: `翻了个身`},
		{cat: `的平地摔`, gpu: `要超5G`},
		{cat: `平地摔`, gpu: `超5G`, cockroach: `翻了个身`},
		{cat: `触发了清空猫堆的特效`, fox: `获取了秘籍`, cockroach: `获得了康复新液`},
		{cat: `触发特效`, gpu: `打了鸡血驱动`, cockroach: `获得康复新液`},
		{cat: `猫堆高度`, gpu: `插槽容量`},
		{cat: `猫堆`, fox: `狐堆`, gpu: `PCIe`, cockroach: `蟑螂巢穴`},
		{cat: `压坏概率`, gpu: `冒烟概率`},
		{cat: `压坏`, gpu: `超冒烟`, cockroach: `压爆浆`},
		{cat: `体重为`, fox: `修为有`},
		{cat: `体重`, fox: `修为`, gpu: `算力`, cockroach: `翼展`},
		{cat: `绒布球`, gpu: `亮机卡`, cockroach: `蟑螂卵鞘`},
		{cat: `奶猫`, fox: `奶狐`, gpu: `GT1030`, cockroach: `德国蟑螂`},
		{cat: `抱枕`, gpu: `GT1630`, cockroach: `美洲蟑螂`},
		{cat: `小可爱`, gpu: `GTX1060`, cockroach: `广东蟑螂`},
		{cat: `大可爱`, gpu: `RTX2060`, cockroach: `澳洲蟑螂`},
		{cat: `幼年猫娘`, fox: `幼年狐娘`, gpu: `RTX3060`, cockroach: `幼年蟑螂娘`},
		{cat: `猫娘萝莉`, fox: `狐娘萝莉`, gpu: `RTX4060`, cockroach: `蟑螂萝莉`},
		{cat: `猫娘少女`, fox: `狐娘少女`, gpu: `RTX4060Ti`, cockroach: `蟑螂少女`},
		{cat: `成年猫娘`, fox: `二尾狐娘`, gpu: `RTX4070`, cockroach: `成年蟑螂娘`},
		{cat: `小老虎`, fox: `三尾狐娘`, gpu: `RTX4070S`, cockroach: `精英蟑螂娘`},
		{cat: `大老虎`, fox: `四尾狐娘`, gpu: `RTX4070Ti`, cockroach: `蟑螂母体`},
		{cat: `猫车`, fox: `五尾狐娘`, gpu: `RTX4070TiS`, cockroach: `蟑螂恶霸`},
		{cat: `猫猫巴士`, fox: `六尾狐娘`, gpu: `RTX4080`, cockroach: `蟑螂基地车`},
		{cat: `猫卡`, fox: `七尾狐娘`, gpu: `RTX4080S`, cockroach: `蟑螂运输船`},
		{cat: `虎式坦克`, fox: `八尾狐娘`, gpu: `RTX4080Ti`, cockroach: `蟑螂轨道炮`},
		{cat: `■■■`, fox: `九尾狐娘`, gpu: `RTX4090`, cockroach: `蟑螂歼星舰`},
		{cat: `猫娘以上`, gpu: `中高端显卡`},
		{cat: `猫娘`, fox: `狐娘`, gpu: `中高端显卡`, cockroach: `蟑螂娘`},
		{cat: `休息`, fox: `闭关`, gpu: `断电`, cockroach: `蛰伏`},
		{cat: `活动`, fox: `出关`, gpu: `通电`},
		{cat: `kg`, fox: `年`, gpu: `TFLOPS`, cockroach: `cm`},
		{cat: `🙀`, fox: `🦊`, gpu: `🖼️`, cockroach: `🪳`},
		{cat: `😿`, fox: `🦊`, gpu: `🧩`, cockroach: `🪳`},
		{cat: `🐅`, fox: `🦊`, gpu: `🎨`, cockroach: `🪳`},
		{cat: `🐯`, fox: `🦊`, gpu: `📦`, cockroach: `🪳`},
		{cat: `喵！`, fox: `嘤！`, gpu: `！`, cockroach: `！`},
		{cat: `喵～`, fox: `嘤～`, gpu: `～`, cockroach: `～`},
		{cat: `总重量为`, fox: `总修为有`, gpu: `总算力为`, cockroach: `总长度为`},
		{cat: `小猫咪`, gpu: `低功耗显卡`},
		{cat: `小猫`, gpu: `刀卡`},
		{cat: `摔成绒布球`, gpu: `沦为亮机卡`},
		{cat: `猫咪`, fox: `狐狸`, gpu: `显卡`},
		{cat: `有老虎`, fox: `有狐仙`, gpu: `有显卡`},
		{cat: `饿`, gpu: `花`},
		{cat: `嗷呜`, gpu: `加电压`},
		{cat: `美味`, gpu: `崭新`},
		{cat: `被老虎吃掉`, fox: `被狐仙吃掉`, gpu: `被拉去炼丹`},
		{cat: `老虎`, fox: `三尾以上狐娘`, gpu: `高端卡`, cockroach: `精英级以上蟑螂`},
		{cat: `今天吃猫`, gpu: `加卡加卡`},
		{cat: `吃掉`, gpu: `NVLink`},
		{cat: `吃掉`, gpu: `抢走`},
		{cat: `吃`, gpu: `抢`},
		{cat: `猫`, fox: `狐`, gpu: `卡`},
		{cat: `猫咪胖胖`, fox: `修为大增`, gpu: `算力暴涨`},
		{cat: `摔下去`, gpu: `掉驱动`},
		{cat: `摔下`, gpu: `掉驱`},
		{cat: `她`, gpu: `它`},
		{cat: `别的猫猫`, gpu: ``},
		{cat: `减肥`, gpu: `降频`},
		{cat: `长大`, gpu: `超频`},
		{cat: `猪咪王`, gpu: `战术核显卡`},
		{cat: `猪咪`, gpu: `核弹`},
		{cat: `猪`, gpu: `矿卡`},
		{cat: `压`, gpu: `超`},
		{cat: `摔`, gpu: `掉`},
		{cat: `床头叠上床尾摔`, gpu: `核心超上显存崩`},
		{cat: `锻炼`, fox: `化功`, gpu: `加速`, cockroach: `起飞`},
	}
	// 地区标记位
	globalLocation loc
	// 字符替换器
	l10n = func() *strings.Replacer {
		if globalLocation == cat {
			// 叠猫猫无需替换
			return strings.NewReplacer()
		}
		s := make([]string, 0, 2*len(l10nStr))
		for _, v := range l10nStr {
			newStr, ok := v[globalLocation]
			if !ok {
				continue
			}
			s = append(s, v[cat], newStr)
		}
		return strings.NewReplacer(s...)
	}()
)

// 叠蟑螂活动日期判断，在愚人节的前三天或后七天范围内返回 true
func checkCockroachDate() bool {
	switch time.Now().Month() {
	case time.March:
		return 28 < time.Now().Day()
	case time.April:
		return 9 > time.Now().Day()
	default:
		return false
	}
}
