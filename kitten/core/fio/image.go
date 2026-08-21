package fio

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
)

// DownloadImage 从 url 下载图片到 path
func (p Path) DownloadImage(ctx context.Context, url string) (int64, error) {
	// 获取 HTTP 响应体，失败则返回
	b, err := shttp.GETWithContext(ctx, url)
	if err != nil {
		return 0, err
	}
	defer b.Close() //nolint:errcheck
	f, err := p.Open(true)
	if err != nil {
		return 0, err
	}
	i, err := f.ReadFrom(b)
	return i, errors.Join(err, f.Close())
}

// ImageToDataURL 将图片路径转为 Data URL（base64 编码），可能为空
func (p Path) ImageToDataURL() (u string) {
	f, err := p.Open(false)
	if err != nil {
		logError(`打开图片失败了喵！`, p, err)
		return ``
	}
	defer func() {
		if err := f.Close(); err != nil {
			u = err.Error()
		}
	}()
	info, err := f.Stat()
	if err != nil {
		logError(`获取图片信息失败了喵！`, p, err)
		return ``
	}
	if info.Size() == 0 {
		logError(`图片为空喵！`, p, nil)
		return ``
	}
	mime := func() string {
		switch ext := strings.ToLower(filepath.Ext(string(p))); ext {
		case `.png`, `.jpeg`, `.webp`, `.gif`, `.bmp`, `.tiff`, `.avif`, `.heic`, `.heif`:
			return `image/` + ext[1:]
		case `.jpg`, `.jfif`, `.pjpeg`, `.pjp`:
			return `image/jpeg`
		case `.tif`:
			return `image/tiff`
		case `.ico`:
			return `image/x-icon`
		case `.svg`:
			return `image/svg+xml`
		default:
			data, err := io.ReadAll(io.LimitReader(f, 1<<9))
			if err != nil {
				logError(`读取图片失败了喵！`, p, err)
				return ``
			}
			m := http.DetectContentType(data)
			if strings.HasPrefix(m, `image/`) {
				return m
			}
			logError(`文件类型未知喵！`+m, p, nil)
			return ``
		}
	}()
	var (
		out     = new(bytes.Buffer)
		encoder = base64.NewEncoder(base64.StdEncoding, out)
	)
	out.Grow(4 * int(info.Size()))
	defer func() {
		if err := encoder.Close(); err != nil {
			u = err.Error()
		}
	}()
	switch mime {
	case ``:
		return ``
	case `image/jpeg`:
		if info.Size() < 1<<19 {
			img, err := jpeg.Decode(f)
			if err != nil {
				logError(`jpeg 解码失败了喵！`, p, err)
				return ``
			}
			if err := png.Encode(encoder, img); err != nil {
				logError(`png 编码失败了喵！`, p, err)
				return ``
			}
			return "data:image/png;base64," + out.String()
		}
		fallthrough
	default:
		_, err := io.Copy(encoder, f)
		if err != nil {
			logError(`base64 编码失败了喵！`, p, err)
			return ``
		}
		return "data:" + mime + ";base64," + out.String()
	}
}
