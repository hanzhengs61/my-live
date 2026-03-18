package service

import (
	"context"
	"my-live/internal/db"
	"my-live/internal/model"
)

type ChatService struct{}

func NewChatService() *ChatService {
	return &ChatService{}
}

// SaveMessage 保存聊天消息
func (s *ChatService) SaveMessage(ctx context.Context, msg *model.ChatMessage) error {
	return db.DB.WithContext(ctx).Create(msg).Error
}

// GetRecentMessages 获取房间最近 N 条消息
func (s *ChatService) GetRecentMessages(ctx context.Context, roomID int64, limit int) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	err := db.DB.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}
