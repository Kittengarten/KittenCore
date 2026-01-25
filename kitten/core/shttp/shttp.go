// Package shttp 处理 HTTP 请求
package shttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/utils"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"

	trsh1 "github.com/fumiama/terasu/http"
	trsh2 "github.com/fumiama/terasu/http2"
)

type (
	// Error 是一个表示 HTTP 错误的结构体
	Error struct {
		URL        string // 请求的 urlStr
		Method     string // 请求的方法
		StatusCode int    // HTTP 状态码
		Message    string // 错误信息
	}

	// 设置用户代理
	uaSetter string
)

const (
	CT        = `Content-Type` // CT Content-Type
	UA        = `User-Agent`   // UA User-Agent
	UserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64)` +
		` AppleWebKit/537.36 (KHTML, like Gecko)` +
		` Chrome/129.0.0.0 Safari/537.36 Edg/129.0.0.0` // UserAgent 用户代理
	TimeoutSeconds = 10                           // TimeOutSeconds 超时时间
	Timeout        = TimeoutSeconds * time.Second // TimeOut 超时时间
)

var (
	TLSClient      = &trsh1.DefaultClient // TLSClient TLS HTTP 客户端
	TLSHTTP2Client = &trsh2.DefaultClient // TLSHTTP2Client TLS HTTP2 客户端
)

func init() {
	// 设置默认 User-Agent
	SetUserAgent(RandomUserAgent())
	// 设置默认超时时间
	SetTimeOut(Timeout)
}

// RandomUserAgent 是一个随机的 User-Agent
func RandomUserAgent() string {
	//nolint:gosec
	return strings.ReplaceAll(UserAgent, `129`, strconv.Itoa(100+rand.N(30)))
}

// SetTimeOut 设置超时时间
func SetTimeOut(d time.Duration) {
	http.DefaultClient.Timeout = d
	TLSClient.Timeout = d
	TLSHTTP2Client.Timeout = d
}

// SetUserAgent 设置用户代理
func SetUserAgent(ua string) {
	if ua == `` {
		return
	}
	s := uaSetter(ua)
	http.DefaultClient.Transport = s
	TLSClient.Transport = s
	TLSHTTP2Client.Transport = s
}

// RoundTrip 实现 http.RoundTripper，设置默认 User-Agent
func (ua uaSetter) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set(UA, string(ua))
	return http.DefaultTransport.RoundTrip(r)
}

// Error 实现 error
func (e *Error) Error() string {
	return `链接：` + e.URL + `
