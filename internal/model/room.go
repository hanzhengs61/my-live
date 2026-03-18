package model

// Room 房间表
type Room struct {
	ID           int64  `gorm:"type:bigint;not null;auto_increment;comment:'主键ID'" json:"id"`
	HostID       int64  `gorm:"type:bigint;not null;comment:'主播ID'" json:"host_id"`
	Title        string `gorm:"type:varchar(100);not null;comment:'房间标题'" json:"title"`
	Announcement string `gorm:"type:varchar(500);comment:'房间公告'" json:"announcement"`
	IsPrivate    bool   `gorm:"default:0;comment:'是否私密房间'" json:"is_private"`
	Password     string `gorm:"type:varchar(100);comment:'房间密码（私密时使用）'" json:"password,omitempty"`
	BaseModel
}
