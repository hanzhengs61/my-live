package service

import (
	"context"
	"my-live/internal/db"
	"my-live/internal/model"

	"gorm.io/gorm"
)

type ChatService struct {
	db *gorm.DB
}

func NewChatService(db *gorm.DB) *ChatService {
	return &ChatService{
		db: db,
	}
}

// SaveMessage 保存聊天消息
func (s *ChatService) SaveMessage(ctx context.Context, msg *model.ChatMessage) error {

	return db.DB.WithContext(ctx).WithContext(ctx).Create(msg).Error
}

// GetRecentMessages 获取房间最近聊天记录
func (s *ChatService) GetRecentMessages(ctx context.Context, roomID int64, limit int) ([]model.ChatMessage, error) {

	var list []model.ChatMessage

	err := db.DB.WithContext(ctx).WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("id desc").
		Limit(limit).
		Find(&list).Error

	return list, err
}