方法：` + e.Method + `
HTTP 错误：` + strconv.Itoa(e.StatusCode) + `
信息：` + e.Message
}

// NewError *Error 的构造函数，创建一个 HTTP 错误
func NewError(urlStr, method string, statusCode int, msg string) *Error {
	return &Error{
		URL:        urlStr,
		Method:     method,
		StatusCode: statusCode,
		Message:    msg,
	}
}

// GETURL 从 fmt.Stringer 获取 HTTP GET 响应体
func GETURL(u fmt.Stringer) (io.ReadCloser, error) {
	return GET(u.String())
}

// GETURLWithContext 从 fmt.Stringer 获取 HTTP GET 响应体（带上下文）
func GETURLWithContext(ctx context.Context, u fmt.Stringer) (io.ReadCloser, error) {
	return GETWithContext(ctx, u.String())
}

// GETDataURL 从 fmt.Stringer 获取 HTTP GET 数据
func GETDataURL(u fmt.Stringer) ([]byte, error) {
	return GETData(u.String())
}

// GETDataURLWithContext 从 fmt.Stringer 获取 HTTP GET 数据（带上下文）
func GETDataURLWithContext(ctx context.Context, u fmt.Stringer) ([]byte, error) {
	return GETDataWithContext(ctx, u.String())
}

// POSTURL 从 fmt.Stringer 获取 HTTP POST 响应体
func POSTURL(u fmt.Stringer, contentType string, body io.Reader) (io.ReadCloser, error) {
	return POST(u.String(), contentType, body)
}

// POSTURLWithContext 从 fmt.Stringer 获取 HTTP POST 响应体（带上下文）
func POSTURLWithContext(ctx context.Context, u fmt.Stringer, contentType string, body io.Reader) (io.ReadCloser, error) {
	return POSTWithContext(ctx, u.String(), contentType, body)
}

// POSTDataURL 从 fmt.Stringer 获取 HTTP POST 数据
func POSTDataURL(u fmt.Stringer, contentType string, body io.Reader) ([]byte, error) {
	return POSTData(u.String(), contentType, body)
}

// POSTDataURLWithContext 从 fmt.Stringer 获取 HTTP POST 数据（带上下文）
func POSTDataURLWithContext(ctx context.Context, u fmt.Stringer, contentType string, body io.Reader) ([]byte, error) {
	return POSTDataWithContext(ctx, u.String(), contentType, body)
}

// GET 获取 HTTP GET 响应体
func GET(urlStr string) (io.ReadCloser, error) {
	return GETWithContext(context.Background(), urlStr)
}

// GETWithContext 获取 HTTP GET 响应体（带上下文）
func GETWithContext(ctx context.Context, urlStr string) (io.ReadCloser, error) {
	res, err := tryTLS(
		ctx,
		func(context.Context, string, string, io.Reader) (*http.Response, error) {
			return getWithContext(ctx, TLSHTTP2Client, urlStr)
		},
		func(context.Context, string, string, io.Reader) (*http.Response, error) {
			return getWithContext(ctx, TLSClient, urlStr)
		},
		http.MethodGet, urlStr, ``, nil)
	if err != nil {
		return nil, err
	}
	return CharsetConv(res.Header.Get(CT), res.Body)
}

// GETData 获取 HTTP GET 数据
func GETData(urlStr string) ([]byte, error) {
	return GETDataWithContext(context.Background(), urlStr)
}

// GETDataWithContext 获取 HTTP GET 数据（带上下文）
func GETDataWithContext(ctx context.Context, urlStr string) ([]byte, error) {
	res, err := tryTLS(
		ctx,
		func(context.Context, string, string, io.Reader) (*http.Response, error) {
			return getWithContext(ctx, TLSHTTP2Client, urlStr)
		},
		func(context.Context, string, string, io.Reader) (*http.Response, error) {
			return getWithContext(ctx, TLSClient, urlStr)
		},
		http.MethodGet, urlStr, ``, nil)
	if err != nil {
		return nil, err
	}
	defer Clear(res.Body)
	return io.ReadAll(res.Body)
}

func getWithContext(
	ctx context.Context,
	c *http.Client,
	url string,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return Retry(ctx, c, req)
}

// POST 获取 HTTP POST 响应体
func POST(urlStr, contentType string, body io.Reader) (io.ReadCloser, error) {
	return POSTWithContext(context.Background(), urlStr, contentType, body)
}

// POSTWithContext 获取 HTTP POST 响应体（带上下文）
func POSTWithContext(ctx context.Context, urlStr, contentType string, body io.Reader) (io.ReadCloser, error) {
	res, err := tryTLS(
		ctx,
		func(context.Context, string, string, io.Reader) (*http.Response, error) {
			return postWithContext(ctx, TLSHTTP2Client, urlStr, contentType, body)
		},
		func(context.Context, string, string, io.Reader) (*http.Response, error) {
			return postWithContext(ctx, TLSClient, urlStr, contentType, body)
		},
		http.MethodPost,
		urlStr, contentType, body,
	)
	if err != nil {
		return nil, err
	}
	return CharsetConv(res.Header.Get(CT), res.Body)
}

// POSTData 获取 HTTP POST 数据
func POSTData(urlStr, contentType string, body io.Reader) ([]byte, error) {
	return POSTDataWithContext(context.Background(), urlStr, contentType, body)
}

// POSTDataWithContext 获取 HTTP POST 数据（带上下文）
func POSTDataWithContext(ctx context.Context, urlStr, contentType string, body io.Reader) ([]byte, error) {
	res, err := tryTLS(
		ctx,
		func(context.Context, string, string, io.Reader) (*http.Response, error) {
			return postWithContext(ctx, TLSHTTP2Client, urlStr, contentType, body)
		},
		func(context.Context, string, string, io.Reader) (*http.Response, error) {
			return postWithContext(ctx, TLSClient, urlStr, contentType, body)
		},
		http.MethodPost,
		urlStr, contentType, body,
	)
	if err != nil {
		return nil, err
	}
	defer Clear(res.Body)
	return io.ReadAll(res.Body)
}

func postWithContext(
	ctx context.Context,
	c *http.Client,
	url, contentType string,
	body io.Reader,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(CT, contentType)
	return Retry(ctx, c, req)
}

// tryTLS 尝试 TLS
func tryTLS(ctx context.Context, f2, f func(context.Context, string, string, io.Reader) (*http.Response, error),
	method, urlStr, contentType string,
	body io.Reader,
) (res *http.Response, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return res, err
	}
	//nolint:nestif
	if u.Scheme == `https` {
		switch u.Host {
		case `multimedia.nt.qq.com.cn`:
			// 临时启用 RSA
			// 避免 remote error: tls: handshake failure
			if err = SetRSA(true); err != nil {
				return res, err
			}
			defer func() {
				if errNew := SetRSA(false); err != nil {
					err = errors.Join(err, errNew)
				}
			}()
		}
		res, err = f2(ctx, urlStr, contentType, body)
		err = checkError(err, res, urlStr)
		logError := func(s string) {
			slog.Error(s,
				slog.Any(`错误`, err),
				slog.String(`Method`, method),
				slog.String(CT, contentType),
				slog.String(`URL`, urlStr),
			)
		}
		if err != nil {
			logError(`TLS HTTP/2 请求失败`)
			res, err = f(ctx, urlStr, contentType, body)
			err = checkError(err, res, urlStr)
		}
		if err == nil {
			return res, nil
		}
		logError(`TLS HTTP 请求失败`)
	}
	return doRequest(ctx, method, urlStr, contentType, body)
}

const (
	debug = `GODEBUG`
	rsa   = `tlsrsakex`
)

// SetRSA 设置 RSA
func SetRSA(enable bool) error {
	return os.Setenv(debug, rsa+`=`+utils.BoolToIntStr(enable))
}

// 执行 HTTP 请求
func doRequest(ctx context.Context, method, urlStr, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(CT, contentType)
	res, err := Retry(ctx, http.DefaultClient, req)
	return res, checkError(err, res, urlStr)
}

// 判断 HTTP 错误
func checkError(err error, res *http.Response, urlStr string) error {
	if err != nil {
		return err
	}
	if res == nil {
		return &Error{URL: urlStr, Message: `响应为空喵！`}
	}
	if http.StatusOK <= res.StatusCode && res.StatusCode < http.StatusMultipleChoices {
		// 不能处理 3xx 重定向状态码
		return nil
	}
	return NewError(urlStr, res.Request.Method, res.StatusCode, ``)
}

// CharsetConv 转换字符集
func CharsetConv(contentType string, body io.ReadCloser) (io.ReadCloser, error) {
	if !strings.HasPrefix(contentType, `text`) &&
		!strings.HasPrefix(contentType, `application/json`) &&
		!strings.HasPrefix(contentType, `application/xml`) {
		return body, nil
	}
	r, err := charset.NewReader(body, contentType)
	if err != nil {
		// 仅当错误时关闭
		// 因为 charset.NewReader 如果调用 transform.NewReader 会复用 body
		_ = Clear(body)
		return nil, err
	}
	return NewConvBody(r, body.Close), nil
}

// 转换后的响应体
type convBody struct {
	io.Reader
	closeFunc func() error
}

// Close 实现 io.Closer
func (b *convBody) Close() error {
	return b.closeFunc()
}

// NewConvBody 新建转换后的响应体
func NewConvBody(r io.Reader, closeFunc func() error) io.ReadCloser {
	return &convBody{
		Reader: r,
		closeFunc: func() error {
			if closeFunc != nil {
				return closeFunc()
			}
			return nil
		},
	}
}

// Clear 清除响应体
func Clear(body io.ReadCloser) error {
	_, err := io.Copy(io.Discard, body)
	if err != nil {
		return errors.Join(err, body.Close())
	}
	return body.Close()
}

// LoadURLWithContext 从指定的 URL 加载 HTML 文档
func LoadURLWithContext(ctx context.Context, url string) (*html.Node, error) {
	res, err := GETWithContext(ctx, url)
	if err != nil {
		return nil, err
	}
	defer Clear(res)
	return html.Parse(res)
}
