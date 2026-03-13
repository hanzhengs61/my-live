package model

// User 用户表
type User struct {
	ID           int64  `gorm:"type:bigint;not null;auto_increment;comment '主键ID'" json:"id"`
	Username     string `gorm:"type:varchar(50);unique;not null;comment:用户名" json:"username,omitempty"`
	PasswordHash string `gorm:"type:varchar(255);not null;comment:密码哈希" json:"password_hash,omitempty"`
	Nickname     string `gorm:"type:varchar(50);comment:昵称" json:"nickname,omitempty"`
	Email        string `gorm:"type:varchar(100);unique;comment:邮箱（可选）" json:"email,omitempty"`
	BaseModel
}
