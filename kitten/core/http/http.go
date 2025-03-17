package http

import (
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

type (
	// Error 是一个表示 HTTP 错误的结构体
	Error struct {
		URL        string // 请求的 URL
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

// RoundTrip 实现 http.RoundTripper，设置默认 User-Agent
func (f uaSetter) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set(UA, f.ua)
	return http.DefaultTransport.RoundTrip(r)
}

// GET 获取 HTTP GET 响应体（无需关闭）
func GET(url string) (io.Reader, error) {
	res, err := tryTLS(
		func(string, string, io.Reader) (*http.Response, error) {
			return TLSHTTP2Client.Get(url)
		},
		func(string, string, io.Reader) (*http.Response, error) {
			return TLSClient.Get(url)
		},
		http.MethodGet, url, ``, nil)
	if err != nil {
		return nil, err
	}
	return charset.NewReader(res.Body, res.Header.Get(CT))
}

// GETData 获取 HTTP GET 数据
func GETData(url string) ([]byte, error) {
	res, err := tryTLS(
		func(string, string, io.Reader) (*http.Response, error) {
			return TLSHTTP2Client.Get(url)
		},
		func(string, string, io.Reader) (*http.Response, error) {
			return TLSClient.Get(url)
		},
		http.MethodGet, url, ``, nil)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(res.Body)
}

// POST 获取 HTTP POST 响应体（无需关闭）
func POST(url, contentType string, body io.Reader) (io.Reader, error) {
	res, err := tryTLS(
		TLSHTTP2Client.Post,
		TLSClient.Post,
		http.MethodPost,
		url, contentType, body,
	)
	if err != nil {
		return nil, err
	}
	return charset.NewReader(res.Body, res.Header.Get(CT))
}

// POSTData 获取 HTTP POST 数据
func POSTData(url, contentType string, body io.Reader) ([]byte, error) {
	res, err := tryTLS(
		TLSHTTP2Client.Post,
		TLSClient.Post,
		http.MethodPost,
		url, contentType, body,
	)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(res.Body)
}

// 执行 HTTP 请求
func doRequest(method, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(CT, contentType)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return res, checkError(res, url)
}

// tryTLS 尝试 TLS
func tryTLS(f2, f func(string, string, io.Reader) (*http.Response, error),
	method, url, contentType string,
	body io.Reader,
) (res *http.Response, err error) {
	if strings.HasPrefix(url, `https`) {
		res, err = f2(url, contentType, body)
		if err != nil {
			res, err = f(url, contentType, body)
		}
		if err == nil {
			return res, checkError(res, url)
		}
	}
	return doRequest(method, url, contentType, body)
}

// 判断 HTTP 错误
func checkError(res *http.Response, url string) error {
	if http.StatusOK <= res.StatusCode && res.StatusCode < http.StatusMultipleChoices {
		// 不能处理 3xx 重定向状态码
		return nil
	}
	defer res.Body.Close()
	return &Error{
		URL:        url,
		Method:     res.Request.Method,
		StatusCode: res.StatusCode,
	}
}

// InnerText 在 *html.Node 中使用 XPath 获取文本
func InnerText(top *html.Node, expr string) string {
	node, err := htmlquery.Query(top, expr)
	if err != nil {
		return err.Error()
	}
	if node == nil {
		return ``
	}
	return htmlquery.InnerText(node)
}

// 遍历 DOM 树
func walk(node *html.Node, f func(*html.Node) bool) {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if f(c) && c.Type == html.ElementNode {
			walk(c, f) // 递归遍历子节点
		}
	}
}

// ExtractText 从 HTML 文档中提取纯文本
func ExtractText(doc *html.Node) string {
	var b strings.Builder
	walk(doc, func(n *html.Node) bool {
		if n.Type == html.TextNode {
			// 移除空格
			b.WriteString(strings.TrimSpace(n.Data))
			return true
		}
		if n.Type == html.ElementNode {
			switch n.Data {
			case `br`:
				// 写入换行符
				b.WriteByte('\n')
			case `p`:
				// 写入段落标记（考虑到小说排版需求，仅使用一次换行）
				b.WriteByte('\n')
			}
		}
		return true // 继续遍历节点
	})
	return b.String()
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
	return strings.ReplaceAll(UserAgent, `129`, strconv.Itoa(100+rand.N(30)))
}
