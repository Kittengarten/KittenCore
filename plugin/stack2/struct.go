package stack2

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/usr"
)

const (
	noReason result = iota // 未知原因
	flat                   // 平地摔
	fall                   // 摔下去
	press                  // 压坏
	pressed                // 被压坏
	eat                    // 吃猫猫
	eaten                  // 被吃
	lorry                  // 撞大运
	fly                    // 被撞飞
)

const (
	绒布球 meowTypeID = iota
	奶猫
	抱枕
	小可爱
	大可爱
	幼年猫娘
	猫娘萝莉
	猫娘少女
	成年猫娘
	小老虎
	大老虎
	猫车
	猫猫巴士
	猫卡
	虎式坦克
	unknown
)

// 猫猫类型数据
var mapMeow = map[meowTypeID]meowType{
	绒布球:     {weight: 2, str: `绒布球`},
	奶猫:      {weight: 10, str: `奶猫`},
	抱枕:      {weight: 50, str: `抱枕`},
	小可爱:     {weight: 100, str: `小可爱`},
	大可爱:     {weight: 200, str: `大可爱`},
	幼年猫娘:    {weight: 300, str: `幼年猫娘`},
	猫娘萝莉:    {weight: 400, str: `猫娘萝莉`},
	猫娘少女:    {weight: 750, str: `猫娘少女`},
	成年猫娘:    {weight: 750, str: `成年猫娘`},
	小老虎:     {weight: 1500, str: `小老虎`},
	大老虎:     {weight: 5000, str: `大老虎`},
	猫车:      {weight: 50000, str: `猫车`},
	猫猫巴士:    {weight: 150000, str: `猫猫巴士`},
	猫卡:      {weight: 500000, str: `猫卡`},
	虎式坦克:    {weight: 1000000, str: `虎式坦克`},
	unknown: {weight: math.MaxInt, str: `■■■`},
}

type (
	// 叠猫猫退出原因
	result = byte

	// 猫猫类型序号
	meowTypeID byte

	// 猫猫类型
	meowType struct {
		str    string // 类型名称
		weight int    // 达到下一个等级的重量
	}

	// 叠猫猫配置
	cfg struct {
		RestHoursPerKG int `comment:"每千克体重的休息小时数" yaml:"rest_hours_per_kg"` // 每千克体重的休息小时数
		MinRestHours   int `comment:"最小休息小时数"     yaml:"min_rest_hours"`    // 最小休息小时数
		OCMinRestHours int `comment:"加速的最小休息小时数"  yaml:"oc_min_rest_hours"` // 加速的最小休息小时数
	}

	// 叠猫猫状态
	status struct {
		MedianWeight int           // 当前猫池中位数重量（0.1 kg 数）
		MaxRestTime  time.Duration // 最大休息时间
	}

	data []meow // 叠猫猫数据

	// 猫猫数据值
	meow struct {
		Time     time.Time   `yaml:",omitzero"` // 如果在叠猫猫中，叠入的时间；如果未在叠猫猫中，休息结束的时间
		Daily    time.Time   `yaml:",omitzero"`
		Name     string      `yaml:",omitzero"` // 群名片或昵称
		usr.QQ   `yaml:"id"` // QQ
		Weight   int         // 体重（0.1 kg 数）
		Status   bool        // 是否在叠猫猫中
		Location location    `yaml:"-"` // 地区标记位
	}
)

// Format 实现 fmt.Formatter，返回叠猫猫字符串
//
//	%s 完整字符串
//	%c 省略过的字符串
func (d *data) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') || state.Flag('#') {
			d.format(state, verb)
			return
		}
		_, _ = fmt.Fprint(state, d.String())
	case 's':
		_, _ = fmt.Fprint(state, d.String())
	case 'c':
		_, _ = fmt.Fprint(state, d.Str())
	default:
		d.format(state, verb)
	}
}

func (d *data) format(state fmt.State, verb rune) {
	type raw *data
	_, _ = fmt.Fprintf(state, fmt.FormatString(state, verb), raw(d))
}

// String 实现 fmt.Stringer
//
//	从叠猫猫队列生成完整字符串（开头有一次换行）
func (d *data) String() string {
	// 克隆一份防止修改源数据
	dr := slices.Clone(*d)
	// 按“后来居上”排列叠猫猫队列
	slices.Reverse(dr)
	s := new(strings.Builder)
	s.Grow(len(dr) << 5)
	for _, k := range dr {
		fmt.Fprint(s, "\n", k)
	}
	return s.String()
}

// 从叠猫猫队列生成省略过的字符串
//
//	队列高度不超过 20 时，无需省略
func (d *data) Str() string {
	var (
		dr = slices.Clone(*d) // 克隆一份防止修改源数据
		l  = len(dr)          // 叠猫猫队列高度
		s  = new(strings.Builder)
		ok bool
	)
	s.Grow(min(l, 20) << 5)
	// 按“后来居上”排列叠猫猫队列
	slices.Reverse(dr)
	for i, k := range dr {
		if l > 20 && 5 <= i && i < l-5 {
			// 当高度 > 20 时，跳过中间的猫猫，只取上下 5 只
			if ok {
				continue
			}
			s.WriteString("\n…………\n")
			for range l - 10 {
				s.WriteRune('🐱')
			}
			s.WriteString("\n…………")
			ok = true
			continue
		}
		s.WriteByte('\n')
		fmt.Fprint(s, k)
	}
	return s.String()
}

// Format 实现 fmt.Formatter
func (m meow) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') || state.Flag('#') {
			m.format(state, verb)
			return
		}
		_, _ = fmt.Fprint(state, m.String())
	case 's':
		_, _ = fmt.Fprint(state, m.String())
	default:
		m.format(state, verb)
	}
}

func (m meow) format(state fmt.State, verb rune) {
	// 改写类型以屏蔽内嵌 usr.QQ 的格式
	_, _ = fmt.Fprintf(state, fmt.FormatString(state, verb), struct {
		Time     time.Time
		Daily    time.Time
		Name     string
		QQ       int64
		Weight   int
		Status   bool
		Location location
	}{
		Time:     m.Time,
		Daily:    m.Daily,
		Name:     m.Name,
		QQ:       int64(m.QQ),
		Weight:   m.Weight,
		Status:   m.Status,
		Location: m.Location,
	})
}

// String 实现 fmt.Stringer
func (m meow) String() string {
	ctx, cancel := context.WithTimeout(context.Background(), shttp.Timeout)
	defer cancel()
	h := msg.NewWithContext(ctx, globalCtx)
	if m.Location == cockroach {
		return fmt.Sprintf(`【%s】	翼展 %.1f cm`, m.getType(h), i2f(m.Weight))
	}
	return fmt.Sprintf(
		`%s	❤	%d	❤	%.1f kg	%s`,
		cmp.Or(m.TitleCardOrNickName(h), m.Name),
		m.Int(),
		i2f(m.Weight),
		m.getType(h),
	)
}
