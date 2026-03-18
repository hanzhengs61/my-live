package service

import (
	"context"
	"errors"
	"my-live/internal/db"
	"my-live/internal/model"

	"gorm.io/gorm"
)

type RechargeService struct{}

func NewRechargeService() *RechargeService {
	return &RechargeService{}
}

// Recharge 充值（支持任意金额）
func (s *RechargeService) Recharge(ctx context.Context, userID uint, amount int64) error {
	if amount <= 0 {
		return errors.New("充值金额必须大于0")
	}

	return db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 增加余额
		if err := tx.Model(&model.User{}).Where("id = ?", userID).
			Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}

		// 记录充值记录（可后续扩展）
		record := model.RechargeRecord{
			UserID: int64(userID),
			Amount: amount,
		}
		return tx.Create(&record).Error
	})
}
