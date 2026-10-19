package msg

import (
	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/msg/mio"
	"github.com/Kittengarten/KittenCore/kitten/msg/seg"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// 比较待发送的消息是否相等
func hasSame(handlers ...*Handler) bool {
	return equal.IsSameFunc(equalContained, handlers...)
}

// 比较含有的两个消息段切片是否相等
func equalContained(a, b *Handler) bool {
	return isEqual(a.Message, b.Message)
}

// 比较多个消息段是否相等
func IsSameSegment(s ...message.Segment) bool {
	return equal.IsSameFunc(equalSegment, s...)
}

// 比较多个消息段切片是否相等
func IsSame(m ...message.Message) bool {
	return equal.IsSameFunc(isEqual, m...)
}

// 比较两个消息段切片是否相等
func isEqual(a, b message.Message) bool {
	if len(a) != len(b) {
		// 如果两个消息段切片的长度不同，则不相等
		return false
	}
	for i, seg := range a {
		if !equalSegment(seg, b[i]) {
			return false
		}
	}
	return true
}

// 比较两个消息段是否相等
func equalSegment(a, b message.Segment) bool {
	if a.Type != b.Type {
		// 如果两个消息段类型不同，则不相等
		return false
	}
	// 按类型的特殊比较路径
	switch a.Type {
	case seg.Image:
		// 图片，比较文件或路径
		return mio.GetImagePath(a) != `` && mio.GetImagePath(a) == mio.GetImagePath(b) ||
			mio.GetImageURL(a) != `` && mio.GetImageURL(a) == mio.GetImageURL(b)
	case seg.Record, seg.Video, seg.Anonymous, seg.Share, seg.Contact,
		seg.Location, seg.Music, seg.Forward, seg.Node, seg.XML, seg.JSON:
		// 忽略的类型，视为不相等
		return false
	default:
		// 直接比较数据
		return equal.IsSameMap(a.Data, b.Data)
	}
}
