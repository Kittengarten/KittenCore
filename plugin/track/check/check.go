// Package check 检测时间记录
package check

import (
	"slices"
	"sync"
	"time"
)

// 检测时间记录器
var recorder = struct {
	sync.Mutex             // Mutex 互斥锁
	times      []time.Time // Times 检测时间
}{
	Mutex: sync.Mutex{},               // Mutex 互斥锁
	times: make([]time.Time, 0, 1<<6), // Times 检测时间
}

// RecordTime 记录检测时间
func RecordTime() {
	recorder.Lock()
	defer recorder.Unlock()
	recorder.times = append(slices.DeleteFunc(recorder.times, func(t time.Time) bool {
		return t.Before(time.Now().Add(-time.Hour))
	}), time.Now())
}

// QueryRate 查询检测频率（次 / 小时）
func QueryRate() int64 {
	recorder.Lock()
	defer recorder.Unlock()
	l := len(recorder.times)
	if l == 0 {
		return 0
	}
	return int64(time.Duration(l) * time.Hour / time.Since(recorder.times[0]))
}
