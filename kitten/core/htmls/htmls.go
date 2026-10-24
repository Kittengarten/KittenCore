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
//
//	dp 段落使用 2 次换行
//	img 提取图片
func ExtractText(doc *html.Node, dp, img bool) string {
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
			case `br`, `li`, `h1`, `h2`, `h3`, `h4`, `h5`, `h6`:
				// 写入换行符
				s.WriteByte('\n')
			case `p`:
				// 写入段落标记
				s.WriteByte('\n')
				if dp {
					s.WriteByte('\n')
				}
			case `img`:
				if !img {
					break
				}
				// 处理图片
				var src, alt string
				// 优先使用实际图片地址，其次尝试懒加载属性
				for _, attr := range n.Attr {
					switch attr.Key {
					case `src`:
						src = attr.Val
					case `data-src`, `data-original`, `data-lazy-src`:
						if src == `` {
							src = attr.Val
						}
					case `alt`:
						alt = attr.Val
					}
				}
				s.WriteString(`![`)
				s.WriteString(alt)
				s.WriteString(`](`)
				s.WriteString(src)
				s.WriteByte(')')
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
