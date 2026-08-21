package stack2

import (
	"strings"
	"time"
)

type (
	location = byte // 地区
	replacer = struct {
		*strings.Replacer
		location
	} // 字符替换器，包含地区标记位
)

const (
	cat       location = iota // 叠猫猫
	fox                       // 叠狐狐
	gpu                       // 叠显卡
	cockroach                 // 叠蟑螂
	modem                     // 叠猫（调制解调器）
)

const (
	cockroachDoNotAnalysis = `蟑螂不会分析，蟑螂只会勇敢地创上去`
)

// 各地区的字符串，运行时不可修改
var l10nStr = [...]map[location]string{
	{cat: `猫猫`, fox: `狐狐`, gpu: `显卡`, cockroach: `蟑螂`, modem: `猫`},
	{cat: `只猫猫`, fox: `只狐狐`, gpu: `张显卡`, modem: `台猫`},
	{cat: ` 只`, gpu: ` 张`, modem: ` 台`},
	{cat: `	只`, gpu: `	张`, modem: `	台`},
	{cat: `平地摔了喵`, gpu: `超到了5G`, fox: `平地摔了嘤`, cockroach: `翻了个身`, modem: `解锁了带宽`},
	{cat: `发生平地摔`, gpu: `超到了5G`, cockroach: `翻了个身`, modem: `解锁了带宽`},
	{cat: `的平地摔`, gpu: `要超5G`, modem: `要解带宽`},
	{cat: `平地摔`, gpu: `超5G`, cockroach: `翻了个身`, modem: `解锁带宽`},
	{cat: `触发了清空猫堆的特效`, fox: `获取了秘籍`, gpu: `打了鸡血驱动`, cockroach: `获得了康复新液`, modem: `换了更好的运营商`},
	{cat: `触发特效`, fox: `获取秘籍`, gpu: `打鸡血驱动`, cockroach: `获得康复新液`, modem: `换好运营商`},
	{cat: `猫堆高度`, gpu: `插槽容量`},
	{cat: `猫堆`, fox: `狐堆`, gpu: `PCIe`, cockroach: `蟑螂巢穴`},
	{cat: `压坏概率`, gpu: `冒烟概率`, modem: `断网概率`},
	{cat: `压坏`, gpu: `超冒烟`, cockroach: `压爆浆`, modem: `断网`},
	{cat: `体重为`, fox: `修为有`, modem: `带宽为`},
	{cat: `体重`, fox: `修为`, gpu: `算力`, cockroach: `翼展`, modem: `带宽`},
	{cat: `绒布球`, gpu: `亮机卡`, cockroach: `蟑螂卵鞘`, modem: `拨号上网猫`},
	{cat: `奶猫`, fox: `奶狐`, gpu: `GT1030`, cockroach: `德国蟑螂`, modem: `ISDN猫`},
	{cat: `抱枕`, gpu: `GT1050`, cockroach: `美洲蟑螂`, modem: `ADSL猫`},
	{cat: `小可爱`, gpu: `GTX1060`, cockroach: `广东蟑螂`, modem: `ADSL2猫`},
	{cat: `大可爱`, gpu: `RTX2060`, cockroach: `澳洲蟑螂`, modem: `ADSL2+猫`},
	{cat: `幼年猫娘`, fox: `幼年狐娘`, gpu: `RTX3060`, cockroach: `幼年蟑螂娘`, modem: `VDSL猫`},
	{cat: `猫娘萝莉`, fox: `狐娘萝莉`, gpu: `RTX4060`, cockroach: `蟑螂萝莉`, modem: `VDSL2猫`},
	{cat: `猫娘少女`, fox: `狐娘少女`, gpu: `RTX5060`, cockroach: `蟑螂少女`, modem: `有线机顶盒`},
	{cat: `成年猫娘`, fox: `二尾狐娘`, gpu: `RTX5060Ti`, cockroach: `成年蟑螂娘`, modem: `APON光猫`},
	{cat: `小老虎`, fox: `三尾狐娘`, gpu: `RTX5070`, cockroach: `精英蟑螂娘`, modem: `BPON光猫`},
	{cat: `大老虎`, fox: `四尾狐娘`, gpu: `RTX5070Ti`, cockroach: `蟑螂母体`, modem: `EPON光猫`},
	{cat: `猫车`, fox: `五尾狐娘`, gpu: `RTX5080`, cockroach: `蟑螂恶霸`, modem: `GPON光猫`},
	{cat: `猫猫巴士`, fox: `六尾狐娘`, gpu: `RTX5090Dv2`, cockroach: `蟑螂基地车`, modem: `XGS-PON光猫`},
	{cat: `猫卡`, fox: `七尾狐娘`, gpu: `RTX5090D`, cockroach: `蟑螂运输船`, modem: `50G-PON光猫`},
	{cat: `虎式坦克`, fox: `八尾狐娘`, gpu: `RTX5090`, cockroach: `蟑螂轨道炮`, modem: `100G-PON光猫`},
	{cat: `■■■`, fox: `九尾狐娘`, gpu: `B200`, cockroach: `蟑螂歼星舰`, modem: `超高速PON光猫`},
	{cat: `猫娘以上`, gpu: `中高端显卡`, modem: `中高端猫`},
	{cat: `猫娘`, fox: `狐娘`, gpu: `中高端显卡`, cockroach: `蟑螂娘`, modem: `中高端猫`},
	{cat: `休息`, fox: `闭关`, gpu: `断电`, cockroach: `蛰伏`, modem: `断网`},
	{cat: `活动`, fox: `出关`, gpu: `通电`, modem: `联网`},
	{cat: `kg`, fox: `年`, gpu: `TFLOPS`, cockroach: `cm`, modem: `Mbps`},
	{cat: `🙀`, fox: `🦊`, gpu: `🖼️`, cockroach: `🪳`, modem: `🛜`},
	{cat: `😿`, fox: `🦊`, gpu: `🧩`, cockroach: `🪳`, modem: `🌐`},
	{cat: `🐅`, fox: `🦊`, gpu: `🎨`, cockroach: `🪳`, modem: `🔗`},
	{cat: `🐯`, fox: `🦊`, gpu: `📦`, cockroach: `🪳`, modem: `📡`},
	{cat: `喵！`, fox: `嘤！`, gpu: `！`, cockroach: `！`, modem: `！`},
	{cat: `喵～`, fox: `嘤～`, gpu: `～`, cockroach: `～`, modem: `～`},
	{cat: `总重量为`, fox: `总修为有`, gpu: `总算力为`, cockroach: `总长度为`, modem: `总带宽为`},
	{cat: `小猫咪`, fox: `小狐狸`, gpu: `低功耗显卡`, cockroach: `小蟑螂`, modem: `古董猫`},
	{cat: `的小猫`, fox: `的小狐`, gpu: `的刀卡`, cockroach: `的小蠊`, modem: `的古董猫`},
	{cat: `摔成绒布球`, gpu: `沦为亮机卡`, cockroach: `被拖鞋打扁`, modem: `上转转回收`},
	{cat: `猫咪`, fox: `狐狸`, gpu: `显卡`, cockroach: `蟑螂`, modem: `猫`},
	{cat: `有老虎`, fox: `有狐仙`, gpu: `有显卡`, cockroach: `有蟑螂`, modem: `有猫`},
	{cat: `饿`, gpu: `花`, modem: `断网`},
	{cat: `嗷呜`, gpu: `加电压`, modem: `多拨`},
	{cat: `美味`, gpu: `崭新`, modem: `传家宝`},
	{cat: `被老虎吃掉`, fox: `被狐仙吃掉`, gpu: `被拉去炼丹`, modem: `被占满带宽`},
	{cat: `老虎`, fox: `三尾以上狐娘`, gpu: `高端卡`, cockroach: `精英级以上蟑螂`, modem: `新光猫`},
	{cat: `今天吃猫`, gpu: `加卡加卡`, modem: `延迟稳定`},
	{cat: `吃掉`, gpu: `NVLink`, modem: `链路聚合`},
	{cat: `吃猫猫`, fox: `吃狐狐`, gpu: `抢显卡`, cockroach: `拍蟑螂`, modem: `抢猫`},
	{cat: `猫`, fox: `狐`, gpu: `卡`},
	{cat: `猫咪胖胖`, fox: `修为大增`, gpu: `算力暴涨`, modem: `带宽充足`},
	{cat: `摔下去`, gpu: `掉驱动`, modem: `过热`},
	{cat: `摔下`, gpu: `掉驱`, modem: `过热`},
	{cat: `她`, gpu: `它`, cockroach: `它`, modem: `它`},
	{cat: `别的猫猫`, gpu: ``, modem: ``},
	{cat: `减肥`, gpu: `降频`, modem: `限速`},
	{cat: `长大`, gpu: `超频`, modem: `提速`},
	{cat: `猪咪王`, gpu: `战术核显卡`, cockroach: `蟑螂王`, modem: `骨干网`},
	{cat: `猪咪`, gpu: `核弹`, cockroach: `白云机场`, modem: `主干`},
	{cat: `请勿给猪染色`, gpu: `请勿给矿卡超频`, cockroach: `请勿饲养蟑螂`, modem: `当心多图杀猫`},
	{cat: `压`, gpu: `超`},
	{cat: `摔`, gpu: `掉`},
	{cat: `床头叠上床尾摔`, gpu: `核心超上显存崩`},
	{cat: `锻炼`, fox: `化功`, gpu: `加速`, cockroach: `起飞`, modem: `重启`},
}

