package chapter

import "sync"

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
