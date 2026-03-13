package model

type BaseModel struct {
	CreatedAt int64  `gorm:"type:bigint;comment:创建时间" json:"created_at,omitempty"`
	CreatedBy string `gorm:"type:varchar(50);comment:创建人" json:"created_by,omitempty"`
	UpdatedAt int64  `gorm:"type:bigint;comment:修改时间" json:"updated_at,omitempty"`
	UpdatedBy string `gorm:"type:varchar(50);comment:修改人" json:"updated_by,omitempty"`
}

type AuditSetter interface {
	SetCreatedBy(string)
	SetUpdatedBy(string)
}

func (b *BaseModel) SetCreatedBy(v string) { b.CreatedBy = v }
func (b *BaseModel) SetUpdatedBy(v string) { b.UpdatedBy = v }
