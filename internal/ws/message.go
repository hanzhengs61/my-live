package ws

import "encoding/json"

// MessageType 定义消息类型（方便后续扩展礼物、连麦等）
type MessageType string

const (
	TypeJoin   MessageType = "join"   // 加入房间
	TypeChat   MessageType = "chat"   // 普通弹幕/聊天
	TypeGift   MessageType = "gift"   // 礼物特效
	TypeLeave  MessageType = "leave"  // 离开房间
	TypePing   MessageType = "ping"   // 心跳ping（服务器发）
	TypePong   MessageType = "pong"   // 心跳pong（客户端回复）
	TypeOnline MessageType = "online" // 在线人数
)

// Message 是所有WebSocket消息的统一结构（JSON结构）
type Message struct {
	Type        MessageType `json:"type"`                   // 消息类型
	RoomID      string      `json:"room_id"`                // 房间ID（直播间唯一标识）
	UserID      int64       `json:"user_id"`                // 用户 ID
	Nickname    string      `json:"nickname"`               // 昵称
	Content     string      `json:"content"`                // 消息内容（弹幕文字、礼物ID等）
	OnlineCount int         `json:"online_count,omitempty"` // 在线人数
	GiftID      string      `json:"gift_id,omitempty"`      // 礼物ID（如 "rose"）
	GiftName    string      `json:"gift_name,omitempty"`    // 礼物名称（如 "玫瑰"）
	Timestamp   int64       `json:"timestamp"`              // 时间戳（毫秒）
}

// ToJSON 把 Message 转换成 JSON 字符串（方便发送）
func (m *Message) ToJSON() []byte {
	data, _ := json.Marshal(m)
	return data
}

// ParseMessage 从 JSON 字节解析成 Message
func ParseMessage(data []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return msg, err
}
