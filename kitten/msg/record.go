package msg

import "github.com/wdvxdr1123/ZeroBot/message"

// SendRecord 发送语音
func (handler *Handler) SendRecord(character, text string) message.ID {
	return message.NewMessageIDFromString(
		handler.SendGroupAIRecord(character, handler.Event().GroupID, text))
}
