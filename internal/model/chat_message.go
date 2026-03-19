package model

type ChatMessage struct {
	ID        int64  `gorm:"type:bigint;not null;auto_increment;comment:主键ID" json:"id"`
	RoomID    int64  `gorm:"type:bigint;not null;comment:房间ID" json:"room_id"`
	UserID    int64  `gorm:"type:bigint;not null;comment:用户ID" json:"user_id"`
	Nickname  string `gorm:"type:varchar(50);not null;comment:昵称" json:"nickname"`
	Type      string `gorm:"type:varchar(20);not null;default:chat;comment:'消息类型'" json:"type"`
	Content   string `gorm:"type:text;not null;comment:消息内容" json:"content"`
	Timestamp int64  `gorm:"type:bigint;comment:发送时间" json:"timestamp"`
	BaseModel
}
