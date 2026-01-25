package chapter

import (
	"context"
	"sync"

	"github.com/Kittengarten/KittenCore/plugin/track/platform"
)

// Pool 章节池
var Pool = sync.Pool{
	New: func() any {
		return new(Chapter)
	},
}

// String 实现 fmt.Stringer
func (cp *Chapter) String() string {
	return cp.Title + "\n" + cp.URL
}

// New 初始化章节
func New(ctx context.Context, p platform.Platform, cpURL string) (*Chapter, error) {
	cpa, err := p.NewChapter(ctx, cpURL)
	if err != nil {
		return nil, err
	}
	return Assert(cpa), nil
}

// Assert 断言为章节，不是章节时返回空章节
func Assert(a any) *Chapter {
	if cp, ok := a.(*Chapter); ok {
		return cp
	}
	return new(Chapter)
}
