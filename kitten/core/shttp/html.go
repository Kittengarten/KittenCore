package shttp

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/RomiChan/syncx"
	"golang.org/x/net/html"
)

// LoadURLWithContext 从指定的 URL 加载 HTML 文档
func LoadURLWithContext(ctx context.Context, url string) (*html.Node, error) {
	res, err := GETWithContext(ctx, url)
	if err != nil {
		return nil, err
	}
	defer res.Close() //nolint:errcheck
	return html.Parse(res)
}

type (
	// 最近状态
	lastState struct {
		hash [1 << 5]byte
		cl   int
	}
	// 带长度的 io.ReadCloser
	lenReadCloser interface {
		io.ReadCloser
		Len() int
	}
	// 响应体
	lenBody struct {
		io.ReadCloser
		cl int
	}
)

// NewLenBody 返回带长度的响应体
func NewLenBody(r io.ReadCloser, cl int) lenReadCloser { return lenBody{r, cl} }

// Len 返回响应体的长度
func (b lenBody) Len() int { return b.cl }

// 最近状态缓存
var last syncx.Map[string, lastState]

// ErrNoUpdate 无更新
var ErrNoUpdate = errors.New(`无更新`)

// UpdateURLWithContext 从指定的 URL 更新 HTML 文档，如无更新会返回 ErrNoUpdate
func UpdateURLWithContext(ctx context.Context, url string) (*html.Node, error) {
	// 初次请求
	res, err := GETWithContext(ctx, url)
	if err != nil {
		return nil, err
	}
	defer res.Close() //nolint:errcheck
	v, ok := last.Load(url)
	if !ok || v.cl != res.Len() {
		// 缓存中没有，或长度不同，无需比较 hash 以及再次请求
		last.Store(url, lastState{cl: res.Len()})
		return html.Parse(res)
	}
	// 长度相同，则比较 hash
	cur, err := hashURLWithContext(res)
	if err != nil {
		return nil, err
	}
	if v.hash == cur.hash {
		// 哈希相同，无更新
		return nil, ErrNoUpdate
	}
	// 哈希不同，保存新的哈希，再次请求
	last.Store(url, cur)
	_ = res.Close()
	res, err = GETWithContext(ctx, url)
	if err != nil {
		return nil, err
	}
	defer res.Close() //nolint:errcheck
	return html.Parse(res)
}

// 获取 URL 的内容哈希值和长度
func hashURLWithContext(res lenReadCloser) (lastState, error) {
	var zero lastState
	h := sha256.New()
	if _, err := io.Copy(h, res); err != nil {
		return zero, err
	}
	return lastState{
		hash: [sha256.Size]byte(h.Sum(nil)),
		cl:   res.Len(),
	}, nil
}
