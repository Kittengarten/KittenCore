// Package check 检测时间记录
package check

import (
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/utils"
)

const (
	// 频率统计窗口
	rateWindow = time.Hour
	// 过期记录清理周期
	cleanupInterval = rateWindow >> 3
	// 缩容触发比例
	//
	//	当有效记录数量 <= 当前容量的 1/4 时，才进行缩容。
	//	使用较大的滞后区间可以避免容量在高低负载之间反复变化。
	shrinkRatio = 4
	// 缩容后的容量倍率
	//
	//	缩容后保留 2 倍有效数据容量，使后续增长有足够余量。
	shrinkCapacityRatio = 2
)

type recorder struct {
	sync.Mutex
	// times 按记录时间递增排列
	//
	//	[0, head) 是已经过期、待回收的数据。
	//	[head, len(times)) 是当前仍然保留的数据。
	times []time.Time
	// 当前有效数据从 head 开始
	head int
}

var rec recorder

func init() {
	utils.Go(`定期清理过期检测报更记录`, func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for now := range ticker.C {
			rec.Lock()
			rec.removeExpired(now)
			rec.Unlock()
		}
	})
}

// RecordTime 记录检测时间
func RecordTime() {
	now := time.Now()
	rec.Lock()
	rec.times = append(rec.times, now)
	rec.Unlock()
}

// QueryRate 查询检测频率（次 / 小时）。
//
//	如果最近一小时内存在记录，则返回最近一小时的检测次数。
//	如果第一条记录距现在不足一小时，则按照实际时间跨度折算为每小时频率。
func QueryRate() int64 {
	now := time.Now()
	rec.Lock()
	// 清理过期记录
	rec.removeExpired(now)
	if rec.head >= len(rec.times) {
		rec.Unlock()
		return 0
	}
	var (
		count   = len(rec.times) - rec.head
		elapsed = now.Sub(rec.times[rec.head])
	)
	rec.Unlock()
	if elapsed <= 0 {
		return int64(count)
	}
	// 记录不足一个小时，按照实际时间跨度折算
	if elapsed < rateWindow {
		return int64(time.Duration(count) * rateWindow / elapsed)
	}
	return int64(count)
}

// 删除已经超过统计窗口的记录
//
//	调用者必须持有 rec.Mutex
func (r *recorder) removeExpired(now time.Time) {
	var (
		times = r.times
		head  = r.head
	)
	if head >= len(times) {
		r.times = nil
		r.head = 0
		return
	}
	cutoff := now.Add(-rateWindow)
	// 找到第一个仍然有效的记录
	for head < len(times) && times[head].Before(cutoff) {
		head++
	}
	if head == r.head {
		// 所有记录都未过期
		return
	}
	if head == len(times) {
		// 所有记录都已经过期
		r.times = nil
		r.head = 0
		return
	}
	// 部分记录过期
	r.head = head
	if n := len(times) - head; n*shrinkRatio <= cap(times) {
		// 缩容
		r.shrink(n)
	}
}

// 将底层数组缩容到合理大小
//
//	调用者必须持有 rec.Mutex
func (r *recorder) shrink(n int) {
	if n == 0 {
		r.times = nil
		r.head = 0
		return
	}
	// 缩容后保留 2 倍有效数据容量。
	//	这样即使负载马上回升，也有较大的增长余量，避免刚缩容就再次扩容
	newTimes := make([]time.Time, n, max(n*shrinkCapacityRatio, n))
	copy(newTimes, r.times[r.head:])
	r.times = newTimes
	r.head = 0
}
