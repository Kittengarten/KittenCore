package track

type (
	// 状态
	status byte

	// 状态错误
	statusError struct {
		url  string // 链接
		stat status // 状态
	}

	// 不支持的平台
	notSupportedError struct {
		p Platform // 平台
	}

	// 没有找到小说
	notFoundError struct {
		key keyword // 搜索关键词
	}
)

const (
	bookUnreachable        status = iota // 无法访问小说
	noChapterURL                         // 没有章节链接
	onlyAChapter                         // 只有一个章节
	bookStatusException                  // 小说状态异常
	timeException                        // 更新时间异常
	chapterUnreachable                   // 无法访问章节
	chapterURLException                  // 章节链接异常
	chapterStatusException               // 章节状态异常
	vipChapterException                  // 付费状态异常
)

// Error 实现 error
func (e *statusError) Error() string {
	if statusErrs := map[status]string{
		bookUnreachable:        `链接 ` + e.url + ` 没有小说喵！`,
		noChapterURL:           `小说 ` + e.url + ` 没有章节链接喵！`,
		onlyAChapter:           `小说 ` + e.url + ` 只有一个章节喵！`,
		bookStatusException:    `小说 ` + e.url + ` 状态异常喵！`,
		timeException:          `小说 ` + e.url + ` 上次更新时间异常喵！`,
		chapterUnreachable:     `链接 ` + e.url + ` 没有章节喵！`,
		chapterURLException:    e.url + ` 不是正常的章节链接喵！`,
		chapterStatusException: `章节 ` + e.url + ` 状态异常喵！`,
		vipChapterException:    `章节 ` + e.url + ` 付费状态异常喵！`,
	}; statusErrs[e.stat] != `` {
		return statusErrs[e.stat]
	}
	return `状态错误`
}

// *statusErr 的构造函数，状态错误
func errStatus(url string, stat status) *statusError {
	return &statusError{
		url:  url,
		stat: stat,
	}
}

// *notSupportedErr 的构造函数，不支持的平台
func notSupported(p Platform) *notSupportedError {
	return &notSupportedError{
		p: p,
	}
}

// Error 实现 error
func (e *notSupportedError) Error() string {
	return string(e.p) + `不是受支持的小说平台喵！`
}

// *notFoundErr 的构造函数，没有找到小说
func notFound(key keyword) *notFoundError {
	return &notFoundError{
		key: key,
	}
}

// Error 实现 error
func (e *notFoundError) Error() string {
	return `没有找到` + string(e.key) + `关键词的小说喵！`
}
