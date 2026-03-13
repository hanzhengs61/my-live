package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init() {
	Client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// 测试连接
	ctx := context.Background()
	if err := Client.Ping(ctx).Err(); err != nil {
		panic("Redis 连接失败: " + err.Error())
	}
	println("Redis 连接成功")
}

// GetOnlineUsersKey 返回房间在线用户 Set 的 key
func GetOnlineUsersKey(roomID string) string {
	return "live:room:" + roomID + ":online_users"
}

// JoinRoom 用户加入房间
func JoinRoom(roomID string, userID uint) error {
	ctx := context.Background()
	key := GetOnlineUsersKey(roomID)
	return Client.SAdd(ctx, key, userID).Err()
}

// LeaveRoom 用户离开房间
func LeaveRoom(roomID string, userID uint) error {
	ctx := context.Background()
	key := GetOnlineUsersKey(roomID)
	return Client.SRem(ctx, key, userID).Err()
}

// GetOnlineCount 获取房间在线人数
func GetOnlineCount(roomID string) (int64, error) {
	ctx := context.Background()
	key := GetOnlineUsersKey(roomID)
	return Client.SCard(ctx, key).Result()
}
