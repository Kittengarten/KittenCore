// Package status 小说状态
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
	// Unknown 未知错误
	Unknown Status = iota
	// BookUnreachable 无法访问小说
	BookUnreachable
	// NoChapterURL	没有章节链接
	NoChapterURL
	// OnlyAChapter 只有一个章节
	OnlyAChapter
	// BookStatusException 小说状态异常
	BookStatusException
	// TimeException 更新时间异常
	TimeException
	// ChapterUnreachable 无法访问章节
	ChapterUnreachable
	// ChapterURLException 章节链接异常
	ChapterURLException
	// ChapterStatusException 章节状态异常
	ChapterStatusException
	// VIPChapterException 付费状态异常
	VIPChapterException
)

// ErrStatus *statusErr 的构造函数，状态错误
func ErrStatus(url string, stat Status) *Error {
	return &Error{url: url, Status: stat}
}

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
	return `状态错误喵！`
}
