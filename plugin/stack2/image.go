package stack2

import (
	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/io"

	"github.com/vicanso/go-charts/v2"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// 生成并发送图片
func sendImage(msgr *kitten.Messager, p *charts.Painter) message.ID {
	buf, err := p.Bytes()
	defer p.Close()
	if err != nil {
		return sendWithImageFail(msgr, err)
	}
	path := io.NewPath(imagePath, `叠猫猫.png`)
	if err = io.NewPath(kitten.ImagePath().String(), path.String()).
		WriteBytes(buf); err != nil {
		return sendWithImageFail(msgr, err)
	}
	return msgr.Reply().Image(path).Send()
}
