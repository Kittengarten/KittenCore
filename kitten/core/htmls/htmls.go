// Package htmls 处理 HTML
package htmls

import (
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

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

// ExtractText 从 HTML 文档中提取纯文本
func ExtractText(doc *html.Node) string {
	if doc == nil {
		return ``
	}
	s := new(strings.Builder)
	s.Grow(1 << 14)
	walk(doc, func(n *html.Node) bool {
		switch n.Data {
		case `head`, `script`, `style`:
			return false // 跳过
		}
		if n.Type == html.TextNode {
			// 移除空格
			s.WriteString(strings.TrimSpace(n.Data))
			return true
		}
		if n.Type == html.ElementNode {
			switch n.Data {
			case `br`, `li`:
				// 写入换行符
				s.WriteByte('\n')
			case `p`:
				// 写入段落标记（考虑到小说排版需求，仅使用一次换行）
				s.WriteByte('\n')
			}
		}
		return true // 继续遍历节点
	})
	return s.String()
}

// 遍历 DOM 树
func walk(node *html.Node, f func(*html.Node) bool) {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if f(c) && c.Type == html.ElementNode {
			walk(c, f) // 递归遍历子节点
		}
	}
}
