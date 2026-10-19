package shttp

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/Kittengarten/KittenCore/kitten/core/lru"
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
		hash   [1 << 5]byte
		length int
	}
	// 带长度的 io.ReadCloser
	lenReadCloser interface {
		io.ReadCloser
		Len() int
	}
	// 响应体
	lenBody struct {
		io.ReadCloser
		length int
	}
)

// NewLenBody 返回带长度的响应体
func NewLenBody(r io.ReadCloser, length int) lenReadCloser { return lenBody{r, length} }

// Len 返回响应体的长度
func (b lenBody) Len() int { return b.length }

// 最近状态缓存，保存 100 个 URL 的最近状态
var last = lru.New[string, lastState](100)

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
	v, ok := last.Get(url)
	if !ok || v.length != res.Len() {
		// 缓存中没有，或长度不同，无需比较 hash 以及再次请求
		last.Add(url, lastState{length: res.Len()})
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
	last.Add(url, cur)
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
		hash:   [sha256.Size]byte(h.Sum(nil)),
		length: res.Len(),
	}, nil
}
