package status

type (
	// Status 状态
	Status byte

	// Error 状态错误
	Error struct {
		url    string // 链接
		Status        // 状态
	}
)

const (
	BookUnreachable        Status = iota // 无法访问小说
	NoChapterURL                         // 没有章节链接
	OnlyAChapter                         // 只有一个章节
	BookStatusException                  // 小说状态异常
	TimeException                        // 更新时间异常
	ChapterUnreachable                   // 无法访问章节
	ChapterURLException                  // 章节链接异常
	ChapterStatusException               // 章节状态异常
	VIPChapterException                  // 付费状态异常
)

// Error 实现 error
func (e *Error) Error() string {
	if statusErrs := map[Status]string{
		BookUnreachable:        `链接 ` + e.url + ` 没有小说喵！`,
		NoChapterURL:           `小说 ` + e.url + ` 没有章节链接喵！`,
		OnlyAChapter:           `小说 ` + e.url + ` 只有一个章节喵！`,
		BookStatusException:    `小说 ` + e.url + ` 状态异常喵！`,
		TimeException:          `小说 ` + e.url + ` 上次更新时间异常喵！`,
		ChapterUnreachable:     `链接 ` + e.url + ` 没有章节喵！`,
		ChapterURLException:    e.url + ` 不是正常的章节链接喵！`,
		ChapterStatusException: `章节 ` + e.url + ` 状态异常喵！`,
		VIPChapterException:    `章节 ` + e.url + ` 付费状态异常喵！`,
	}; statusErrs[e.Status] != `` {
		return statusErrs[e.Status]
	}
	return `状态错误`
}

// ErrStatus *statusErr 的构造函数，状态错误
func ErrStatus(url string, stat Status) *Error {
	return &Error{
		url:    url,
		Status: stat,
	}
}
