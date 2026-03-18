package service

import (
	"context"
	"fmt"
	"my-live/internal/db"
	"my-live/internal/model"
	"my-live/internal/redis"
	"strconv"
	"time"
)

type RankService struct{}

func NewRankService() *RankService {
	return &RankService{}
}

// AddGiftScore 送礼时原子增加用户分数（Redis）
func (s *RankService) AddGiftScore(ctx context.Context, roomID int64, userID uint, score int64) error {
	key := fmt.Sprintf("live:room:%d:gift_rank", roomID)
	userIDStr := strconv.Itoa(int(userID))

	_, err := redis.Client.ZIncrBy(ctx, key, float64(score), userIDStr).Result()
	if err != nil {
		return err
	}

	// 设置过期时间（7天），防止冷房间占用内存
	redis.Client.Expire(ctx, key, 7*24*time.Hour)

	return nil
}

// GetTopGiftRank 获取房间礼物排行 Top N（Redis + 补昵称）
func (s *RankService) GetTopGiftRank(ctx context.Context, roomID int64, topN int64) ([]map[string]interface{}, error) {
	key := fmt.Sprintf("live:room:%d:gift_rank", roomID)

	// ZREVRANGE 获取前 N 名（从高到低）
	result, err := redis.Client.ZRevRangeWithScores(ctx, key, 0, topN-1).Result()
	if err != nil {
		return nil, err
	}

	rankList := make([]map[string]interface{}, 0, len(result))
	for rank, z := range result {
		userIDStr := z.Member.(string)
		userID, _ := strconv.ParseInt(userIDStr, 10, 64)

		var nickname string
		var user model.User
		if err := db.DB.WithContext(ctx).Select("nickname").First(&user, userID).Error; err == nil {
			nickname = user.Nickname
		} else {
			nickname = "用户" + userIDStr
		}

		rankList = append(rankList, map[string]interface{}{
			"rank":     rank + 1,
			"user_id":  userID,
			"nickname": nickname,
			"total":    int64(z.Score),
		})
	}

	return rankList, nil
}
