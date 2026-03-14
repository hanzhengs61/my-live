package model

// ChatMessage 聊天记录表
// 用于存储直播间所有聊天消息、礼物消息等
type ChatMessage struct {
	ID       int64  `gorm:"type:bigint;primaryKey;autoIncrement;comment:主键ID"`
	RoomID   int64  `gorm:"type:bigint;index;not null;comment:房间ID"`      // 房间ID
	UserID   int64  `gorm:"type:bigint;index;not null;comment:用户ID"`      // 用户ID
	Nickname string `gorm:"type:varchar(100);not null;comment:用户昵称"`      // 冗余昵称，避免查询用户表
	Type     string `gorm:"type:varchar(20);index;not null;comment:消息类型"` // 消息类型 chat/gift/system
	Content  string `gorm:"type:text;comment:聊天内容"`                       // 聊天内容
	GiftID   string `gorm:"type:varchar(50);comment:礼物ID"`                // 礼物ID
	GiftName string `gorm:"type:varchar(100);comment:礼物名称"`               // 礼物名称
	BaseModel
}
