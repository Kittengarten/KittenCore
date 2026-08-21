package stack2

import (
	"github.com/Kittengarten/KittenCore/internal/config"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/msg"

	"github.com/vicanso/go-charts/v2"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// 生成并发送图片
func sendImage(handler *msg.Handler, p *charts.Painter, r ...replacer) message.ID {
	if len(r) == 0 {
		r = []replacer{l10n(cat)}
	}
	buf, err := p.Bytes()
	p.Close()
	p = nil
	if err != nil {
		return sendWithImageFail(handler, r[0], err)
	}
	path := fio.NewPath(imagePath, `叠猫猫.png`)
	if err = fio.NewPath(config.ImagePath().String(), path.String()).
		WriteBytes(buf); err != nil {
		return sendWithImageFail(handler, r[0], err)
	}
	return handler.Quote().Image(path).Send()
}
