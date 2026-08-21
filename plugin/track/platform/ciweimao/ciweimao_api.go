// Package ciweimao 刺猬猫阅读
package ciweimao

import (
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/plugin/track/platform"
)

// CWM 刺猬猫阅读
type CWM utils.Object

// HOST 网站
const HOST = `https://www.ciweimao.com`

// Export 导出平台接口
var Export platform.Platform
