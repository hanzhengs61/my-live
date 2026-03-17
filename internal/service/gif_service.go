package service

import (
	"context"
	"errors"
	"my-live/internal/db"
	"my-live/internal/model"
	"my-live/internal/request"

	"gorm.io/gorm"
)

var (
	ErrInsufficientBalance = errors.New("余额不足")
	ErrGiftNotFound        = errors.New("礼物不存在")
)

type GiftService struct {
}

func NewGiftService() *GiftService {
	return &GiftService{}
}

// SendGift 发送礼物（事务扣费 + 记录交易）
func (s *GiftService) SendGift(ctx context.Context, userID int64, req request.SendGiftReq) (gifName string, err error) {
	// 查询礼物
	var gift model.Gift
	if giftErr := db.DB.WithContext(ctx).First(&gift, req.GiftID).Error; giftErr != nil {
		return "", ErrGiftNotFound
	}

	totalPrice := gift.Price * req.Count

	// 查询房间（获取主播ID）
	var room model.Room
	if roomErr := db.DB.WithContext(ctx).First(&room, req.RoomID).Error; roomErr != nil {
		return "", ErrRoomNotFound
	}
	// 事务处理（扣费 + 记录）
	TransactionErr := db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 扣用户余额
		result := tx.Model(&model.User{}).
			Where("id = ? AND balance >= ?", userID, totalPrice).
			Update("balance", gorm.Expr("balance - ?", totalPrice))

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return ErrInsufficientBalance
		}

		// 记录交易
		txRecord := model.Transaction{
			FromUserID: userID,
			ToUserID:   int64(room.HostID),
			RoomID:     req.RoomID,
			GiftID:     req.GiftID,
			Count:      req.Count,
			TotalPrice: totalPrice,
		}
		if err := tx.Create(&txRecord).Error; err != nil {
			return err
		}
		return nil
	})
	if TransactionErr != nil {
		db.DB.WithContext(ctx).Rollback()
		return "", TransactionErr
	}
	return gift.Name, nil
}
