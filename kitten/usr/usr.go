package usr

import (
	"context"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// TODO: 泛型方法支持后，改造为返回泛型
// SetContext
// Set
// Reset

type (
	// Handler 消息处理器，收取消息并发送
	Handler interface {
		Context
		Sender
	}
	// Context ZeroBot 上下文
	Context interface {
		// Context 上下文
		context.Context
		// SetContext 设置上下文
		SetContext(ctx context.Context) Context
		// CallAction 使用 context 调用 cqhttp API
		CallAction(action string, params zero.H) zero.APIResponse
		// GetStrangerInfo 获取陌生人信息
		GetStrangerInfo(userID int64, noCache bool) gjson.Result
		// GetGroupMemberListNoCache 无缓存获取群员列表
		GetGroupMemberListNoCache(groupID int64) gjson.Result
		// Check 检查上下文的项目是否均有效且不为空
		Check(i ...kitten.Item) bool
		// Event 返回当前 ZeroBot 上下文的事件
		Event() *zero.Event
	}
	// Sender 待发送的消息
	Sender interface {
		// Get 待发送的消息
		Get() message.Message
		// Set 设置待发送的消息
		Set(message.Message) Handler
		// Send 发送消息
		Send(u ...QQ) message.ID
		// SendMulti 发送多条消息
		SendMulti(u ...QQ) (id []message.ID)
		// SendGroupMessage 发送群消息
		SendGroupMessage(groupID int64) int64
		// SendPrivateMessage 发送私聊消息
		SendPrivateMessage(userID int64) int64
		// QuoteID 引用消息 ID
		QuoteID() message.ID
		// Reset 重置消息
		Reset() Sender
	}
)
