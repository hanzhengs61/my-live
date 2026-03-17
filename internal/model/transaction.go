package model

// Transaction 礼物打赏交易记录表
type Transaction struct {
	ID         int64 `gorm:"type:bigint;not null;auto_increment;comment:主键ID" json:"id"`
	FromUserID int64 `gorm:"type:bigint;index;comment:送礼用户ID" json:"from_user_id"`
	ToUserID   int64 `gorm:"type:bigint;index;comment:收到礼物的主播ID" json:"to_user_id"`
	RoomID     int64 `gorm:"type:bigint;index;comment:房间ID" json:"room_id"`
	GiftID     int64 `gorm:"type:bigint;comment:礼物ID" json:"gift_id"`
	Count      int64 `gorm:"type:bigint;not null;comment:礼物数量" json:"count"`
	TotalPrice int64 `gorm:"type:bigint;not null;comment:总金额" json:"total_price"`
	BaseModel
}
