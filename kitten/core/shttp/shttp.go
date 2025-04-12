// Package shttp 处理 HTTP 请求
package shttp

import (
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/utils"

	"golang.org/x/net/html/charset"
)

type (
	// Error 是一个表示 HTTP 错误的结构体
	Error struct {
		URL        string // 请求的 urlStr
		Method     string // 请求的方法
		StatusCode int    // HTTP 状态码
	}
	// 设置用户代理
	uaSetter struct {
		ua string
	}
)

const (
	CT        = `Content-Type` // CT Content-Type
	UA        = `User-Agent`   // UA User-Agent
	UserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64)` +
		` AppleWebKit/537.36 (KHTML, like Gecko)` +
		` Chrome/129.0.0.0 Safari/537.36 Edg/129.0.0.0` // UserAgent 用户代理
	TimeOutSeconds = 10 // TimeOutSeconds 超时时间
)

var (
	TLSClient      http.Client // TLSClient TLS HTTP 客户端
	TLSHTTP2Client http.Client // TLSHTTP2Client TLS HTTP2 客户端
)

func init() {
	// 设置默认 User-Agent
	SetUserAgent(RandomUserAgent())
	// 设置默认超时时间
	SetTimeOut(TimeOutSeconds * time.Second)
}

// Error 实现 error
func (e *Error) Error() string {
	return `链接：` + e.URL + `
方法：` + e.Method + `
HTTP 错误：` + strconv.Itoa(e.StatusCode)
}

// NewError *Error 的构造函数，创建一个 HTTP 错误
func NewError(urlStr, method string, statusCode int) *Error {
	return &Error{
		URL:        urlStr,
		Method:     method,
		StatusCode: statusCode,
	}
}

// Errorf *Error 的自定义构造函数，创建一个自定义 HTTP 错误
func Errorf(urlStr, method string, statusCode int, msg ...any) error {
	return fmt.Errorf(`%w
%s`,
		NewError(urlStr, method, statusCode),
		msg)
}

// RoundTrip 实现 http.RoundTripper，设置默认 User-Agent
func (f uaSetter) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set(UA, f.ua)
	return http.DefaultTransport.RoundTrip(r)
}

// GETURL 从 fmt.Stringer 获取 HTTP GET 响应体（无需关闭）
func GETURL(u fmt.Stringer) (io.Reader, error) {
	return GET(u.String())
}

// GETDataURL 从 fmt.Stringer 获取 HTTP GET 数据
func GETDataURL(u fmt.Stringer) ([]byte, error) {
	return GETData(u.String())
}

// POSTURL 从 fmt.Stringer 获取 HTTP POST 响应体（无需关闭）
func POSTURL(u fmt.Stringer, contentType string, body io.Reader) (io.Reader, error) {
	return POST(u.String(), contentType, body)
}

// POSTDataURL 从 fmt.Stringer 获取 HTTP POST 数据
func POSTDataURL(u fmt.Stringer, contentType string, body io.Reader) ([]byte, error) {
	return POSTData(u.String(), contentType, body)
}

// GET 获取 HTTP GET 响应体（无需关闭）
func GET(urlStr string) (io.Reader, error) {
	res, err := tryTLS(
		func(string, string, io.Reader) (*http.Response, error) {
			return TLSHTTP2Client.Get(urlStr)
		},
		func(string, string, io.Reader) (*http.Response, error) {
			return TLSClient.Get(urlStr)
		},
		http.MethodGet, urlStr, ``, nil)
	if err != nil {
		return nil, err
	}
	contentType := res.Header.Get(CT)
	if strings.HasPrefix(contentType, `text`) {
		return charset.NewReader(res.Body, contentType)
	}
	return res.Body, nil
}

// GETData 获取 HTTP GET 数据
func GETData(urlStr string) ([]byte, error) {
	res, err := tryTLS(
		func(string, string, io.Reader) (*http.Response, error) {
			return TLSHTTP2Client.Get(urlStr)
		},
		func(string, string, io.Reader) (*http.Response, error) {
			return TLSClient.Get(urlStr)
		},
		http.MethodGet, urlStr, ``, nil)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(res.Body)
}

// POST 获取 HTTP POST 响应体（无需关闭）
func POST(urlStr, contentType string, body io.Reader) (io.Reader, error) {
	res, err := tryTLS(
		TLSHTTP2Client.Post,
		TLSClient.Post,
		http.MethodPost,
		urlStr, contentType, body,
	)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(contentType, `text`) {
		return charset.NewReader(res.Body, contentType)
	}
	return res.Body, nil
}

// POSTData 获取 HTTP POST 数据
func POSTData(urlStr, contentType string, body io.Reader) ([]byte, error) {
	res, err := tryTLS(
		TLSHTTP2Client.Post,
		TLSClient.Post,
		http.MethodPost,
		urlStr, contentType, body,
	)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(res.Body)
}

// 执行 HTTP 请求
func doRequest(method, urlStr, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(CT, contentType)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return res, checkError(res, urlStr)
}

// tryTLS 尝试 TLS
func tryTLS(f2, f func(string, string, io.Reader) (*http.Response, error),
	method, urlStr, contentType string,
	body io.Reader,
) (res *http.Response, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return res, err
	}
	//nolint:nestif
	if u.Scheme == `https` {
		if u.Host == `multimedia.nt.qq.com.cn` {
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
		res, err = f2(urlStr, contentType, body)
		if err != nil {
			res, err = f(urlStr, contentType, body)
		}
		if err == nil {
			return res, checkError(res, urlStr)
		}
	}
	return doRequest(method, urlStr, contentType, body)
}

// 判断 HTTP 错误
func checkError(res *http.Response, urlStr string) error {
	if http.StatusOK <= res.StatusCode && res.StatusCode < http.StatusMultipleChoices {
		// 不能处理 3xx 重定向状态码
		return nil
	}
	defer res.Body.Close()
	return &Error{
		URL:        urlStr,
		Method:     res.Request.Method,
		StatusCode: res.StatusCode,
	}
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
	s := uaSetter{ua: ua}
	http.DefaultClient.Transport = s
	TLSClient.Transport = s
	TLSHTTP2Client.Transport = s
}

// RandomUserAgent 是一个随机的 User-Agent
func RandomUserAgent() string {
	//nolint:gosec
	return strings.ReplaceAll(UserAgent, `129`, strconv.Itoa(100+rand.N(30)))
}

const (
	debug = `GODEBUG`
	rsa   = `tlsrsakex`
)

// SetRSA 设置 RSA
func SetRSA(enable bool) error {
	return os.Setenv(debug, fmt.Sprintf(`%s=%d`, rsa, utils.BoolToInt(enable)))
}
