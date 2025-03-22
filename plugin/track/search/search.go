package search

type (
	// Keyword 搜索关键词
	Keyword string

	// 没有找到小说
	notFoundError struct {
		key Keyword // 搜索关键词
	}
)

// NotFound *notFoundErr 的构造函数，没有找到小说
func (key Keyword) NotFound() *notFoundError {
	return &notFoundError{
		key: key,
	}
}

// Error 实现 error
func (e *notFoundError) Error() string {
	return `没有找到` + string(e.key) + `关键词的小说喵！`
}
