package model

// Gift 礼物表
type Gift struct {
	ID         int64  `gorm:"type:bigint;not null;auto_increment;comment:主键ID" json:"id"`
	Name       string `gorm:"type:varchar(50);not null;comment:礼物名称" json:"name"`
	Price      int64  `gorm:"type:bigint;not null;comment:单价（虚拟币）" json:"price"`
	EffectType string `gorm:"type:varchar(50);comment:特效类型" json:"effect_type"`
	ImageURL   string `gorm:"type:varchar(255);comment:礼物图片URL" json:"image_url"`
	BaseModel
}
