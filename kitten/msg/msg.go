package msg

import (
	"bytes"
	"cmp"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/internal/config"
	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/times/repeat"
	"github.com/Kittengarten/KittenCore/kitten/msg/mio"
	"github.com/Kittengarten/KittenCore/kitten/msg/seg"
	"github.com/Kittengarten/KittenCore/kitten/usr"

	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// Handler 消息处理器，收取消息并发送
type Handler struct {
	message.ID            // 引用消息 ID
	message.Message       // 待发送的消息
	context.Context       // 上下文
	err             error // 错误（不使用内嵌，避免 Handler 实现 error）
	*zero.Ctx             // Zerobot 上下文
}

// New 创建消息处理器
func New(ctx *zero.Ctx) *Handler {
	return &Handler{Ctx: ctx, Context: context.Background()}
}

// NewWithContext 创建消息处理器（带上下文）
func NewWithContext(c context.Context, ctx *zero.Ctx) *Handler {
	return &Handler{Ctx: ctx, Context: c}
}

// Event 返回当前 ZeroBot 上下文的事件
func (m *Handler) Event() *zero.Event {
	return m.Ctx.Event
}

// QuoteID 引用消息 ID
func (m *Handler) QuoteID() message.ID {
	return m.ID
}

// Get 待发送的消息
func (m *Handler) Get() message.Message {
	return m.Message
}

// Set 设置待发送的消息
func (m *Handler) Set(msg message.Message) usr.Handler {
	m.Message = msg
	return m
}

// GetStrangerInfo 获取陌生人信息
// https://github.com/botuniverse/onebot-11/blob/master/api/public.md#get_stranger_info-%E8%8E%B7%E5%8F%96%E9%99%8C%E7%94%9F%E4%BA%BA%E4%BF%A1%E6%81%AF
func (handler *Handler) GetStrangerInfo(userID int64, noCache bool) gjson.Result {
	return handler.CallActionWithContext(`get_stranger_info`, zero.H{
		`user_id`:  userID,
		`no_cache`: noCache,
	}).Data
}

// GetGroupMemberListNoCache 无缓存获取群员列表
// https://github.com/botuniverse/onebot-11/blob/master/api/public.md#get_group_member_list-%E8%8E%B7%E5%8F%96%E7%BE%A4%E6%88%90%E5%91%98%E5%88%97%E8%A1%A8
func (handler *Handler) GetGroupMemberListNoCache(groupID int64) gjson.Result {
	return handler.CallActionWithContext(`get_group_member_list`, zero.H{
		`group_id`: groupID,
		`no_cache`: true,
	}).Data
}

// 设置回复消息 ID
func (m *Handler) id(id message.ID) *Handler {
	m.ID = id
	return m
}

// Quote 引用消息，id 为引用的消息 ID（仅限一个）
//
//	如引用消息为空则引用本消息的触发来源消息
//	已经设置过引用消息 ID 时，会被参数覆盖，不提供参数时无效
func (m *Handler) Quote(id ...message.ID) *Handler {
	if len(id) > 0 {
		// 正常引用
		return m.id(cmp.Or(id...))
	}
	// 已经设置过引用消息 ID
	if m.ID != (message.ID{}) {
		return m
	}
	// 引用本消息的触发来源消息
	if !m.Check(kitten.Event) {
		log.Warn(ErrNoEvent)
		return m
	}
	if m.Event().MessageID == nil {
		return m
	}
	switch id := m.Event().MessageID.(type) {
	case int64:
		return m.id(message.NewMessageIDFromInteger(id))
	case string:
		return m.id(message.NewMessageIDFromString(id))
	default:
		if m.Event().PostType == `message` {
			log.Infof(`引用消息 ID 断言不成功：%+v`, m.Event)
		}
		return m
	}
}

// At 附带 @，u 为 @ 对象，如 @ 对象为空则 @ Handler 的来源
func (m *Handler) At(u ...usr.QQ) *Handler {
	if m.Event().DetailType == Private {
		// 私聊中的 @ 无效
		return m
	}
	if len(u) == 0 {
		// 如果@ 对象为空，则 @ Handler 的来源
		return m.Seg(message.At(m.Event().UserID))
	}
	for _, i := range u {
		// @ 对象
		if i.IsQQ() {
			m.Seg(i.At())
		}
	}
	return m
}

// AtAll 附带 @ 全体成员
func (m *Handler) AtAll(g ...usr.QQ) *Handler {
	if seg := atAll(m, g...); seg.Type != `` {
		return m.Seg(seg)
	}
	return m
}

// Text 附带文本
func (m *Handler) Text(text ...any) *Handler {
	m.Seg(Text(text...))
	return m
}

// Lf 附带换行
func (m *Handler) Lf(n ...int) *Handler {
	if len(n) == 0 {
		return m.Text("\n")
	}
	return m.Text(strings.Repeat("\n", cmp.Or(n...)))
}

// AtLf 附带 @ 并换行
func (m *Handler) AtLf(qq ...usr.QQ) *Handler {
	if n := *m; !hasSame(m, n.At(qq...)) {
		return n.Lf()
	}
	return m
}

// AtAllLf 附带 @ 全体成员 并换行
func (m *Handler) AtAllLf(g ...usr.QQ) *Handler {
	if n := *m; !hasSame(m, n.AtAll(g...)) {
		return n.Lf()
	}
	return m
}

// 比较待发送的消息是否相等
func hasSame(m ...*Handler) bool {
	return equal.IsSameFunc(equalContained, m...)
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

// Textf 附带格式化文本
func (m *Handler) Textf(format string, a ...any) *Handler {
	m.Seg(Textf(format, a...))
	return m
}

// Imager 图片接口
type Imager interface {
	Image() (string, error)
}

// Image 从图片的相对 | 绝对路径（文件夹），
//
// 或相对 | 绝对路径文件中保存的相对 | 绝对路径，
//
// 或网络路径中附带图片
func (m *Handler) Image(name ...fio.Path) *Handler {
	for _, n := range name {
		if n == `` {
			continue
		}
		img, err := config.ImagePath().Image(n)
		if err != nil {
			m.err = errors.Join(m.err, fmt.Errorf(`附带图片错误：%w`, err))
			img, err = config.ImagePath().Image(fio.NewPath(`error.png`))
			if err != nil {
				m.err = errors.Join(m.err, fmt.Errorf(`附带图片错误：%w`, err))
			}
		}
		m.Seg(img)
	}
	return m
}

// Record 附带语音，支持网络路径
//
//	只支持附带一条语音
func (m *Handler) Record(name ...string) *Handler {
	if len(name) == 0 {
		// 未附带语音
		return m
	}
	for _, s := range m.Message {
		if s.Type == seg.Record {
			// 已经有语音
			return m
		}
	}
	if n := cmp.Or(name...); n != `` {
		return m.Seg(message.Record(n))
	}
	return m
}

// Location 附带位置
func (m *Handler) Location(title, content, lat, lon string) *Handler {
	for _, s := range m.Message {
		if s.Type == seg.Location {
			// 已经有位置
			return m
		}
	}
	return m.Seg(message.Segment{
		Type: seg.Location,
		Data: map[string]string{
			`title`:   title,
			`content`: content,
			`lat`:     lat,
			`lon`:     lon,
		},
	})
}

// Message 附带消息段
func (m *Handler) Seg(seg ...message.Segment) *Handler {
	m.Message = append(m.Message, seg...)
	return m
}

// Send 发送消息，u 为可选的发送对象，如 u 为空则发送给上下文的来源
func (m *Handler) Send(u ...usr.QQ) message.ID {
	if ids := m.SendMulti(u...); len(ids) != 0 {
		// 返回第一个 ID
		return ids[0]
	}
	return message.ID{}
}

// SendMulti 发送多条消息
func (m *Handler) SendMulti(u ...usr.QQ) (id []message.ID) {
	defer m.Reset()
	if m.err != nil {
		// 有错误，将其打包进消息
		m = m.Text("\n", m.err)
	}
	if len(m.Message) == 0 {
		// 没有消息段，无法发送
		return nil
	}
	// 判断 At 后是否添加空格
	if config.AddSpaceAfterAt() {
		var (
			l      = len(m.Message) // 消息段长度
			newMsg = make(message.Message, 0, l+1)
		)
		for i, msg := range m.Message {
			newMsg = append(newMsg, msg)
			if i < l-1 && msg.Type == seg.At &&
				(m.Message[i+1].Type != seg.Text ||
					!strings.HasPrefix(m.Message[i+1].Data[seg.Text], ` `) &&
						!strings.HasPrefix(m.Message[i+1].Data[seg.Text], `　`) &&
						!strings.HasPrefix(m.Message[i+1].Data[seg.Text], "\n")) {
				newMsg = append(newMsg, message.Text(` `))
			}
		}
		m.Message = newMsg
	}
	if len(u) != 0 {
		// 发送对象不为空，向发送对象发送
		if err := repeat.IterS(
			m,
			repeat.New(0, time.Second, 2*time.Second),
			u,
			func(_ int, o usr.QQ) error {
				switch {
				case o.IsGroup(), o.IsQQ():
					id = append(id, o.Send(m))
				}
				return nil
			},
		); err != nil {
			log.Error(err)
		}
		return id
	}
	// 发送对象为空，向 Handler 的来源发送
	if !m.Check(kitten.Caller, kitten.Event) {
		// 没有 APICaller 或 Event ，无法发送
		log.Warn(m)
		return nil
	}
	if m.Event().PostType != `message` || m.QuoteID().ID() == 0 {
		// 不是消息引发的发送或没有回复，不予回复
		return []message.ID{m.SendWithContext(m.Message)}
	}
	for _, e := range m.Message {
		switch e.Type {
		case seg.Text, seg.Face, seg.Image, seg.At:
			// 消息段兼容回复，不执行操作
		default:
			// 消息段不兼容回复，或未经验证，跳过回复程序
			return []message.ID{m.SendWithContext(m.Message)}
		}
	}
	// 有回复
	return []message.ID{m.SendWithContext(message.ReplyWithMessage(m.ID, m.Message...))}
}

// SendWithContext 发送消息（带上下文）
//
//	ctx.Send 的封装
func (handler *Handler) SendWithContext(msg any) message.ID {
	event := handler.Event()
	m, ok := msg.(message.Message)
	if !ok {
		var p *message.Message
		p, ok = msg.(*message.Message)
		if ok {
			m = *p
		}
	}
	if ok && len(m) > 0 && m[0].Type == `node` && event.DetailType != `guild` {
		if event.GroupID != 0 {
			return message.NewMessageIDFromInteger(handler.sendGroupForwardMessageWithContext(event.GroupID, m).Get(`message_id`).Int())
		}
		return message.NewMessageIDFromInteger(handler.sendPrivateForwardMessageWithContext(event.UserID, m).Get(`message_id`).Int())
	}
	if event.DetailType == `guild` {
		return message.NewMessageIDFromString(handler.sendGuildChannelMessageWithContext(event.GuildID, event.ChannelID, msg))
	}
	if event.GroupID != 0 {
		return message.NewMessageIDFromInteger(handler.SendGroupMessageWithContext(event.GroupID, msg))
	}
	return message.NewMessageIDFromInteger(handler.SendPrivateMessageWithContext(event.UserID, msg))
}

// 发送合并转发（群，带上下文）
// https://github.com/Mrs4s/go-cqhttp/blob/master/docs/cqhttp.md#%E5%8F%91%E9%80%81%E5%90%88%E5%B9%B6%E8%BD%AC%E5%8F%91%E7%BE%A4
func (handler *Handler) sendGroupForwardMessageWithContext(groupID int64, message message.Message) gjson.Result {
	return handler.CallActionWithContext(`send_group_forward_msg`, zero.H{
		`group_id`: groupID,
		`messages`: message,
	}).Data
}

// 发送合并转发（私聊，带上下文）
// https://github.com/Mrs4s/go-cqhttp/blob/master/docs/cqhttp.md#%E5%8F%91%E9%80%81%E5%90%88%E5%B9%B6%E8%BD%AC%E5%8F%91%E7%BE%A4
func (handler *Handler) sendPrivateForwardMessageWithContext(userID int64, message message.Message) gjson.Result {
	return handler.CallActionWithContext(`send_private_forward_msg`, zero.H{
		`user_id`:  userID,
		`messages`: message,
	}).Data
}

// 发送频道消息（带上下文）
func (handler *Handler) sendGuildChannelMessageWithContext(guildID, channelID string, message any) string {
	rsp := handler.CallActionWithContext(`send_guild_channel_msg`, zero.H{
		`guild_id`:   guildID,
		`channel_id`: channelID,
		`message`:    message,
	}).Data.Get(`message_id`)
	if rsp.Exists() {
		log.Skip(2).Infof(`[api] 发送频道消息(%v-%v): %v (id=%v)`, guildID, channelID, formatMessage(message), rsp.Int())
		return rsp.String()
	}
	return `0` // 无法获取返回值
}

var base64Reg = regexp.MustCompile(`"type":"image","data":\{"file":"base64://[\w/\+=]+`)

// formatMessage 格式化消息数组
//
//	仅用在 log 打印
func formatMessage(msg any) string {
	switch m := msg.(type) {
	case string:
		return m
	case message.CQCoder:
		return m.CQCode()
	case fmt.Stringer:
		return m.String()
	default:
		s := new(strings.Builder)
		if err := json.NewEncoder(s).Encode(m); err != nil {
			return err.Error()
		}
		return base64Reg.ReplaceAllStringFunc(s.String(), func(s string) string {
			var (
				buf    = bytes.NewBufferString(`"type":"image","data":{"file":"`)
				b, err = base64.StdEncoding.DecodeString(s[40:])
			)
			if err != nil {
				buf.WriteString(err.Error())
				return buf.String()
			}
			m := md5.Sum(b)
			_, err = hex.NewEncoder(buf).Write(m[:])
			if err != nil {
				buf.WriteString(err.Error())
				return buf.String()
			}
			buf.WriteString(`.image`)
			return buf.String()
		})
	}
}

// SendGroupMessageWithContext 发送群消息（带上下文）
// https://github.com/botuniverse/onebot-11/blob/master/api/public.md#send_group_msg-%E5%8F%91%E9%80%81%E7%BE%A4%E6%B6%88%E6%81%AF
func (handler *Handler) SendGroupMessageWithContext(groupID int64, message any) int64 {
	rsp := handler.CallActionWithContext(`send_group_msg`, zero.H{ // 调用并保存返回值
		`group_id`: groupID,
		`message`:  message,
	}).Data.Get(`message_id`)
	if rsp.Exists() {
		log.Skip(2).Infof(`[api] 发送群消息(%v): %v (id=%v)`, groupID, formatMessage(message), rsp.Int())
		return rsp.Int()
	}
	return 0 // 无法获取返回值
}

// SendPrivateMessageWithContext 发送私聊消息（带上下文）
// https://github.com/botuniverse/onebot-11/blob/master/api/public.md#send_private_msg-%E5%8F%91%E9%80%81%E7%A7%81%E8%81%8A%E6%B6%88%E6%81%AF
func (handler *Handler) SendPrivateMessageWithContext(userID int64, message any) int64 {
	rsp := handler.CallActionWithContext(`send_private_msg`, zero.H{
		`user_id`: userID,
		`message`: message,
	}).Data.Get(`message_id`)
	if rsp.Exists() {
		log.Skip(2).Infof(`[api] 发送私聊消息(%v): %v (id=%v)`, userID, formatMessage(message), rsp.Int())
		return rsp.Int()
	}
	return 0 // 无法获取返回值
}

// Reset 重置 Handler，保留 Zerobot 上下文，不保留上下文
func (m *Handler) Reset() usr.Handler {
	m.Context = context.Background()
	m.Message = nil
	m.ID = message.ID{}
	m.err = nil
	return m
}

// CallAction 调用 cqhttp API
func (m *Handler) CallAction(action string, params zero.H) zero.APIResponse {
	return m.Ctx.CallAction(action, params)
}

// CallActionWithContext 调用 cqhttp API（带上下文）
func (m *Handler) CallActionWithContext(action string, params zero.H) zero.APIResponse {
	return m.Ctx.CallActionWithContext(m, action, params)
}

// Textf 格式化构建 message.Segment 文本，格式同 fmt.Sprintf
func Textf(format string, a ...any) message.Segment {
	return Text(fmt.Sprintf(format, a...))
}

// Text 构建 message.Segment 文本，格式同 fmt.Sprint
func Text(text ...any) message.Segment {
	checkErr(text)
	return message.Text(text...)
}

// 检查切片的每个元素是否为错误，如果为非空错误则记录日志
func checkErr(v []any) {
	for _, i := range v {
		if err, ok := i.(error); ok && err != nil {
			log.Error(err)
		}
	}
}

// Image 将收到的图片文件名 | 绝对路径 | 网络 URL | Base64 编码转换为图片消息
func Image(file string, summary ...any) message.Segment {
	return message.Image(file, summary...)
}

// GetImage 获取图片
// https://github.com/botuniverse/onebot-11/blob/master/api/public.md#get_image-%E8%8E%B7%E5%8F%96%E5%9B%BE%E7%89%87
func (handler *Handler) GetImage(file string) gjson.Result {
	return handler.CallActionWithContext(`get_image`, zero.H{
		`file`: file,
	}).Data
}
