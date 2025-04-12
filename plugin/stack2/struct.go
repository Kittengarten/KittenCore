package stack2

import (
	"math"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
)

const (
	noReason result = iota // 未知原因
	flat                   // 平地摔
	fall                   // 摔下去
	press                  // 压坏
	pressed                // 被压坏
	eat                    // 吃猫猫
	eaten                  // 被吃
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
		weight int    // 达到下一个等级的重量
		str    string // 字符串表示
	}

	// 叠猫猫配置
	config struct {
		RestHoursPerKG int // 每千克体重的休息小时数
		MinRestHours   int // 最小休息小时数
		OCMinRestHours int // 加速的最小休息小时数
	}

	// 叠猫猫缓存
	buffer struct {
		MedianWeight int           // 当前猫池中位数重量（0.1 kg 数）
		MaxRestTime  time.Duration // 最大休息时间
	}

	data []meow // 叠猫猫数据

	// 猫猫数据值
	meow struct {
		kitten.QQ `yaml:"id"` // QQ
		Name      string      `yaml:",omitempty"` // 群名片或昵称
		Weight    int         // 体重（0.1 kg 数）
		Status    bool        // 是否在叠猫猫中
		Time      time.Time   // 如果在叠猫猫中，叠入的时间；如果未在叠猫猫中，休息结束的时间
		// Stat             // 统计信息
	}

	// Stat 统计信息
	Stat struct {
		In        In        // 加入次数
		Exit      Exit      // 退出次数
		Time      time.Time // 总时长
		Max       int       // 曾经达到的最大高度
		MaxWeight int       // 曾经达到的最大重量
	}

	// In 加入统计信息
	In struct {
		Success int  // 成功
		Fall    Fail // 摔下去
		Press   Fail // 压坏
		Flat    int  // 平地摔次数
	}

	// Fail 失败统计信息
	Fail struct {
		Count int    // 失败次数
		Max   Record // 单次导致退出猫猫的最大值
		Total Record // 导致退出的猫猫总和
	}

	// Record 纪录
	Record struct {
		Count  int // 最大数量
		Weight int // 最大重量（0.1 kg 数）
	}

	// Exit 退出统计信息
	Exit struct {
		Fall    int // 摔下去次数
		Pressed int // 被压坏次数
		// Active  int // 主动退出次数
	}
)
