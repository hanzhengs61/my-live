package model

// Room 房间表
type Room struct {
	ID          int64  `gorm:"type:bigint;not null;auto_increment;comment:主键ID" json:"id"`
	Title       string `gorm:"type:varchar(100);not null;comment:房间标题"`
	Description string `gorm:"type:text;comment:房间描述"`
	HostID      uint   `gorm:"index;comment:主播用户ID"`
	IsPrivate   bool   `gorm:"default:false;comment:是否私密房间"`
	Password    string `gorm:"type:varchar(100);comment:私密房间密码（明文或哈希）"`
	Status      string `gorm:"index;type:varchar(20);default:'waiting';comment:状态：waiting/live/ended"`
	BaseModel
}
