package abuse

type (
	Response struct {
		Id     int    `yaml:"id"`     // 回复编号
		String string `yaml:"string"` // 回复字符串
		Image  string `yaml:"image"`  // 回复图片文件名
		Chance int    `yaml:"chance"` // 回复权重
	}
)

// 实现该项目的信息
func (re Response) GetInformation() string {
	return re.String + re.Image
}

// 实现该项目的权重
func (re Response) GetChance() int {
	return re.Chance
}
