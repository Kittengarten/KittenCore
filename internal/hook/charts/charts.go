package charts

import (
	_ "unsafe"
	"weak"

	"github.com/Kittengarten/KittenCore/internal/hook"

	"github.com/brahma-adshonor/gohook"
	"github.com/golang/freetype/truetype"
	"github.com/vicanso/go-charts/v2"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
)

const funcGetFont = `GetFont`

// 备份 charts.GetFont
var chartsGetFontBackup = charts.GetFont

func init() {
	if err := gohook.Hook(
		charts.GetFont,
		getFont,
		chartsGetFontBackup,
	); err != nil {
		panic(err)
	}
}

// FontName 字体名称
const FontName = `glow`

// Font 字体
var Font weak.Pointer[truetype.Font]

// LoadFont 加载字体
func LoadFont() (*truetype.Font, error) {
	if f := Font.Value(); f != nil {
		return f, nil
	}
	// 获取字体数据
	buf, err := file.GetLazyData(text.GlowSansFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	return truetype.Parse(buf)
}

// hook charts.GetFont
func getFont(fontFamily string) (*truetype.Font, error) {
	if fontFamily == FontName {
		hook.DebugLog(funcGetFont)
		f, err := LoadFont()
		if err != nil {
			return nil, err
		}
		Font = weak.Make(f)
		return f, nil
	}
	return chartsGetFontBackup(fontFamily)
}
