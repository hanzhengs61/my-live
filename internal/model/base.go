package model

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseModel struct {
	CreatedAt int64  `gorm:"type:bigint;comment:创建时间" json:"created_at,omitempty"`
	CreatedBy string `gorm:"type:varchar(50);comment:创建人" json:"created_by,omitempty"`
	UpdatedAt int64  `gorm:"type:bigint;comment:修改时间" json:"updated_at,omitempty"`
	UpdatedBy string `gorm:"type:varchar(50);comment:修改人" json:"updated_by,omitempty"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now().Unix()

	if gcVal := tx.Statement.Context.Value("gin_context"); gcVal != nil {
		if gc, ok := gcVal.(*gin.Context); ok && gc != nil {
			if u, exists := gc.Get("username"); exists {
				if username, ok := u.(string); ok && username != "" {
					b.CreatedBy = username
					b.UpdatedBy = username
				}
			}
		}
	}

	b.CreatedAt = now
	b.UpdatedAt = now
	if b.CreatedBy == "" {
		b.CreatedBy = "system"
		b.UpdatedBy = "system"
	}
	return nil
}

func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now().Unix()
	b.UpdatedAt = now

	if gcVal := tx.Statement.Context.Value("gin_context"); gcVal != nil {
		if gc, ok := gcVal.(*gin.Context); ok && gc != nil {
			if u, exists := gc.Get("username"); exists {
				if username, ok := u.(string); ok && username != "" {
					b.UpdatedBy = username
				}
			}
		}
	}

	if b.UpdatedBy == "" {
		b.UpdatedBy = "system"
	}
	return nil
}
