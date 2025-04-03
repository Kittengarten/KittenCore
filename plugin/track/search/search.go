// Package search 搜索小说
package search

type (
	// Keyword 搜索关键词
	Keyword string

	// NotFoundError 没有找到小说
	NotFoundError struct {
		key Keyword // 搜索关键词
	}
)

// NotFound *notFoundErr 的构造函数，没有找到小说
func (key Keyword) NotFound() *NotFoundError {
	return &NotFoundError{
		key: key,
	}
}

// Error 实现 error
func (e *NotFoundError) Error() string {
	return `没有找到` + string(e.key) + `关键词的小说喵！`
}
