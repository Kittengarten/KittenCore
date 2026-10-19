package msg

import (
	"cmp"
	"context"
	"fmt"
	"image"
	"os"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/msg/mio"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// ScanQRCode 扫描二维码
func ScanQRCode(ctx context.Context, img message.Segment) (fmt.Stringer, error) {
	var (
		n = fio.NewPath(cmp.Or(os.TempDir(), `data/zbp`), `kitten`, `code.png`)
	)
	bytes, err := n.DownloadImage(ctx, mio.GetImageURL(img))
	if err != nil {
		return nil, fmt.Errorf(`图片下载错误喵！%w`, err)
	}
	log.Info(`正在扫描二维码喵！字节数：`, bytes)
	imgfile, err := os.Open(n.String())
	if err != nil {
		return nil, err
	}
	return scanQRCode(imgfile)
}

// ScanQRCode 扫描 QQ 消息中的二维码图片
func (m *Handler) ScanQRCode(file string) (fmt.Stringer, error) {
	imgfile, err := os.Open(m.GetImage(file).Get(`file`).String())
	if err != nil {
		return nil, err
	}
	return scanQRCode(imgfile)
}

// 扫描二维码
func scanQRCode(imgfile *os.File) (s fmt.Stringer, err error) {
	defer func() {
		if e := imgfile.Close(); e != nil {
			err = fmt.Errorf(`关闭二维码文件失败喵！%w`, e)
		}
	}()
	img, _, err := image.Decode(imgfile)
	if err != nil {
		return nil, err
	}
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, err
	}
	return qrcode.NewQRCodeReader().DecodeWithoutHints(bmp)
}