// 转换地区标记位及其可用性
func parseLoc(loc string) (replacer, bool) {
	if strings.Contains(loc, `光猫`) ||
		strings.Contains(loc, `调制解调器`) ||
		strings.Contains(loc, `路由器`) ||
		strings.Contains(loc, `宽带`) ||
		strings.ContainsAny(loc, `网络`) {
		return l10n(modem), true
	}
	switch {
	case strings.ContainsAny(loc, `狐狸`):
		return l10n(fox), true // 狐狐
	case strings.Contains(loc, `显卡`):
		return l10n(gpu), true // 显卡
	case strings.ContainsAny(loc, `蟑螂`),
		strings.ContainsAny(loc, `蜚蠊`),
		strings.Contains(loc, `小强`):
		return l10n(cockroach), checkCockroachDate() // 蟑螂
	case strings.ContainsAny(loc, `猫虎喵貓`):
		fallthrough // 猫猫
	default:
		return l10n(cat), true // 默认叠猫猫
	}
}

// 字符替换器

func l10n(loc location) replacer {
	if loc == cat {
		// 叠猫猫无需替换
		return replacer{strings.NewReplacer(), loc}
	}
	s := make([]string, 0, 2*len(l10nStr))
	for _, v := range l10nStr {
		newStr, ok := v[loc]
		if !ok {
			continue
		}
		s = append(s, v[cat], newStr)
	}
	return replacer{strings.NewReplacer(s...), loc}
}

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
