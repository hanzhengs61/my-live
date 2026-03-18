package model

type RechargeRecord struct {
	ID     int64 `gorm:"type:bigint;not null;auto_increment;comment:'主键ID'" json:"id"`
	UserID int64 `gorm:"type:bigint;comment:用户ID" json:"user_id"`
	Amount int64 `gorm:"type:bigint;comment:充值金额" json:"amount"`
	BaseModel
}
